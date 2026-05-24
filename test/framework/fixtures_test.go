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
