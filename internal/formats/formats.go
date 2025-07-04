package formats

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Format represents a rules format
type Format struct {
	Name            string
	DirectoryPrefix string
	FileExtension   string
	IsSingleFile    bool
	SingleFilePath  string
	Description     string
}

// GetFormat returns a Format for the given format name
func GetFormat(formatName string) Format {
	// Default to ".rules" if no format is specified or format is "default"
	if formatName == "" || formatName == "default" {
		return Format{
			Name:            "default",
			DirectoryPrefix: ".rules",
			FileExtension:   ".md",
			IsSingleFile:    false,
			Description:     "Default rules format",
		}
	}

	// Define formats based on render-formats.md
	switch formatName {
	case "continue":
		return Format{
			Name:            "continue",
			DirectoryPrefix: ".continue/rules",
			FileExtension:   ".md",
			IsSingleFile:    false,
			Description:     "Continue Dev rules",
		}
	case "cursor":
		return Format{
			Name:            "cursor",
			DirectoryPrefix: ".cursor/rules",
			FileExtension:   ".mdc",
			IsSingleFile:    false,
			Description:     "Cursor rules",
		}
	case "windsurf":
		return Format{
			Name:            "windsurf",
			DirectoryPrefix: ".windsurf/rules",
			FileExtension:   ".md",
			IsSingleFile:    false,
			Description:     "Windsurf rules",
		}
	case "claude":
		return Format{
			Name:            "claude",
			DirectoryPrefix: "",
			FileExtension:   ".md",
			IsSingleFile:    true,
			SingleFilePath:  "CLAUDE.md",
			Description:     "Claude Code single file",
		}
	case "copilot":
		return Format{
			Name:            "copilot",
			DirectoryPrefix: ".github/instructions",
			FileExtension:   ".instructions.md",
			IsSingleFile:    false,
			Description:     "GitHub Copilot instructions",
		}
	case "codex":
		return Format{
			Name:            "codex",
			DirectoryPrefix: "",
			FileExtension:   ".md",
			IsSingleFile:    true,
			SingleFilePath:  "AGENT.md",
			Description:     "Codex single file",
		}
	case "cline":
		return Format{
			Name:            "cline",
			DirectoryPrefix: ".clinerules",
			FileExtension:   ".md",
			IsSingleFile:    false,
			Description:     "Cline rules",
		}
	case "cody":
		return Format{
			Name:            "cody",
			DirectoryPrefix: ".sourcegraph",
			FileExtension:   ".rule.md",
			IsSingleFile:    false,
			Description:     "Sourcegraph Cody rules",
		}
	case "amp":
		return Format{
			Name:            "amp",
			DirectoryPrefix: "",
			FileExtension:   ".md",
			IsSingleFile:    true,
			SingleFilePath:  "AGENT.md",
			Description:     "Amp single file",
		}
	default:
		// For any other format, use .<format>/rules
		return Format{
			Name:            formatName,
			DirectoryPrefix: fmt.Sprintf(".%s/rules", formatName),
			FileExtension:   ".md",
			IsSingleFile:    false,
			Description:     fmt.Sprintf("%s rules", formatName),
		}
	}
}

// InitializeFormat creates the directory structure for a format
func InitializeFormat(formatName string) error {
	format := GetFormat(formatName)

	if !format.IsSingleFile {
		dirPath := format.DirectoryPrefix
		if err := os.MkdirAll(dirPath, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dirPath, err)
		}
	}

	// Create rules.json in the root directory if it doesn't exist
	rulesJSONPath := "rules.json"
	if _, err := os.Stat(rulesJSONPath); os.IsNotExist(err) {
		defaultRulesJSON := `{
  "name": "new-rules",
  "description": "",
  "author": "",
  "license": "CC0-1.0",
  "version": "1.0.0",
  "rules": {}
}`

		if err := os.WriteFile(rulesJSONPath, []byte(defaultRulesJSON), 0644); err != nil {
			return fmt.Errorf("failed to create rules.json: %w", err)
		}
	}

	return nil
}

// GetRulesDirectory returns the rules directory for a given format
func GetRulesDirectory(formatName string) (string, error) {
	format := GetFormat(formatName)
	return format.DirectoryPrefix, nil
}

// GetRulesJSONPath returns the path to rules.json file in the root directory
func GetRulesJSONPath(formatName string) (string, error) {
	// Return the path to rules.json in the root directory
	return "rules.json", nil
}

// FindRulesFormats looks for any top-level folder with the structure ".{folder-name}/rules"
// and returns a list of folder names without the dot prefix
func FindRulesFormats() ([]string, error) {
	entries, err := os.ReadDir(".")
	if err != nil {
		return nil, err
	}

	var formatFolders []string
	for _, entry := range entries {
		if !entry.IsDir() || !strings.HasPrefix(entry.Name(), ".") {
			continue
		}

		// Check if this directory has a "rules" subdirectory
		rulesDir := filepath.Join(entry.Name(), "rules")
		if info, err := os.Stat(rulesDir); err == nil && info.IsDir() {
			// Found a format folder - add the format name without the dot prefix
			formatName := strings.TrimPrefix(entry.Name(), ".")
			formatFolders = append(formatFolders, formatName)
		}
	}

	return formatFolders, nil
}

// GetFormatSuggestionMessage returns a standardized message for suggesting format rendering
// when existing rule format folders are found
func GetFormatSuggestionMessage() (string, error) {
	formatFolders, err := FindRulesFormats()
	if err != nil {
		return "", err
	}

	if len(formatFolders) > 0 {
		return fmt.Sprintf("Found existing rules folder(s): %s\nConsider running 'rules render %s' to initialize rules.json from existing rules",
			strings.Join(formatFolders, ", "), formatFolders[0]), nil
	}

	return "", nil
}

// RenderRules renders rules from the source directory to the target format
func RenderRules(sourceDir string, targetFormat Format) error {
	// Use the transformer to process the rule files
	return ProcessRuleFiles(sourceDir, targetFormat)
}

// GetAllFormats returns a list of all supported formats
func GetAllFormats() []Format {
	formats := []Format{
		GetFormat("continue"),
		GetFormat("cursor"),
		GetFormat("windsurf"),
		GetFormat("claude"),
		GetFormat("copilot"),
		GetFormat("codex"),
		GetFormat("cline"),
		GetFormat("cody"),
		GetFormat("amp"),
	}
	return formats
}
