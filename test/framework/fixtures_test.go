package framework

import (
    "path/filepath"
    "testing"
)

func TestCreateTarGzFixture(t *testing.T) {
    tempDir, cleanup := SetupTestEnvironment(t)
    defer cleanup()

    tarFile := filepath.Join(tempDir, "test.tar.gz")

    files := map[string]string{
        "file1.txt": "content 1",
        "file2.txt": "content 2",
    }

    err := CreateTarGzFixture(tarFile, files)
    if err != nil {
        t.Fatalf("failed to create tar.gz fixture: %v", err)
    }

    AssertFileExists(t, tarFile)
}

func TestCreateGzipFixture(t *testing.T) {
    tempDir, cleanup := SetupTestEnvironment(t)
    defer cleanup()

    gzipFile := filepath.Join(tempDir, "test.gz")
    content := "test gzip content"

    err := CreateGzipFixture(gzipFile, content)
    if err != nil {
        t.Fatalf("failed to create gzip fixture: %v", err)
    }

    AssertFileExists(t, gzipFile)
}

func TestNewHTTPTestServer(t *testing.T) {
    server := NewHTTPTestServer()

    if server.BaseURL != "" {
        t.Fatal("BaseURL should be empty initially")
    }

    if server.Responses == nil {
        t.Fatal("Responses should not be nil")
    }

    if server.StatusCode == nil {
        t.Fatal("StatusCode should not be nil")
    }
}