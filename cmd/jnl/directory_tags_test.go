package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGetDirectoryTags(t *testing.T) {
	// Save original directoryTags
	originalDirectoryTags := directoryTags
	defer func() { directoryTags = originalDirectoryTags }()

	// Setup test directory mapping
	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("Failed to get home directory: %v", err)
	}

	directoryTags = map[string][]string{
		filepath.Join(home, "work"):                    {"work", "secret"},
		filepath.Join(home, "personal"):                {"personal"},
		filepath.Join(home, "projects", "public-blog"): {"public", "blog"},
		"/tmp/test": {"test"},
	}

	tests := []struct {
		name         string
		currentDir   string
		expectedTags []string
	}{
		{
			name:         "exact match work directory",
			currentDir:   filepath.Join(home, "work"),
			expectedTags: []string{"work", "secret"},
		},
		{
			name:         "subdirectory of work",
			currentDir:   filepath.Join(home, "work", "project1"),
			expectedTags: []string{"work", "secret"},
		},
		{
			name:         "deep subdirectory of work",
			currentDir:   filepath.Join(home, "work", "project1", "subdir"),
			expectedTags: []string{"work", "secret"},
		},
		{
			name:         "personal directory",
			currentDir:   filepath.Join(home, "personal"),
			expectedTags: []string{"personal"},
		},
		{
			name:         "public blog directory",
			currentDir:   filepath.Join(home, "projects", "public-blog"),
			expectedTags: []string{"public", "blog"},
		},
		{
			name:         "subdirectory of public blog",
			currentDir:   filepath.Join(home, "projects", "public-blog", "posts"),
			expectedTags: []string{"public", "blog"},
		},
		{
			name:         "no matching directory",
			currentDir:   filepath.Join(home, "other"),
			expectedTags: []string{},
		},
		{
			name:         "test directory",
			currentDir:   "/tmp/test",
			expectedTags: []string{"test"},
		},
		{
			name:         "subdirectory of test",
			currentDir:   "/tmp/test/subdir",
			expectedTags: []string{"test"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := getDirectoryTags(test.currentDir)

			if len(result) != len(test.expectedTags) {
				t.Errorf("Expected %d tags, got %d. Expected: %v, Got: %v",
					len(test.expectedTags), len(result), test.expectedTags, result)
				return
			}

			// Check if all expected tags are present
			expectedMap := make(map[string]bool)
			for _, tag := range test.expectedTags {
				expectedMap[tag] = true
			}

			for _, tag := range result {
				if !expectedMap[tag] {
					t.Errorf("Unexpected tag %s in result %v", tag, result)
				}
				delete(expectedMap, tag)
			}

			if len(expectedMap) > 0 {
				t.Errorf("Missing expected tags: %v", expectedMap)
			}
		})
	}
}

func TestDirectoryTagsIntegration(t *testing.T) {
	// Save original directoryTags
	originalDirectoryTags := directoryTags
	defer func() { directoryTags = originalDirectoryTags }()

	// Setup test configuration
	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("Failed to get home directory: %v", err)
	}

	directoryTags = map[string][]string{
		filepath.Join(home, "work"): {"work", "secret"},
	}

	// Test that directory tags would be added correctly
	wd := filepath.Join(home, "work", "project1")
	tags := getDirectoryTags(wd)

	expectedTags := []string{"work", "secret"}
	if len(tags) != len(expectedTags) {
		t.Errorf("Expected %d tags, got %d", len(expectedTags), len(tags))
	}

	for i, expectedTag := range expectedTags {
		if i >= len(tags) || tags[i] != expectedTag {
			t.Errorf("Expected tag %s at position %d, got %v", expectedTag, i, tags)
		}
	}
}

func TestDirectoryTagsWithRelativePaths(t *testing.T) {
	// Save original directoryTags
	originalDirectoryTags := directoryTags
	defer func() { directoryTags = originalDirectoryTags }()

	// Create a temporary directory structure for testing
	tempDir := t.TempDir()
	var err error

	testWorkDir := filepath.Join(tempDir, "work")
	testProjectDir := filepath.Join(testWorkDir, "project")
	err = os.MkdirAll(testProjectDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create test directories: %v", err)
	}

	// Setup configuration with relative path that we'll resolve
	directoryTags = map[string][]string{
		testWorkDir: {"work", "secret"},
	}

	// Test exact match
	tags := getDirectoryTags(testWorkDir)
	expectedTags := []string{"work", "secret"}
	if len(tags) != len(expectedTags) {
		t.Errorf("Expected %d tags, got %d. Expected: %v, Got: %v",
			len(expectedTags), len(tags), expectedTags, tags)
	}

	// Test subdirectory match
	tags = getDirectoryTags(testProjectDir)
	if len(tags) != len(expectedTags) {
		t.Errorf("Expected %d tags for subdirectory, got %d. Expected: %v, Got: %v",
			len(expectedTags), len(tags), expectedTags, tags)
	}

	// Test non-matching directory
	otherDir := filepath.Join(tempDir, "other")
	tags = getDirectoryTags(otherDir)
	if len(tags) != 0 {
		t.Errorf("Expected no tags for non-matching directory, got %v", tags)
	}
}

// TestDirectoryTagsBlockPublish verifies the full security chain:
// DirectoryTags assigns "secret" → tags flow into buildHeader → the YAML
// header is written and parsed → hasBlockedTag sees "secret" → publish is blocked.
func TestDirectoryTagsBlockPublish(t *testing.T) {
	// Save originals
	originalDirectoryTags := directoryTags
	originalBlockedTags := blockedTags
	originalPublishTag := publishTag
	defer func() {
		directoryTags = originalDirectoryTags
		blockedTags = originalBlockedTags
		publishTag = originalPublishTag
	}()

	publishTag = "public"
	blockedTags = []string{"secret", "private", "confidential"}

	tempDir := t.TempDir()

	workDir := filepath.Join(tempDir, "work", "client-acme")
	if err := os.MkdirAll(workDir, 0755); err != nil {
		t.Fatalf("Failed to mkdir: %v", err)
	}

	// Configure: ~/work → tags "work" + "secret"
	directoryTags = map[string][]string{
		filepath.Join(tempDir, "work"): {"work", "secret"},
	}

	// Simulate the tag-building pipeline from the add command:
	// 1. Start with path segments as tags
	tagArray := []string{"work", "client-acme"}

	// 2. Add git branch (simulate)
	tagArray = append(tagArray, "main")

	// 3. Add directory-specific tags
	dirTags := getDirectoryTags(workDir)
	tagArray = append(tagArray, dirTags...)

	// 4. Sort and deduplicate
	tagArray = sortAndUnique(tagArray)

	// 5. Join (no @ prefix anymore)
	tags := strings.Join(tagArray, ", ")

	// Verify "secret" ended up in the tag list
	if !containsTag(tags, "secret") {
		t.Fatalf("Expected 'secret' tag from DirectoryTags, got tags: %q", tags)
	}

	// Also add "public" so the entry would be a candidate for publishing
	tags = tags + ", public"

	// 6. Build a header and write a journal file
	header := map[string]string{
		"date":  "2026-02-24T10:00:00-03:00",
		"title": "Client ACME Meeting",
		"tags":  tags,
	}
	content := buildHeader(header) + "# Meeting notes\n\nConfidential discussion.\n"

	sourceDir := filepath.Join(tempDir, "journal")
	targetDir := filepath.Join(tempDir, "published")
	if err := os.MkdirAll(sourceDir, 0755); err != nil {
		t.Fatalf("Failed to mkdir: %v", err)
	}

	entryFile := filepath.Join(sourceDir, "acme-meeting.md")
	if err := os.WriteFile(entryFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write entry: %v", err)
	}

	// 7. Attempt to publish — must be BLOCKED by "secret" tag
	err := publishEntry(entryFile, targetDir)
	if err == nil {
		t.Fatal("SECURITY: entry with directory-assigned 'secret' tag was published!")
	}

	targetFile := filepath.Join(targetDir, "acme-meeting.md")
	if _, statErr := os.Stat(targetFile); statErr == nil {
		t.Fatal("SECURITY: published file exists on disk for entry with 'secret' tag!")
	}

	// 8. Verify the round-trip: read back the file and confirm blocked tags survive
	parsedHeader, _ := parseHeader([]byte(content))
	if !hasBlockedTag(parsedHeader["tags"]) {
		t.Errorf("hasBlockedTag should detect blocked tag in parsed header, tags: %q", parsedHeader["tags"])
	}

	// 9. Verify a clean entry from a non-secret directory IS published
	cleanHeader := map[string]string{
		"date":  "2026-02-24T11:00:00-03:00",
		"title": "Public Blog Post",
		"tags":  "blog, public",
	}
	cleanContent := buildHeader(cleanHeader) + "# Blog post\n\nPublic content.\n"
	cleanFile := filepath.Join(sourceDir, "blog-post.md")
	if err := os.WriteFile(cleanFile, []byte(cleanContent), 0644); err != nil {
		t.Fatalf("Failed to write clean entry: %v", err)
	}

	err = publishEntry(cleanFile, targetDir)
	if err != nil {
		t.Errorf("Clean entry without blocked tags should publish: %v", err)
	}
}
