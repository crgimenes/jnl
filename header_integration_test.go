package main

import (
	"reflect"
	"testing"
	"time"
)

func TestHeaderIntegration(t *testing.T) {
	// Test the integration between buildHeader and parseHeader
	originalInfo := map[string]string{
		"date":   time.Now().Format(time.RFC3339),
		"dir":    "~/Projects/test",
		"user":   "testuser",
		"branch": "main",
		"tags":   "@test, @go, @main",
		"title":  "Test Entry",
	}

	// Build header
	headerStr := buildHeader(originalInfo)

	// Add some body content
	fullContent := headerStr + "# Test Content\n\nThis is a test entry.\n"

	// Parse back
	parsedInfo, body := parseHeader([]byte(fullContent))

	// Check if all original info is preserved
	if !reflect.DeepEqual(originalInfo, parsedInfo) {
		t.Errorf("Header info mismatch.\nOriginal: %v\nParsed: %v", originalInfo, parsedInfo)
	}

	// Check body content (buildHeader adds a blank line after header)
	expectedBody := "\n# Test Content\n\nThis is a test entry.\n"
	if string(body) != expectedBody {
		t.Errorf("Body mismatch.\nExpected: %q\nGot: %q", expectedBody, string(body))
	}
}

func TestJournalFilenameWithHeader(t *testing.T) {
	// Test journalFilename function with header containing title
	headerInfo := map[string]string{
		"title": "My Important Meeting Notes",
		"date":  "2025-07-05T10:30:00-03:00",
	}

	content := buildHeader(headerInfo) + "Meeting details here..."

	filename := journalFilename([]byte(content))

	// Should use title from header for filename
	if !contains(filename, "my-important-meeting-notes") {
		t.Errorf("Expected filename to contain slugified title, got: %s", filename)
	}
}

func TestJournalFilenameWithoutHeaderTitle(t *testing.T) {
	// Test journalFilename function without title in header
	headerInfo := map[string]string{
		"date": "2025-07-05T10:30:00-03:00",
	}

	content := buildHeader(headerInfo) + "# Project Update\n\nSome content here..."

	filename := journalFilename([]byte(content))

	// Should use first line of body for filename
	if !contains(filename, "project-update") {
		t.Errorf("Expected filename to contain slugified first line, got: %s", filename)
	}
}

// Helper function to check if string contains substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) &&
		s[:len(substr)] != "" &&
		findSubstring(s, substr) >= 0
}

func findSubstring(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}
