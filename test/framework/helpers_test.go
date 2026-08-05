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
