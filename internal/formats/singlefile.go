package formats

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// renderToSingleFile renders all rules with alwaysApply: true to a single file
func renderToSingleFile(sourceDir string, format Format) error {
	var combinedContent bytes.Buffer

	// Add a header to the file
	combinedContent.WriteString("# Rules\n\n")

	// Walk through all files in the source directory
	err := filepath.Walk(sourceDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if info.IsDir() {
			return nil
		}

		// Only process Markdown files
		if !strings.HasSuffix(path, ".md") {
			return nil
		}

		// Read file
		content, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("failed to read file %s: %w", path, err)
		}

		// Check if the file has alwaysApply: true in the frontmatter
		if isAlwaysApply(content) {
			// Get the rule name from the path
			ruleName, err := GetRuleName(path, sourceDir)
			if err != nil {
				return fmt.Errorf("failed to get rule name: %w", err)
			}

			// Try to extract a better title from the content
			title := ExtractRuleTitle(content)
			if title == "" {
				// Fall back to the rule name
				title = ruleName
			}

			// Add the rule to the combined content
			combinedContent.WriteString(fmt.Sprintf("## %s\n\n", title))

			// Add the content without frontmatter, trimming any leading whitespace
			ruleContent := stripFrontmatter(content)
			// Ensure there's no trailing whitespace after each rule
			ruleContent = bytes.TrimRight(ruleContent, " \t\n\r")
			combinedContent.Write(ruleContent)
			combinedContent.WriteString("\n\n")
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("failed to process rules: %w", err)
	}

	// Write the combined content to the target file
	if err := os.WriteFile(format.SingleFilePath, combinedContent.Bytes(), 0644); err != nil {
		return fmt.Errorf("failed to write to target file %s: %w", format.SingleFilePath, err)
	}

	return nil
}

// isAlwaysApply checks if a rule file has alwaysApply: true in the frontmatter
func isAlwaysApply(content []byte) bool {
	// First, check if the file has frontmatter at all
	hasFrontmatter := hasFrontmatter(content)
	if !hasFrontmatter {
		// If there's no frontmatter, treat it as alwaysApply: true
		return true
	}

	// Parse the frontmatter
	metadata, _, err := ParseFrontmatter(content)
	if err != nil {
		return false
	}

	// Check if alwaysApply is true
	if alwaysApply, ok := metadata["alwaysApply"]; ok {
		// Handle different types of values
		switch v := alwaysApply.(type) {
		case bool:
			return v
		case string:
			return strings.ToLower(v) == "true"
		}
	}

	// If alwaysApply is not explicitly set, default to true
	// This ensures rules without the alwaysApply field are included in single-file formats
	return true
}

// hasFrontmatter checks if a file has frontmatter
func hasFrontmatter(content []byte) bool {
	// Check if the file starts with "---"
	lines := bytes.SplitN(content, []byte("\n"), 2)
	if len(lines) == 0 {
		return false
	}

	return bytes.Equal(bytes.TrimSpace(lines[0]), []byte("---"))
}

// stripFrontmatter removes the frontmatter from a rule file
// and trims any leading whitespace from the markdown content
func stripFrontmatter(content []byte) []byte {
	// Check if the file has frontmatter
	if !hasFrontmatter(content) {
		// If there's no frontmatter, return the original content
		return content
	}

	_, bodyContent, err := ParseFrontmatter(content)
	if err != nil {
		// If there's an error parsing the frontmatter, return the original content
		return content
	}

	// Trim leading whitespace from the body content
	return bytes.TrimLeft(bodyContent, " \t\n\r")
}
