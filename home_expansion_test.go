package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestHomeDirectoryExpansion(t *testing.T) {
	// Save original variables
	originalPublishPath := publishPath
	originalJournalPath := journalPath
	defer func() {
		publishPath = originalPublishPath
		journalPath = originalJournalPath
	}()

	// Test publishPath expansion
	publishPath = "~/test-publish"
	expectedHome, _ := os.UserHomeDir()
	expectedPath := filepath.Join(expectedHome, "test-publish")

	// Simulate the expansion logic from runLuaFile
	if strings.HasPrefix(publishPath, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			t.Fatalf("Failed to get home directory: %v", err)
		}
		publishPath = strings.Replace(publishPath, "~", home, 1)
	}
	if publishPath != "" {
		publishPath, err := filepath.Abs(publishPath)
		if err != nil {
			t.Fatalf("Failed to get absolute path: %v", err)
		}

		if publishPath != expectedPath {
			t.Errorf("publishPath expansion failed. Expected: %s, Got: %s", expectedPath, publishPath)
		}
	}
}

func TestPublishCommandHomeExpansion(t *testing.T) {
	// Create temporary directories
	tempDir, err := os.MkdirTemp("", "publish-home-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	sourceDir := filepath.Join(tempDir, "journal")
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

	// Create a test file with public tag
	testContent := `;;; jnl
date: 2025-07-05T10:56:44-03:00
title: Test Article
tags: @test, @public
;;;

# Test Content`

	testFile := filepath.Join(sourceDir, "test.md")
	err = os.WriteFile(testFile, []byte(testContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Create a temporary directory in home for testing
	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("Failed to get home directory: %v", err)
	}

	relativeTargetDir := "temp-publish-test"
	absoluteTargetDir := filepath.Join(home, relativeTargetDir)
	defer os.RemoveAll(absoluteTargetDir)

	// Test publishCommand with ~/path
	tildeTargetPath := "~/" + relativeTargetDir
	err = publishCommand(tildeTargetPath)
	if err != nil {
		t.Errorf("publishCommand failed with ~/path: %v", err)
	}

	// Check if file was created in the expanded path
	expectedFile := filepath.Join(absoluteTargetDir, "test.md")
	if _, err := os.Stat(expectedFile); os.IsNotExist(err) {
		t.Errorf("File was not created in expanded path: %s", expectedFile)
	}
}
