// Copyright 2025 The Toodofun Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package framework

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSetupTestEnvironment(t *testing.T) {
	tempDir, cleanup := SetupTestEnvironment(t)
	defer cleanup()

	if tempDir == "" {
		t.Fatal("temp dir should not be empty")
	}

	// Verify directory exists
	info, err := os.Stat(tempDir)
	if err != nil {
		t.Fatalf("temp dir should exist: %v", err)
	}

	if !info.IsDir() {
		t.Fatal("temp path should be a directory")
	}
}

func TestSetupTestConfig(t *testing.T) {
	tempDir, cleanup := SetupTestEnvironment(t)
	defer cleanup()

	SetupTestConfig(t, tempDir)

	// Verify directories were created
	AssertDirExists(t, filepath.Join(tempDir, "versions"))
	AssertDirExists(t, filepath.Join(tempDir, "archives"))
	AssertDirExists(t, filepath.Join(tempDir, "config"))
}

func TestCreateTestFile(t *testing.T) {
	tempDir, cleanup := SetupTestEnvironment(t)
	defer cleanup()

	SetupTestConfig(t, tempDir)

	testFile := filepath.Join(tempDir, "test", "file.txt")
	content := "test content"

	CreateTestFile(t, testFile, content)

	AssertFileExists(t, testFile)
	AssertFileContains(t, testFile, content)
}
