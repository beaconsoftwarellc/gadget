package stringutil

import (
	"testing"
)

func TestPlural(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "single consonant ending in y",
			input:    "berry",
			expected: "berries",
		},
		{
			name:     "word ending with s",
			input:    "class",
			expected: "classes",
		},
		{
			name:     "word ending with ch",
			input:    "match",
			expected: "matches",
		},
		{
			name:     "word ending with sh",
			input:    "brush",
			expected: "brushes",
		},
		{
			name:     "word ending with o after a consonant",
			input:    "hero",
			expected: "heroes",
		},
		{
			name:     "word ending with f",
			input:    "wolf",
			expected: "wolves",
		},
		{
			name:     "word ending with fe",
			input:    "knife",
			expected: "knives",
		},
		{
			name:     "word ending in y after a vowel",
			input:    "key",
			expected: "keys",
		},
		{
			name:     "word ending with x",
			input:    "box",
			expected: "boxes",
		},
		{
			name:     "word shorter than 3 characters",
			input:    "it",
			expected: "it",
		},
		{
			name:     "general case with no special rules",
			input:    "cat",
			expected: "cats",
		},
		{
			name:     "word ending with z",
			input:    "quiz",
			expected: "quizzes",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Plural(tt.input)
			if result != tt.expected {
				t.Errorf("Plural(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}
