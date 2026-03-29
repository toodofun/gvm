package framework

import (
	"os"
	"path/filepath"
	"testing"
)

// SetupTestEnvironment creates a temporary test environment
// Returns temporary directory path and cleanup function
func SetupTestEnvironment(t *testing.T) (string, func()) {
	t.Helper()

	tempDir, err := os.MkdirTemp("", "gvm-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}

	cleanup := func() {
		os.RemoveAll(tempDir)
	}

	return tempDir, cleanup
}

// SetupTestConfig creates test configuration directory structure
func SetupTestConfig(t *testing.T, baseDir string) {
	t.Helper()

	dirs := []string{
		filepath.Join(baseDir, "versions"),
		filepath.Join(baseDir, "archives"),
		filepath.Join(baseDir, "config"),
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatalf("failed to create directory %s: %v", dir, err)
		}
	}
}

// CreateTestFile creates a file in the test directory
func CreateTestFile(t *testing.T, path, content string) {
	t.Helper()

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatalf("failed to create directory: %v", err)
	}

	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}
}
