package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"unicode/utf8"
)

func TestContainsTag(t *testing.T) {
	tests := []struct {
		tags      string
		searchTag string
		expected  bool
	}{
		// Legacy @-prefixed format
		{"@Projects, @jnl, @public", "public", true},
		{"@Projects, @jnl, @trunk", "public", false},
		{"@secret, @work", "secret", true},
		{"@Projects, @jnl, @trunk", "jnl", true},
		{"@single", "single", true},

		// New bare format (no @ prefix)
		{"Projects, jnl, public", "public", true},
		{"Projects, jnl, trunk", "public", false},
		{"secret, work", "secret", true},
		{"Projects, jnl, trunk", "jnl", true},
		{"single", "single", true},

		// Edge cases
		{"", "anything", false},
		{"public", "public", true},
		{"  public  ", "public", true},
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
		// Legacy @-prefixed format
		{"@Projects, @jnl, @public", false},
		{"@Projects, @secret, @public", true},
		{"@private, @work", true},
		{"@confidential, @notes", true},
		{"@work, @public", false},

		// New bare format (no @ prefix)
		{"Projects, jnl, public", false},
		{"Projects, secret, public", true},
		{"private, work", true},
		{"confidential, notes", true},
		{"work, public", false},

		// Edge cases
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
		"tags":   "Projects, jnl, public, golang",
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
		"tags = [\"Projects\", \"jnl\", \"public\", \"golang\"]", // all tags should be included
		"+++",
	}

	for _, part := range expectedParts {
		if !strings.Contains(result, part) {
			t.Errorf("buildHugoFrontmatter() result missing expected part: %q", part)
		}
	}

	// Check that all tags are included in Hugo frontmatter
	if !strings.Contains(result, "\"public\"") {
		t.Error("buildHugoFrontmatter() should include all tags, including the publish tag")
	}
}

// TestBuildHugoFrontmatterFromYAML verifies the full pipeline from a YAML
// frontmatter file (as jnl now generates) through parseHeader to Hugo TOML
// output. This is the critical path for Hugo publishing.
func TestBuildHugoFrontmatterFromYAML(t *testing.T) {
	tests := []struct {
		name         string
		input        string
		expectedTags string // substring expected in the Hugo output
	}{
		{
			name:         "YAML list tags",
			input:        "---\ndate: 2025-07-05T10:00:00-03:00\ntitle: List Tags\ntags:\n  - dev\n  - public\n  - golang\n---\n\n# Body\n",
			expectedTags: `tags = ["dev", "public", "golang"]`,
		},
		{
			name:         "YAML comma string tags",
			input:        "---\ndate: 2025-07-05T10:00:00-03:00\ntitle: Comma Tags\ntags: dev, public, golang\n---\n\n# Body\n",
			expectedTags: `tags = ["dev", "public", "golang"]`,
		},
		{
			name:         "legacy format with @ prefix",
			input:        ";;; jnl\ndate: 2025-07-05T10:00:00-03:00\ntitle: Legacy Tags\ntags: @dev, @public, @golang\n;;;\n\n# Body\n",
			expectedTags: `tags = ["dev", "public", "golang"]`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			header, body := parseHeader([]byte(tc.input))
			result := buildHugoFrontmatter(header, body)

			if !strings.Contains(result, tc.expectedTags) {
				t.Errorf("Hugo output missing expected tags.\nExpected substring: %s\nGot:\n%s", tc.expectedTags, result)
			}

			// Verify structure: must start and end with +++
			if !strings.HasPrefix(result, "+++\n") {
				t.Error("Hugo frontmatter must start with +++")
			}
			if !strings.Contains(result, "\n+++\n") {
				t.Error("Hugo frontmatter must contain closing +++")
			}
		})
	}
}

// TestPublishEntryHugoContent verifies that a published file contains
// correct Hugo TOML frontmatter with properly formatted tags.
func TestPublishEntryHugoContent(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "hugo-content-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	sourceDir := filepath.Join(tempDir, "source")
	targetDir := filepath.Join(tempDir, "target")
	if err := os.MkdirAll(sourceDir, 0755); err != nil {
		t.Fatalf("Failed to create source dir: %v", err)
	}

	originalPublishTag := publishTag
	originalBlockedTags := blockedTags
	defer func() {
		publishTag = originalPublishTag
		blockedTags = originalBlockedTags
	}()

	publishTag = "public"
	blockedTags = []string{"secret"}

	// Write a YAML-list entry (the format jnl now generates)
	entry := `---
date: 2025-07-05T10:56:44-03:00
title: Hugo Integration Test
tags:
  - golang
  - public
  - tutorial
---

# Getting Started with Go

This is a tutorial about Go.
`
	sourceFile := filepath.Join(sourceDir, "go-tutorial.md")
	if err := os.WriteFile(sourceFile, []byte(entry), 0644); err != nil {
		t.Fatalf("Failed to write source file: %v", err)
	}

	if err := publishEntry(sourceFile, targetDir); err != nil {
		t.Fatalf("publishEntry failed: %v", err)
	}

	targetFile := filepath.Join(targetDir, "go-tutorial.md")
	published, err := os.ReadFile(targetFile)
	if err != nil {
		t.Fatalf("Failed to read published file: %v", err)
	}

	content := string(published)

	checks := []struct {
		desc     string
		expected string
	}{
		{"Hugo TOML opening", "+++\n"},
		{"date field", `date = "2025-07-05T10:56:44-03:00"`},
		{"title field", `title = "Hugo Integration Test"`},
		{"tags array", `tags = ["golang", "public", "tutorial"]`},
		{"body content", "# Getting Started with Go"},
	}

	for _, c := range checks {
		if !strings.Contains(content, c.expected) {
			t.Errorf("Published file missing %s.\nExpected substring: %q\nFull content:\n%s", c.desc, c.expected, content)
		}
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
	content1 := `---
date: 2025-07-05T10:56:44-03:00
title: Test Article
tags:
  - Projects
  - public
---

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
	content2 := `---
date: 2025-07-05T10:56:44-03:00
title: Non-Public Article
tags:
  - Projects
  - private
---

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
	content3 := `---
date: 2025-07-05T10:56:44-03:00
title: Secret Article
tags:
  - Projects
  - public
  - secret
---

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
			`---
date: 2025-07-05T10:56:44-03:00
title: Public Article 1
tags:
  - Projects
  - public
---

# Public Content 1`,
		},
		{
			"public2.md",
			`---
date: 2025-07-05T10:56:44-03:00
title: Public Article 2
tags:
  - Work
  - public
---

# Public Content 2`,
		},
		{
			"private.md",
			`---
date: 2025-07-05T10:56:44-03:00
title: Private Article
tags:
  - Projects
  - work
---

# Private Content`,
		},
		{
			"secret.md",
			`---
date: 2025-07-05T10:56:44-03:00
title: Secret Article
tags:
  - Projects
  - public
  - secret
---

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

// TestBlockedTagsSecurity is a thorough end-to-end test of the blocked-tags
// safety net. It exercises every combination of header format (YAML list,
// YAML comma-string, legacy ;;; with @-prefix) against all configured
// blocked tags, ensuring none of them slip through to publication.
func TestBlockedTagsSecurity(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "blocked-tags-security-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	sourceDir := filepath.Join(tempDir, "source")
	targetDir := filepath.Join(tempDir, "target")
	if err := os.MkdirAll(sourceDir, 0755); err != nil {
		t.Fatalf("Failed to create source dir: %v", err)
	}

	originalPublishTag := publishTag
	originalBlockedTags := blockedTags
	defer func() {
		publishTag = originalPublishTag
		blockedTags = originalBlockedTags
	}()

	publishTag = "public"
	blockedTags = []string{"secret", "private", "confidential"}

	// Each entry has the publish tag AND a blocked tag.
	// None of them should be published.
	entries := []struct {
		name    string
		content string
		desc    string
	}{
		{
			name:    "yaml-list-secret.md",
			content: "---\ntitle: YAML list secret\ntags:\n  - public\n  - secret\n---\n\nBody\n",
			desc:    "YAML list with blocked tag 'secret'",
		},
		{
			name:    "yaml-list-private.md",
			content: "---\ntitle: YAML list private\ntags:\n  - work\n  - public\n  - private\n---\n\nBody\n",
			desc:    "YAML list with blocked tag 'private'",
		},
		{
			name:    "yaml-list-confidential.md",
			content: "---\ntitle: YAML list confidential\ntags:\n  - public\n  - confidential\n---\n\nBody\n",
			desc:    "YAML list with blocked tag 'confidential'",
		},
		{
			name:    "yaml-comma-secret.md",
			content: "---\ntitle: YAML comma secret\ntags: public, secret\n---\n\nBody\n",
			desc:    "YAML comma string with blocked tag 'secret'",
		},
		{
			name:    "legacy-at-secret.md",
			content: ";;; jnl\ntitle: Legacy secret\ntags: @public, @secret\n;;;\n\nBody\n",
			desc:    "Legacy ;;; with @-prefixed blocked tag 'secret'",
		},
		{
			name:    "legacy-at-private.md",
			content: ";;; jnl\ntitle: Legacy private\ntags: @public, @private\n;;;\n\nBody\n",
			desc:    "Legacy ;;; with @-prefixed blocked tag 'private'",
		},
	}

	for _, e := range entries {
		path := filepath.Join(sourceDir, e.name)
		if err := os.WriteFile(path, []byte(e.content), 0644); err != nil {
			t.Fatalf("Failed to write %s: %v", e.name, err)
		}

		err := publishEntry(path, targetDir)
		if err == nil {
			t.Errorf("SECURITY: %s was published but should have been blocked (%s)", e.name, e.desc)
		}

		target := filepath.Join(targetDir, e.name)
		if _, statErr := os.Stat(target); statErr == nil {
			t.Errorf("SECURITY: file %s exists in target dir but should never have been created (%s)", e.name, e.desc)
		}
	}

	// Also verify that a clean entry (public, no blocked tags) IS published.
	cleanFile := filepath.Join(sourceDir, "clean-public.md")
	cleanContent := "---\ntitle: Clean\ntags:\n  - public\n  - work\n---\n\nSafe body\n"
	if err := os.WriteFile(cleanFile, []byte(cleanContent), 0644); err != nil {
		t.Fatalf("Failed to write clean file: %v", err)
	}
	if err := publishEntry(cleanFile, targetDir); err != nil {
		t.Errorf("Clean public entry should have been published: %v", err)
	}
	cleanTarget := filepath.Join(targetDir, "clean-public.md")
	if _, err := os.Stat(cleanTarget); os.IsNotExist(err) {
		t.Error("Clean public entry file was not created in target dir")
	}
}

func TestBuildHugoFrontmatterUTF8Description(t *testing.T) {
	// A long Portuguese text with multi-byte chars that, when truncated
	// at byte boundaries, would produce invalid UTF-8.
	longBody := "Recordando o passado com emoção, lembranças e saudade. " +
		"As memórias são como estrelas no céu: distantes, mas sempre presentes. " +
		"Não há como esquecer os momentos que vivemos juntos naquele verão."

	header := map[string]string{
		"date":  "2026-03-02T07:51:50-03:00",
		"title": "Recordando o passado",
		"tags":  "public",
	}

	result := buildHugoFrontmatter(header, []byte(longBody))

	// Extract the description value from the TOML output
	for line := range strings.SplitSeq(result, "\n") {
		if !strings.HasPrefix(line, "description = ") {
			continue
		}
		// The entire line must be valid UTF-8
		if !utf8.ValidString(line) {
			t.Errorf("description line contains invalid UTF-8: %q", line)
		}
		// Extract value between quotes
		start := strings.Index(line, "\"") + 1
		end := strings.LastIndex(line, "\"")
		desc := line[start:end]
		if !utf8.ValidString(desc) {
			t.Errorf("description value contains invalid UTF-8: %q", desc)
		}
		if len([]rune(desc)) > 150 {
			t.Errorf("description too long: %d runes", len([]rune(desc)))
		}
		return
	}
	t.Error("description field not found in frontmatter output")
}
