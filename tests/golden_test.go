package tests

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// CommandConfig represents a mapping between a command and its golden file
type CommandConfig struct {
	Command    string
	GoldenFile string
}

// readCommandConfigs reads the configuration file and returns a map of golden files to commands
func readCommandConfigs() (map[string]string, error) {
	configFile := "../tests/golden_commands.txt"
	file, err := os.Open(configFile)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// Map golden files to their commands
	goldenToCommand := make(map[string]string)

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		// Skip comments and empty lines
		if strings.HasPrefix(line, "#") || strings.TrimSpace(line) == "" {
			continue
		}

		parts := strings.Split(line, "|")
		if len(parts) == 2 {
			cmd := strings.TrimSpace(parts[0])
			goldenFile := strings.TrimSpace(parts[1])
			// Store with the golden file as the key
			goldenToCommand[goldenFile] = cmd
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return goldenToCommand, nil
}

// matchWithPlaceholders checks if the output matches the expected text with placeholders
func matchWithPlaceholders(actual, expected string) bool {
	// First try exact match
	if actual == expected {
		return true
	}

	// Handle login.golden placeholders
	if strings.Contains(expected, "<STATE_PLACEHOLDER>") {
		// Extract parts before and after the placeholder
		beforeState := expected[:strings.Index(expected, "<STATE_PLACEHOLDER>")]
		afterState := expected[strings.Index(expected, "<STATE_PLACEHOLDER>")+len("<STATE_PLACEHOLDER>"):]

		// Check if the output contains the parts before and after
		if !strings.HasPrefix(actual, beforeState) {
			return false
		}

		// Handle the error placeholder if it exists
		if strings.Contains(afterState, "<ERROR_PLACEHOLDER>") {
			beforeError := afterState[:strings.Index(afterState, "<ERROR_PLACEHOLDER>")]
			afterError := afterState[strings.Index(afterState, "<ERROR_PLACEHOLDER>")+len("<ERROR_PLACEHOLDER>"):]

			// The middle part should be the UUID - find where the beforeError starts
			stateEnd := strings.Index(actual[len(beforeState):], beforeError) + len(beforeState)
			if stateEnd < len(beforeState) {
				return false
			}

			// Extract the UUID
			stateValue := actual[len(beforeState):stateEnd]

			// Validate UUID format (simple check)
			if !strings.Contains(stateValue, "-") || len(stateValue) < 30 {
				return false
			}

			// Remaining text after error message
			remainingText := actual[stateEnd+len(beforeError):]

			// Check if error message is valid (either "unexpected newline" or "EOF")
			if !strings.HasPrefix(remainingText, "unexpected newline") &&
				!strings.HasPrefix(remainingText, "EOF") {
				return false
			}

			// Check the final part after the error message
			errorMsgEnd := 0
			if strings.HasPrefix(remainingText, "unexpected newline") {
				errorMsgEnd = len("unexpected newline")
			} else if strings.HasPrefix(remainingText, "EOF") {
				errorMsgEnd = len("EOF")
			}

			return strings.HasPrefix(remainingText[errorMsgEnd:], afterError)
		} else {
			// Just check if the remaining text appears after the state
			stateEnd := strings.Index(actual[len(beforeState):], afterState)
			return stateEnd >= 0
		}
	}

	return false
}

func TestGoldenFiles(t *testing.T) {
	// Read command configurations
	goldenToCommand, err := readCommandConfigs()
	if err != nil {
		t.Fatalf("Failed to read command configurations: %v", err)
	}

	if len(goldenToCommand) == 0 {
		t.Fatal("No commands configured. Check golden_commands.txt file.")
	}

	// Find all golden files
	goldenFiles, err := filepath.Glob("golden/**/*.golden")
	if err != nil {
		t.Fatalf("Failed to find golden files: %v", err)
	}

	if len(goldenFiles) == 0 {
		t.Fatal("No golden files found. Run scripts/generate_golden.sh first.")
	}

	// Make sure the CLI binary exists
	cliPath := testBinaryName("../rules-cli")
	if _, err := os.Stat(cliPath); os.IsNotExist(err) {
		// Build the CLI if it doesn't exist
		// Get version from package.json
		versionCmd := exec.Command("node", "-p", "require('../package.json').version")
		versionBytes, err := versionCmd.Output()
		if err != nil {
			t.Fatalf("Failed to get version from package.json: %v", err)
		}
		version := strings.TrimSpace(string(versionBytes))

		cmd := exec.Command("go", "build", "-ldflags", fmt.Sprintf("-X main.Version=%s -X rules-cli/internal/utils.Version=%s", version, version), "-o", cliPath, "../main.go")
		if err := cmd.Run(); err != nil {
			t.Fatalf("Failed to build CLI: %v", err)
		}
	}

	for _, goldenFile := range goldenFiles {
		t.Run(goldenFile, func(t *testing.T) {
			// Look up the command for this golden file
			slashPath := filepath.ToSlash(goldenFile)
			cmd, found := goldenToCommand["tests/"+slashPath]
			if !found {
				t.Skipf("No command configured for golden file: %s", goldenFile)
				return
			}

			// Create a temporary directory for this test
			tempDir, err := os.MkdirTemp("", "rules-cli-test-*")
			if err != nil {
				t.Fatalf("Failed to create temporary directory: %v", err)
			}
			defer cleanupTestPath(t, tempDir)

			// Copy the CLI binary to the temporary directory
			tempCliPath := filepath.Join(tempDir, testBinaryName("rules-cli"))
			if err := copyFile(cliPath, tempCliPath); err != nil {
				t.Fatalf("Failed to copy CLI binary: %v", err)
			}

			// Change to the temporary directory
			originalDir, err := os.Getwd()
			if err != nil {
				t.Fatalf("Failed to get current directory: %v", err)
			}
			defer os.Chdir(originalDir)

			if err := os.Chdir(tempDir); err != nil {
				t.Fatalf("Failed to change to temporary directory: %v", err)
			}

			// Run prerequisite commands based on the current command
			if strings.HasPrefix(cmd, "add ") {
				// For add commands, run init first to create rules.json
				initCmd := exec.Command(tempCliPath, "init")
				if err := initCmd.Run(); err != nil {
					t.Logf("Init command failed (this might be expected): %v", err)
				}
			} else if cmd == "install" {
				// For install command, run init and add to set up rules.json with rules
				initCmd := exec.Command(tempCliPath, "init")
				if err := initCmd.Run(); err != nil {
					t.Logf("Init command failed (this might be expected): %v", err)
				}
				addCmd := exec.Command(tempCliPath, "add", "starter/nextjs-rules")
				if err := addCmd.Run(); err != nil {
					t.Logf("Add command failed (this might be expected): %v", err)
				}
			} else if strings.HasPrefix(cmd, "remove ") {
				// For remove commands, run init and add to set up rules.json with rules
				initCmd := exec.Command(tempCliPath, "init")
				if err := initCmd.Run(); err != nil {
					t.Logf("Init command failed (this might be expected): %v", err)
				}
				addCmd := exec.Command(tempCliPath, "add", "starter/nextjs-rules")
				if err := addCmd.Run(); err != nil {
					t.Logf("Add command failed (this might be expected): %v", err)
				}
			}

			// Split the command into arguments
			var args []string
			if cmd != "" {
				args = strings.Fields(cmd)
			}

			// Run the command
			execCmd := exec.Command(tempCliPath, args...)
			output, err := execCmd.CombinedOutput()

			// We don't fail the test if the command returns non-zero
			// because some golden files might be testing error cases

			// Read expected output
			expectedBytes, err := os.ReadFile(filepath.Join(originalDir, goldenFile))
			if err != nil {
				t.Fatalf("Failed to read golden file: %v", err)
			}

			expected := string(expectedBytes)
			actual := string(output)

			// Normalize line endings for cross-platform compatibility
			expected = strings.ReplaceAll(expected, "\r\n", "\n")
			actual = strings.ReplaceAll(actual, "\r\n", "\n")
			actual = filepath.ToSlash(actual)

			// Special case for login test which has placeholders
			if strings.Contains(expected, "<STATE_PLACEHOLDER>") {
				if !matchWithPlaceholders(actual, expected) {
					t.Errorf("Output does not match golden file with placeholders.\nCommand: %s\nExpected:\n%s\n\nGot:\n%s",
						cmd, expected, actual)
				}
			} else if strings.Contains(expected, "<VERSION_PLACEHOLDER>") {
				// Match version output with placeholder
				// Replace the actual version in the output with the placeholder and compare
				actualNormalized := versionPlaceholderNormalize(actual)
				expectedNormalized := expected

				if actualNormalized != expectedNormalized {
					t.Errorf("Output does not match golden file with version placeholder.\nCommand: %s\nExpected:\n%s\n\nGot (normalized):\n%s\n\nOriginal Output:\n%s",
						cmd, expectedNormalized, actualNormalized, actual)
				}
			} else {
				// Standard equality check for other files
				if actual != expected {
					t.Errorf("Output does not match golden file.\nCommand: %s\nExpected:\n%s\n\nGot:\n%s",
						cmd, expected, actual)
				}
			}
		})
	}
}

// versionPlaceholderNormalize replaces the version number in the version output with the placeholder
func versionPlaceholderNormalize(s string) string {
	lines := strings.Split(s, "\n")
	for i, line := range lines {
		if strings.HasPrefix(line, "rules version ") {
			// Replace everything after 'rules version '
			lines[i] = "rules version <VERSION_PLACEHOLDER>"
		}

		// Check for version in add command output
		if strings.Contains(line, "(version ") && !strings.Contains(line, "<VERSION_PLACEHOLDER>") {
			// Replace the version between "(version " and ")"
			parts := strings.SplitN(line, "(version ", 2)
			if len(parts) == 2 {
				versionPart := strings.SplitN(parts[1], ")", 2)
				if len(versionPart) == 2 {
					lines[i] = parts[0] + "(version <VERSION_PLACEHOLDER>)" + versionPart[1]
				}
			}
		}
	}
	return strings.Join(lines, "\n")
}

// copyFile copies a file from src to dst
func copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = destFile.ReadFrom(sourceFile)
	if err != nil {
		return err
	}

	// Make the copied file executable
	return os.Chmod(dst, 0755)
}

// TestHelp provides a simple example to ensure our testing framework works
func TestHelp(t *testing.T) {
	// This test just makes sure the CLI can run and produce some output
	// It's useful for initial verification of the test setup

	cmd := exec.Command("../rules-cli", "--help")
	output, err := cmd.CombinedOutput()

	if err != nil {
		t.Logf("Command output: %s", string(output))
		t.Fatalf("Command failed: %v", err)
	}

	if len(output) == 0 {
		t.Error("Help command produced no output")
	}
}
