// Copyright 2025 The Toodofun Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http:www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package compress

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"context"
	"os"
	"path/filepath"
	"testing"
)

func createTestTarGz(t *testing.T, filename string, files map[string]string) {
	file, err := os.Create(filename)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()

	gw := gzip.NewWriter(file)
	tw := tar.NewWriter(gw)

	for name, content := range files {
		hdr := &tar.Header{
			Name: name,
			Mode: 0600,
			Size: int64(len(content)),
		}
		if err := tw.WriteHeader(hdr); err != nil {
			t.Fatal(err)
		}
		if _, err := tw.Write([]byte(content)); err != nil {
			t.Fatal(err)
		}
	}

	if err := tw.Close(); err != nil {
		t.Fatal(err)
	}
	if err := gw.Close(); err != nil {
		t.Fatal(err)
	}
}

func createTestZip(t *testing.T, filename string, files map[string]string) {
	file, err := os.Create(filename)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()

	zw := zip.NewWriter(file)
	for name, content := range files {
		f, err := zw.Create(name)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := f.Write([]byte(content)); err != nil {
			t.Fatal(err)
		}
	}
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestUnTarGz(t *testing.T) {
	tmpDir := t.TempDir()
	tarGzPath := filepath.Join(tmpDir, "test.tar.gz")
	files := map[string]string{
		"001/foo.txt": "hello tar",
	}

	createTestTarGz(t, tarGzPath, files)

	dest := filepath.Join(tmpDir, "untarred")
	err := UnTarGz(context.Background(), tarGzPath, dest)
	if err != nil {
		t.Fatalf("UnTarGz failed: %v", err)
	}

	for name, content := range files {
		data, err := os.ReadFile(filepath.Join(dest, name)) // 支持目录
		if err != nil {
			t.Errorf("failed to read file: %v", err)
		}
		if string(data) != content {
			t.Errorf("expected %q, got %q", content, string(data))
		}
	}
}

func TestUnZip(t *testing.T) {
	tmpDir := t.TempDir()
	zipPath := filepath.Join(tmpDir, "test.zip")
	files := map[string]string{
		"bar.txt": "hello zip",
	}

	createTestZip(t, zipPath, files)

	dest := filepath.Join(tmpDir, "unzipped")
	err := UnZip(context.Background(), zipPath, dest)
	if err != nil {
		t.Fatalf("UnZip failed: %v", err)
	}

	for name, content := range files {
		data, err := os.ReadFile(filepath.Join(dest, name))
		if err != nil {
			t.Errorf("failed to read file: %v", err)
		}
		if string(data) != content {
			t.Errorf("expected %q, got %q", content, string(data))
		}
	}
}
