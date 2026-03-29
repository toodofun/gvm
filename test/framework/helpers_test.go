package framework

import (
    "path/filepath"
    "testing"
)

func TestAssertDirExists(t *testing.T) {
    tempDir, cleanup := SetupTestEnvironment(t)
    defer cleanup()

    // Test existing directory
    SetupTestConfig(t, tempDir)
    AssertDirExists(t, filepath.Join(tempDir, "versions"))
}

func TestAssertFileExists(t *testing.T) {
    tempDir, cleanup := SetupTestEnvironment(t)
    defer cleanup()

    // Test existing file
    testFile := filepath.Join(tempDir, "test.txt")
    CreateTestFile(t, testFile, "content")
    AssertFileExists(t, testFile)
}

func TestAssertFileContains(t *testing.T) {
    tempDir, cleanup := SetupTestEnvironment(t)
    defer cleanup()

    testFile := filepath.Join(tempDir, "test.txt")
    content := "hello world"

    CreateTestFile(t, testFile, content)

    // Test contains existing content
    AssertFileContains(t, testFile, "hello")
    AssertFileContains(t, testFile, "world")
}

func TestSkipIfShort(t *testing.T) {
    // This test will be skipped in short mode
    SkipIfShort(t)
    t.Log("Test would run in normal mode, skipped in short mode")
}
