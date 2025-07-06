package main

import (
	"os"
	"path/filepath"
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
	tempDir, err := os.MkdirTemp("", "directory-tags-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

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
