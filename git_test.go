package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestIsInGitRepository(t *testing.T) {
	// Create a temporary directory structure for testing
	tempDir, err := os.MkdirTemp("", "git-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Test case 1: Directory without .git
	subDir1 := filepath.Join(tempDir, "no-git")
	err = os.MkdirAll(subDir1, 0755)
	if err != nil {
		t.Fatalf("Failed to create subdir: %v", err)
	}

	// Change to the directory without .git
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)

	os.Chdir(subDir1)
	if isInGitRepository() {
		t.Error("Should not detect git repository in directory without .git")
	}

	// Test case 2: Directory with .git folder
	subDir2 := filepath.Join(tempDir, "with-git")
	err = os.MkdirAll(subDir2, 0755)
	if err != nil {
		t.Fatalf("Failed to create subdir: %v", err)
	}

	gitDir := filepath.Join(subDir2, ".git")
	err = os.MkdirAll(gitDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create .git dir: %v", err)
	}

	os.Chdir(subDir2)
	if !isInGitRepository() {
		t.Error("Should detect git repository in directory with .git folder")
	}

	// Test case 3: Nested directory in git repository
	nestedDir := filepath.Join(subDir2, "nested", "deep")
	err = os.MkdirAll(nestedDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create nested dir: %v", err)
	}

	os.Chdir(nestedDir)
	if !isInGitRepository() {
		t.Error("Should detect git repository in nested directory within git repo")
	}

	// Test case 4: Git worktree (where .git is a file)
	subDir3 := filepath.Join(tempDir, "worktree")
	err = os.MkdirAll(subDir3, 0755)
	if err != nil {
		t.Fatalf("Failed to create subdir: %v", err)
	}

	gitFile := filepath.Join(subDir3, ".git")
	err = os.WriteFile(gitFile, []byte("gitdir: /some/path/to/git"), 0644)
	if err != nil {
		t.Fatalf("Failed to create .git file: %v", err)
	}

	os.Chdir(subDir3)
	if !isInGitRepository() {
		t.Error("Should detect git repository when .git is a file (worktree)")
	}
}

func TestGetGitBranch(t *testing.T) {
	// This test will work in the actual project directory
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)

	// Test in a non-git directory
	tempDir, err := os.MkdirTemp("", "non-git-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	os.Chdir(tempDir)
	branch, ok := getGitBranch()
	if ok {
		t.Errorf("Should not get git branch in non-git directory, got: %s", branch)
	}

	// Test in the actual project directory (assuming this is a git repo)
	os.Chdir(originalDir)
	if !isInGitRepository() {
		t.Skip("Not in a git repository, skipping git branch test")
	}
	branch, ok = getGitBranch()
	if !ok {
		t.Error("Should get git branch in actual git repository")
	} else if branch == "" {
		t.Error("Git branch should not be empty in actual git repository")
	}
	t.Logf("Current git branch: %s", branch)
}
