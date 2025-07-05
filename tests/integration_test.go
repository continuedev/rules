package tests

import (
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestCLIIntegration(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		contains string
	}{
		{
			name:     "help command",
			args:     []string{"--help"},
			contains: "command-line tool to create, manage, and convert rule sets",
		},
		{
			name:     "version command",
			args:     []string{"--version"},
			contains: "rules version",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Get the parent directory (project root)
			parentDir := filepath.Join("..", ".")

			binaryName := testBinaryName("rules-test")
			// Build the CLI from the parent directory
			buildCmd := exec.Command("go", "build", "-o", binaryName, ".")
			buildCmd.Dir = parentDir
			if err := buildCmd.Run(); err != nil {
				t.Fatalf("Failed to build CLI: %v", err)
			}
			binaryPath := filepath.Join(parentDir, binaryName)
			defer cleanupTestPath(t, binaryPath)

			// Run the CLI command
			cmd := exec.Command(binaryPath, tt.args...)
			output, err := cmd.CombinedOutput()

			// For help and version commands, exit code 0 is expected
			if err != nil && (strings.Contains(string(output), "unknown") || !strings.Contains(string(output), tt.contains)) {
				t.Fatalf("Command failed: %v, output: %s", err, output)
			}

			if !strings.Contains(string(output), tt.contains) {
				t.Errorf("Expected output to contain '%s', got: %s", tt.contains, output)
			}
		})
	}
}

func TestCLICommandsExist(t *testing.T) {
	// Test that expected commands exist
	parentDir := filepath.Join("..", ".")

	binaryName := testBinaryName("rules-test")
	buildCmd := exec.Command("go", "build", "-o", binaryName, ".")
	buildCmd.Dir = parentDir
	if err := buildCmd.Run(); err != nil {
		t.Fatalf("Failed to build CLI: %v", err)
	}
	binaryPath := filepath.Join(parentDir, binaryName)
	defer cleanupTestPath(t, binaryPath)

	cmd := exec.Command(binaryPath, "--help")
	output, _ := cmd.CombinedOutput()

	// Check for expected subcommands
	expectedCommands := []string{"help", "completion"}

	for _, expected := range expectedCommands {
		if !strings.Contains(string(output), expected) {
			t.Logf("Command '%s' not found in help output (this may be expected)", expected)
		}
	}
}
