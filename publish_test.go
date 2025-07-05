package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestContainsTag(t *testing.T) {
	tests := []struct {
		tags      string
		searchTag string
		expected  bool
	}{
		{"@Projects, @jnl, @public", "public", true},
		{"@Projects, @jnl, @trunk", "public", false},
		{"@secret, @work", "secret", true},
		{"@Projects, @jnl, @trunk", "jnl", true},
		{"", "anything", false},
		{"@single", "single", true},
	}

	for _, test := range tests {
		result := containsTag(test.tags, test.searchTag)
		if result != test.expected {
			t.Errorf("containsTag(%q, %q) = %v, expected %v", test.tags, test.searchTag, result, test.expected)
		}
	}
}

func TestHasBlockedTag(t *testing.T) {
	// Save original blocked tags
	originalBlockedTags := blockedTags
	defer func() { blockedTags = originalBlockedTags }()

	// Set test blocked tags
	blockedTags = []string{"secret", "private", "confidential"}

	tests := []struct {
		tags     string
		expected bool
	}{
		{"@Projects, @jnl, @public", false},
		{"@Projects, @secret, @public", true},
		{"@private, @work", true},
		{"@confidential, @notes", true},
		{"@work, @public", false},
		{"", false},
	}

	for _, test := range tests {
		result := hasBlockedTag(test.tags)
		if result != test.expected {
			t.Errorf("hasBlockedTag(%q) = %v, expected %v", test.tags, result, test.expected)
		}
	}
}

func TestFormatHugoDate(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"2025-07-05T10:56:44-03:00", "2025-07-05T10:56:44-03:00"},
		{"2022-01-30T16:00:03-03:00", "2022-01-30T16:00:03-03:00"},
		{"invalid-date", "invalid-date"}, // Should return original if parsing fails
	}

	for _, test := range tests {
		result := formatHugoDate(test.input)
		if result != test.expected {
			t.Errorf("formatHugoDate(%q) = %q, expected %q", test.input, result, test.expected)
		}
	}
}

func TestBuildHugoFrontmatter(t *testing.T) {
	header := map[string]string{
		"date":   "2025-07-05T10:56:44-03:00",
		"title":  "Test Article",
		"tags":   "@Projects, @jnl, @public, @golang",
		"user":   "testuser",
		"branch": "main",
	}

	body := []byte("# Content\n\nThis is test content.")

	result := buildHugoFrontmatter(header, body)

	// Check that the result contains expected elements
	expectedParts := []string{
		"+++",
		"date = \"2025-07-05T10:56:44-03:00\"",
		"lastmod = \"2025-07-05T10:56:44-03:00\"",
		"title = \"Test Article\"",
		"tags = [\"Projects\", \"jnl\", \"golang\"]", // public should be excluded
		"+++",
	}

	for _, part := range expectedParts {
		if !strings.Contains(result, part) {
			t.Errorf("buildHugoFrontmatter() result missing expected part: %q", part)
		}
	}

	// Check that the public tag is not included in tags
	if strings.Contains(result, "\"public\"") {
		t.Error("buildHugoFrontmatter() should not include the publish tag in Hugo tags")
	}
}

func TestPublishEntry(t *testing.T) {
	// Create temporary directories
	tempDir, err := os.MkdirTemp("", "publish-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	sourceDir := filepath.Join(tempDir, "source")
	targetDir := filepath.Join(tempDir, "target")

	err = os.MkdirAll(sourceDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create source dir: %v", err)
	}

	// Save original variables
	originalPublishTag := publishTag
	originalBlockedTags := blockedTags
	defer func() {
		publishTag = originalPublishTag
		blockedTags = originalBlockedTags
	}()

	// Set test values
	publishTag = "public"
	blockedTags = []string{"secret"}

	// Test case 1: Valid entry with publish tag
	sourceFile1 := filepath.Join(sourceDir, "test1.md")
	content1 := `;;; jnl
date: 2025-07-05T10:56:44-03:00
title: Test Article
tags: @Projects, @public
;;;

# Test Content

This is a test article.`

	err = os.WriteFile(sourceFile1, []byte(content1), 0644)
	if err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	// Test publishing valid entry
	err = publishEntry(sourceFile1, targetDir)
	if err != nil {
		t.Errorf("publishEntry() failed for valid entry: %v", err)
	}

	// Check if target file was created
	targetFile1 := filepath.Join(targetDir, "test1.md")
	if _, err := os.Stat(targetFile1); os.IsNotExist(err) {
		t.Error("publishEntry() did not create target file")
	} else {
		// Read and verify content
		targetContent, err := os.ReadFile(targetFile1)
		if err != nil {
			t.Errorf("Failed to read target file: %v", err)
		} else {
			contentStr := string(targetContent)
			if !strings.Contains(contentStr, "+++") {
				t.Error("Target file should contain Hugo frontmatter")
			}
			if !strings.Contains(contentStr, "# Test Content") {
				t.Error("Target file should contain original body content")
			}
		}
	}

	// Test case 2: Entry without publish tag
	sourceFile2 := filepath.Join(sourceDir, "test2.md")
	content2 := `;;; jnl
date: 2025-07-05T10:56:44-03:00
title: Non-Public Article
tags: @Projects, @private
;;;

This should not be published.`

	err = os.WriteFile(sourceFile2, []byte(content2), 0644)
	if err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	err = publishEntry(sourceFile2, targetDir)
	if err == nil {
		t.Error("publishEntry() should fail for entry without publish tag")
	}

	// Test case 3: Entry with blocked tag
	sourceFile3 := filepath.Join(sourceDir, "test3.md")
	content3 := `;;; jnl
date: 2025-07-05T10:56:44-03:00
title: Secret Article
tags: @Projects, @public, @secret
;;;

This has both public and secret tags.`

	err = os.WriteFile(sourceFile3, []byte(content3), 0644)
	if err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	err = publishEntry(sourceFile3, targetDir)
	if err == nil {
		t.Error("publishEntry() should fail for entry with blocked tags")
	}
}

func TestPublishCommand(t *testing.T) {
	// Create temporary directories
	tempDir, err := os.MkdirTemp("", "publish-cmd-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	sourceDir := filepath.Join(tempDir, "journal")
	targetDir := filepath.Join(tempDir, "hugo")

	err = os.MkdirAll(sourceDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create source dir: %v", err)
	}

	// Save original variables
	originalJournalPath := journalPath
	originalPublishTag := publishTag
	originalBlockedTags := blockedTags
	defer func() {
		journalPath = originalJournalPath
		publishTag = originalPublishTag
		blockedTags = originalBlockedTags
	}()

	// Set test values
	journalPath = sourceDir
	publishTag = "public"
	blockedTags = []string{"secret"}

	// Create test files
	testFiles := []struct {
		filename string
		content  string
	}{
		{
			"public1.md",
			`;;; jnl
date: 2025-07-05T10:56:44-03:00
title: Public Article 1
tags: @Projects, @public
;;;

# Public Content 1`,
		},
		{
			"public2.md",
			`;;; jnl
date: 2025-07-05T10:56:44-03:00
title: Public Article 2
tags: @Work, @public
;;;

# Public Content 2`,
		},
		{
			"private.md",
			`;;; jnl
date: 2025-07-05T10:56:44-03:00
title: Private Article
tags: @Projects, @work
;;;

# Private Content`,
		},
		{
			"secret.md",
			`;;; jnl
date: 2025-07-05T10:56:44-03:00
title: Secret Article
tags: @Projects, @public, @secret
;;;

# Secret Content`,
		},
	}

	for _, file := range testFiles {
		err = os.WriteFile(filepath.Join(sourceDir, file.filename), []byte(file.content), 0644)
		if err != nil {
			t.Fatalf("Failed to write test file %s: %v", file.filename, err)
		}
	}

	// Test publish command
	err = publishCommand(targetDir)
	if err != nil {
		t.Errorf("publishCommand() failed: %v", err)
	}

	// Check that only public files were published
	expectedFiles := []string{"public1.md", "public2.md"}
	for _, filename := range expectedFiles {
		targetFile := filepath.Join(targetDir, filename)
		if _, err := os.Stat(targetFile); os.IsNotExist(err) {
			t.Errorf("Expected file %s was not published", filename)
		}
	}

	// Check that private and secret files were not published
	unexpectedFiles := []string{"private.md", "secret.md"}
	for _, filename := range unexpectedFiles {
		targetFile := filepath.Join(targetDir, filename)
		if _, err := os.Stat(targetFile); err == nil {
			t.Errorf("File %s should not have been published", filename)
		}
	}
}
