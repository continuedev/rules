package formats

import (
	"strings"
	"testing"
)

func TestTransformRuleContent_EmptyMetadataHandling(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		format   Format
		expected string
	}{
		{
			name: "Empty description should become empty value not null",
			content: `---
description: 
alwaysApply: true
globs: "**/*.go"
---

# Test Rule

Test content.`,
			format:   GetFormat("continue"),
			expected: "description:",
		},
		{
			name: "Nil description should become empty value not null",
			content: `---
description: null
alwaysApply: true
globs: "**/*.go"
---

# Test Rule

Test content.`,
			format:   GetFormat("continue"),
			expected: "description:",
		},
		{
			name: "Missing description should be handled gracefully",
			content: `---
alwaysApply: true
globs: "**/*.go"
---

# Test Rule

Test content.`,
			format: GetFormat("continue"),
			// Should not contain description field at all when missing
			expected: "alwaysApply: true",
		},
		{
			name: "Empty globs should become empty value not null",
			content: `---
description: "Test rule"
alwaysApply: true
globs: 
---

# Test Rule

Test content.`,
			format:   GetFormat("continue"),
			expected: "globs:",
		},
		{
			name: "Windsurf format with empty description",
			content: `---
description: 
alwaysApply: true
globs: "**/*.go"
---

# Test Rule

Test content.`,
			format:   GetFormat("windsurf"),
			expected: "description:",
		},
		{
			name: "Copilot format with empty description",
			content: `---
description: 
globs: "**/*.go"
---

# Test Rule

Test content.`,
			format:   GetFormat("copilot"),
			expected: "description:",
		},
		{
			name: "Cursor format with no frontmatter - should use fallback",
			content: `# Test Rule

Test content.`,
			format:   GetFormat("cursor"),
			expected: "alwaysApply: true",
		},
		{
			name: "Cursor format with existing frontmatter - should transform normally",
			content: `---
description: "Test rule"
alwaysApply: false
globs: "**/*.go"
---

# Test Rule

Test content.`,
			format:   GetFormat("cursor"),
			expected: "alwaysApply: false",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := TransformRuleContent([]byte(tt.content), tt.format)
			if err != nil {
				t.Fatalf("TransformRuleContent failed: %v", err)
			}

			resultStr := string(result)

			// Should not contain "null" values
			if strings.Contains(resultStr, ": null") {
				t.Errorf("Result should not contain null values. Got:\n%s", resultStr)
			}

			// Should not contain quoted empty strings
			if strings.Contains(resultStr, `: ""`) || strings.Contains(resultStr, `: ''`) {
				t.Errorf("Result should not contain quoted empty strings. Got:\n%s", resultStr)
			}

			// Should contain expected content
			if !strings.Contains(resultStr, tt.expected) {
				t.Errorf("Result should contain '%s'. Got:\n%s", tt.expected, resultStr)
			}

			// Should preserve body content
			if !strings.Contains(resultStr, "# Test Rule") {
				t.Errorf("Result should preserve body content. Got:\n%s", resultStr)
			}
		})
	}
}

func TestCleanMetadataForYAML(t *testing.T) {
	tests := []struct {
		name     string
		input    RuleMetadata
		expected RuleMetadata
	}{
		{
			name: "Nil values become EmptyYAMLValue",
			input: RuleMetadata{
				"description": nil,
				"globs":       "**/*.go",
				"alwaysApply": true,
			},
			expected: RuleMetadata{
				"description": EmptyYAMLValue{},
				"globs":       "**/*.go",
				"alwaysApply": true,
			},
		},
		{
			name: "Empty strings become EmptyYAMLValue",
			input: RuleMetadata{
				"description": "",
				"globs":       "**/*.go",
			},
			expected: RuleMetadata{
				"description": EmptyYAMLValue{},
				"globs":       "**/*.go",
			},
		},
		{
			name: "Non-empty strings are preserved",
			input: RuleMetadata{
				"description": "Test description",
				"globs":       "**/*.go",
			},
			expected: RuleMetadata{
				"description": "Test description",
				"globs":       "**/*.go",
			},
		},
		{
			name: "Mixed types are handled correctly",
			input: RuleMetadata{
				"description": nil,
				"globs":       "",
				"alwaysApply": true,
				"count":       42,
			},
			expected: RuleMetadata{
				"description": EmptyYAMLValue{},
				"globs":       EmptyYAMLValue{},
				"alwaysApply": true,
				"count":       42,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := cleanMetadataForYAML(tt.input)

			// Check each expected key-value pair
			for key, expectedValue := range tt.expected {
				actualValue := result[key]

				// Special handling for EmptyYAMLValue comparison
				if _, expectedIsEmpty := expectedValue.(EmptyYAMLValue); expectedIsEmpty {
					if _, actualIsEmpty := actualValue.(EmptyYAMLValue); !actualIsEmpty {
						t.Errorf("For key '%s': expected EmptyYAMLValue, got %#v", key, actualValue)
					}
				} else if actualValue != expectedValue {
					t.Errorf("For key '%s': expected %#v, got %#v", key, expectedValue, actualValue)
				}
			}

			// Check that no extra keys were added
			if len(result) != len(tt.expected) {
				t.Errorf("Expected %d keys, got %d", len(tt.expected), len(result))
			}
		})
	}
}

func TestTransformRuleContent_YAMLSerialization(t *testing.T) {
	// Test that the cleaned metadata serializes correctly to YAML
	content := `---
description: 
alwaysApply: true
globs: null
tags: []
---

# Test Rule

Test content.`

	result, err := TransformRuleContent([]byte(content), GetFormat("continue"))
	if err != nil {
		t.Fatalf("TransformRuleContent failed: %v", err)
	}

	resultStr := string(result)

	// Verify YAML structure is correct
	if !strings.Contains(resultStr, "---\n") {
		t.Error("Result should start with YAML frontmatter delimiter")
	}

	if !strings.Contains(resultStr, "\n---\n") {
		t.Error("Result should end YAML frontmatter with delimiter")
	}

	// Verify no null values in YAML
	yamlLines := strings.Split(resultStr, "\n")
	for i, line := range yamlLines {
		if strings.Contains(line, ": null") {
			t.Errorf("Line %d should not contain null value: %s", i+1, line)
		}
		if strings.Contains(line, `: ""`) || strings.Contains(line, `: ''`) {
			t.Errorf("Line %d should not contain quoted empty strings: %s", i+1, line)
		}
	}

	// Verify empty values are formatted correctly
	if !strings.Contains(resultStr, "description:") {
		t.Error("Empty description should be formatted as 'description:'")
	}
	if !strings.Contains(resultStr, "globs:") {
		t.Error("Empty globs should be formatted as 'globs:'")
	}
}

func TestIsAlwaysApply(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		expected bool
	}{
		{
			name: "No frontmatter should return true",
			content: `# Test Rule

This is a test rule without frontmatter.`,
			expected: true,
		},
		{
			name: "Frontmatter with alwaysApply: true should return true",
			content: `---
alwaysApply: true
description: "Test rule"
---

# Test Rule

Test content.`,
			expected: true,
		},
		{
			name: "Frontmatter with alwaysApply: false should return false",
			content: `---
alwaysApply: false
description: "Test rule"
---

# Test Rule

Test content.`,
			expected: false,
		},
		{
			name: "Frontmatter with alwaysApply: 'true' (string) should return true",
			content: `---
alwaysApply: "true"
description: "Test rule"
---

# Test Rule

Test content.`,
			expected: true,
		},
		{
			name: "Frontmatter with alwaysApply: 'false' (string) should return false",
			content: `---
alwaysApply: "false"
description: "Test rule"
---

# Test Rule

Test content.`,
			expected: false,
		},
		{
			name: "Frontmatter without alwaysApply field should return true (default behavior)",
			content: `---
name: "Next.js + Common Libraries"
description: "Test rule"
---

# Test Rule

Test content.`,
			expected: true,
		},
		{
			name: "Empty frontmatter should return true",
			content: `---
---

# Test Rule

Test content.`,
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isAlwaysApply([]byte(tt.content))
			if result != tt.expected {
				t.Errorf("isAlwaysApply() = %v, expected %v for content:\n%s", result, tt.expected, tt.content)
			}
		})
	}
}

func TestTransformRuleContent_CursorFallback(t *testing.T) {
	tests := []struct {
		name           string
		content        string
		expectedFields []string
	}{
		{
			name: "No frontmatter should use fallback",
			content: `# Test Rule

This is a test rule without frontmatter.`,
			expectedFields: []string{
				"description:",
				"globs:",
				"alwaysApply: true",
				"# Test Rule",
				"This is a test rule without frontmatter.",
			},
		},
		{
			name: "Empty frontmatter should use fallback",
			content: `---
---

# Test Rule

This is a test rule with empty frontmatter.`,
			expectedFields: []string{
				"description:",
				"globs:",
				"alwaysApply: true",
				"# Test Rule",
				"This is a test rule with empty frontmatter.",
			},
		},
		{
			name: "Existing frontmatter should not use fallback",
			content: `---
description: "Existing rule"
globs: "**/*.js"
alwaysApply: false
---

# Test Rule

This is a test rule with existing frontmatter.`,
			expectedFields: []string{
				"description: Existing rule",
				"globs: '**/*.js'",
				"alwaysApply: false",
				"# Test Rule",
				"This is a test rule with existing frontmatter.",
			},
		},
	}

	format := GetFormat("cursor")

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := TransformRuleContent([]byte(tt.content), format)
			if err != nil {
				t.Fatalf("TransformRuleContent failed: %v", err)
			}

			resultStr := string(result)

			// Check that all expected fields are present
			for _, field := range tt.expectedFields {
				if !strings.Contains(resultStr, field) {
					t.Errorf("Expected field '%s' not found in result:\n%s", field, resultStr)
				}
			}

			// Verify YAML structure is correct for fallback cases
			if strings.Contains(tt.name, "fallback") {
				if !strings.Contains(resultStr, "---\n") {
					t.Error("Result should start with YAML frontmatter delimiter")
				}
				if !strings.Contains(resultStr, "\n---\n") {
					t.Error("Result should end YAML frontmatter with delimiter")
				}
			}
		})
	}
}
