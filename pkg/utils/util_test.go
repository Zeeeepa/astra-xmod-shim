package utils

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGenerateSimpleID(t *testing.T) {
	// Test generating simple ID
	id1 := GenerateSimpleID()
	id2 := GenerateSimpleID()

	// Verify ID format
	assert.NotEmpty(t, id1)
	assert.NotEmpty(t, id2)
	assert.Len(t, id1, 8) // Should be 8-character hex string
	assert.Len(t, id2, 8)

	// Verify ID uniqueness (note: extremely low probability of collision)
	assert.NotEqual(t, id1, id2)

	// Verify ID contains only hex characters
	assert.Regexp(t, "^[0-9a-f]{8}$", id1)
	assert.Regexp(t, "^[0-9a-f]{8}$", id2)
}

func TestGenerateSimpleIDConsistency(t *testing.T) {
	// Test ID generation consistency
	ids := make(map[string]bool)

	// Generate 100 IDs and verify no duplicates
	for i := 0; i < 100; i++ {
		id := GenerateSimpleID()
		assert.False(t, ids[id], "Generated duplicate ID: %s", id)
		ids[id] = true
	}

	// Verify all IDs are valid
	for id := range ids {
		assert.Len(t, id, 8)
		assert.Regexp(t, "^[0-9a-f]{8}$", id)
	}
}

func TestIsLetter(t *testing.T) {
	tests := []struct {
		char     byte
		expected bool
	}{{
			char:     'a',
			expected: true,
		}, {
			char:     'z',
			expected: true,
		}, {
			char:     'm',
			expected: true,
		}, {
			char:     'A',
			expected: false, // Uppercase returns false
		}, {
			char:     '0',
			expected: false, // Digit returns false
		}, {
			char:     '-',
			expected: false, // Hyphen returns false
		}, {
			char:     '@',
			expected: false, // Special char returns false
		},
	}

	for _, tt := range tests {
		result := isLetter(tt.char)
		assert.Equal(t, tt.expected, result, "Character %c should return %v", tt.char, tt.expected)
	}
}

func TestIsAlnum(t *testing.T) {
	tests := []struct {
		char     byte
		expected bool
	}{{
			char:     'a',
			expected: true,
		}, {
			char:     'z',
			expected: true,
		}, {
			char:     'm',
			expected: true,
		}, {
			char:     '0',
			expected: true, // Digit returns true
		}, {
			char:     '9',
			expected: true, // Digit returns true
		}, {
			char:     '5',
			expected: true, // Digit returns true
		}, {
			char:     'A',
			expected: false, // Uppercase returns false
		}, {
			char:     '-',
			expected: false, // Hyphen returns false
		}, {
			char:     '@',
			expected: false, // Special char returns false
		},
	}

	for _, tt := range tests {
		result := isAlnum(tt.char)
		assert.Equal(t, tt.expected, result, "Character %c should return %v", tt.char, tt.expected)
	}
}

func TestModelNameToDeploymentNameEdgeCases(t *testing.T) {
	// Test edge cases
	tests := []struct {
		name      string
		modelName string
		expected  string
	}{{
			name:      "single character (letter)",
			modelName: "a",
			expected:  "a", // Starts with letter → no prefix
		}, {
			name:      "single digit",
			modelName: "1",
			expected:  "model-1", // Starts with non-letter → add prefix
		}, {
			name:      "single special character",
			modelName: "@",
			expected:  "model", // Empty after cleanup → fallback to "model"
		}, {
			name:      "exactly 63 characters (letter start)",
			modelName: strings.Repeat("a", 63),
			expected:  strings.Repeat("a", 63), // No prefix, exactly 63 chars
		}, {
			name:      "64 characters (should be truncated)",
			modelName: strings.Repeat("a", 64),
			expected:  strings.Repeat("a", 63), // Truncate to 63 chars
		}, {
			name:      "exactly 63 characters starting with number",
			modelName: "1" + strings.Repeat("a", 62),
			expected:  "model-1" + strings.Repeat("a", 56), // "model-" (6) + "1" + 56*a = 63
		}, {
			name:      "empty string",
			modelName: "",
			expected:  "model", // Function fallback
		}, {
			name:      "only hyphens",
			modelName: "-----",
			expected:  "model", // Empty after cleanup → "model"
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ModelNameToDeploymentName(tt.modelName)
			assert.Equal(t, tt.expected, result, "Input: %q", tt.modelName)
			assert.True(t, len(result) <= 63, "Result should not exceed 63 characters")

			// Verify result contains only valid characters
			assert.Regexp(t, "^[a-z0-9-]*$", result)

			// Verify does not start or end with hyphen
			if result != "" {
				assert.False(t, strings.HasPrefix(result, "-"), "Should not start with '-'")
				assert.False(t, strings.HasSuffix(result, "-"), "Should not end with '-'")
			}
		})
	}
}

func TestModelNameToDeploymentNamePerformance(t *testing.T) {
	// Performance and compliance tests
	testCases := []struct {
		modelName string
	}{{
			modelName: "llama",
		}, {
			modelName: "llama-2-7b-chat-hf",
		}, {
			modelName: "meta-llama/Llama-2-7b-chat-hf",
		}, {
			modelName: "models--meta-llama--Llama-2-7b-chat-hf",
		}, {
			modelName: "very-long-model-name-with-many-characters-and-special-chars-@#$%^&*()",
		}, {
			modelName: "1model",
		}, {
			modelName: "@#$%",
		}, {
			modelName: "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.modelName, func(t *testing.T) {
			result := ModelNameToDeploymentName(tc.modelName)
			assert.NotEmpty(t, result)
			assert.True(t, len(result) <= 63)

			// Verify contains only lowercase letters, numbers, and hyphens
			assert.Regexp(t, "^[a-z0-9-]*$", result)

			// If result is not empty, must start with a letter (K8s requirement)
			if result != "" {
				assert.True(t, result[0] >= 'a' && result[0] <= 'z',
					"Result must start with a lowercase letter: %q", result)
			}

			// Does not end with hyphen
			assert.False(t, strings.HasSuffix(result, "-"), "Should not end with '-'")
		})
	}
}

func BenchmarkGenerateSimpleID(b *testing.B) {
	for i := 0; i < b.N; i++ {
		GenerateSimpleID()
	}
}

func BenchmarkModelNameToDeploymentName(b *testing.B) {
	modelNames := []string{
		"llama-2-7b-chat-hf",
		"meta-llama/Llama-2-7b-chat-hf",
		"very-long-model-name-with-special-chars",
		"1model",
		"@#$",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ModelNameToDeploymentName(modelNames[i%len(modelNames)])
	}
}