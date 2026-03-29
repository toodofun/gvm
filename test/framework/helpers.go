package framework

import (
	"os"
	"strings"
	"testing"
)

// AssertDirExists asserts that a directory exists
func AssertDirExists(t *testing.T, path string) {
	t.Helper()

	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("directory %s does not exist: %v", path, err)
	}

	if !info.IsDir() {
		t.Fatalf("%s is not a directory", path)
	}
}

// AssertFileExists asserts that a file exists
func AssertFileExists(t *testing.T, path string) {
	t.Helper()

	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("file %s does not exist: %v", path, err)
	}

	if info.IsDir() {
		t.Fatalf("%s is a directory, not a file", path)
	}
}

// AssertFileContains asserts that a file contains specific content
func AssertFileContains(t *testing.T, path, content string) {
	t.Helper()

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read file %s: %v", path, err)
	}

	contentStr := string(data)
	if !strings.Contains(contentStr, content) {
		t.Fatalf("file %s does not contain expected content:\ngot: %s\nwant: %s",
			path, contentStr, content)
	}
}

// SkipIfShort skips test in short mode
func SkipIfShort(t *testing.T) {
	t.Helper()
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}
}
