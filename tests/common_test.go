package tests

import (
	"os"
	"runtime"
	"testing"
)

func testBinaryName(name string) string {
	if runtime.GOOS == "windows" {
		name += ".exe"
	}
	return name
}

func cleanupTestPath(t *testing.T, path string) {
	t.Helper()
	if err := os.RemoveAll(path); err != nil && !os.IsNotExist(err) {
		t.Fatalf("Failed to remove %s: %v", path, err)
	}
}
