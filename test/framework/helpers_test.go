package framework

import (
    "os"
    "path/filepath"
    "testing"
)

func TestAssertDirExists(t *testing.T) {
    tempDir, cleanup := SetupTestEnvironment(t)
    defer cleanup()

    // Test existing directory
    SetupTestConfig(t, tempDir)
    AssertDirExists(t, filepath.Join(tempDir, "versions"))

    // Test non-existing directory should fail
    t.Run("non-existent", func(t *testing.T) {
        defer func() {
            if r := recover(); r == nil {
                t.Fatal("expected AssertDirExists to panic for non-existent directory")
            }
        }()
        AssertDirExists(t, filepath.Join(tempDir, "non-existent"))
    })
}

func TestAssertFileExists(t *testing.T) {
    tempDir, cleanup := SetupTestEnvironment(t)
    defer cleanup()

    // Test existing file
    testFile := filepath.Join(tempDir, "test.txt")
    CreateTestFile(t, testFile, "content")
    AssertFileExists(t, testFile)

    // Test non-existing file should fail
    t.Run("non-existent", func(t *testing.T) {
        defer func() {
            if r := recover(); r == nil {
                t.Fatal("expected AssertFileExists to panic for non-existent file")
            }
        }()
        AssertFileExists(t, filepath.Join(tempDir, "non-existent.txt"))
    })
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

    // Test missing content should fail
    t.Run("missing content", func(t *testing.T) {
        defer func() {
            if r := recover(); r == nil {
                t.Fatal("expected AssertFileContains to panic for missing content")
            }
        }()
        AssertFileContains(t, testFile, "missing")
    })
}

func TestSkipIfShort(t *testing.T) {
    // This test will be skipped in short mode
    SkipIfShort(t)
    t.Log("Test would run in normal mode, skipped in short mode")
}