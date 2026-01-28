package main

import (
	"reflect"
	"testing"
)

func TestSortAndUnique(t *testing.T) {
	tests := []struct {
		name     string
		input    []string
		expected []string
	}{
		{
			name:     "no duplicates",
			input:    []string{"c", "a", "b"},
			expected: []string{"a", "b", "c"},
		},
		{
			name:     "simple duplicates",
			input:    []string{"a", "b", "a", "c", "b"},
			expected: []string{"a", "b", "c"},
		},
		{
			name:     "all same",
			input:    []string{"a", "a", "a"},
			expected: []string{"a"},
		},
		{
			name:     "empty",
			input:    []string{},
			expected: []string{},
		},
		{
			name:     "mixed case (case sensitive)",
			input:    []string{"a", "A", "a"},
			expected: []string{"A", "a"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := sortAndUnique(tt.input)

			if !reflect.DeepEqual(got, tt.expected) {
				t.Errorf("sortAndUnique() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestTagProcessingIntegration(t *testing.T) {
	// Simulate the logic in main.go
	tagArray := []string{"work", "projects"}
	directorySpecificTags := []string{"work", "projects"}
	tagArray = append(tagArray, directorySpecificTags...)

	// Sort and deduplicate tags (mocking main.go logic)
	tagArray = sortAndUnique(tagArray)

	expected := []string{"projects", "work"}

	if !reflect.DeepEqual(tagArray, expected) {
		t.Errorf("Integrated tag processing failed. Got %v, want %v", tagArray, expected)
	}
}
