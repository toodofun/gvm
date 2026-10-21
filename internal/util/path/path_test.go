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

package path_test

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/toodofun/gvm/internal/core"
	gvmpath "github.com/toodofun/gvm/internal/util/path"
)

const (
	windowsOS = "windows"
)

// mock root dir for testing
func setupTestRoot(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	core.GetRootDir = func() string {
		return root
	}
	return root
}

func TestGetLangRoot(t *testing.T) {
	dir := t.TempDir()
	original := core.GetRootDir()
	defer func() {
		core.GetRootDir = func() string {
			return original
		}
	}()

	core.GetRootDir = func() string {
		return dir
	}
	got := gvmpath.GetLangRoot("go")
	expected := filepath.Join(dir, "go")

	if got != expected {
		t.Errorf("expected %s, got %s", expected, got)
	}
}

func TestGetInstalledVersion_Empty(t *testing.T) {
	_ = setupTestRoot(t)
	versions, err := gvmpath.GetInstalledVersion("go", "bin/go")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(versions) != 0 {
		t.Errorf("expected no versions, got: %v", versions)
	}
}

func TestGetInstalledVersion_IgnoreNonDirAndCurrent(t *testing.T) {
	root := setupTestRoot(t)
	langDir := filepath.Join(root, "go")
	os.MkdirAll(langDir, 0755)
	_ = os.WriteFile(filepath.Join(langDir, "notadir.txt"), []byte("x"), 0644)
	_ = os.Mkdir(filepath.Join(langDir, "Current"), 0755)

	versions, err := gvmpath.GetInstalledVersion("go", "bin/go")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(versions) != 0 {
		t.Errorf("expected no valid versions, got: %v", versions)
	}
}

func TestGetInstalledVersion_BinExists(t *testing.T) {
	root := setupTestRoot(t)
	dir := filepath.Join(root, "go", "1.18.0", "bin")
	os.MkdirAll(dir, 0755)
	os.WriteFile(filepath.Join(dir, "go"), []byte("#!/bin/bash"), 0755)

	versions, err := gvmpath.GetInstalledVersion("go", "bin/go")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(versions) != 1 || versions[0] != "1.18.0" {
		t.Errorf("expected [1.18.0], got: %v", versions)
	}
}

func TestSetSymlink_CreateAndOverwrite(t *testing.T) {
	if runtime.GOOS == windowsOS {
		t.Skip("symlink test skipped on Windows")
	}
	dir := t.TempDir()
	source := filepath.Join(dir, "source")
	target := filepath.Join(dir, "target")

	_ = os.WriteFile(source, []byte("data"), 0644)
	_ = os.Symlink(source, target)

	err := gvmpath.SetSymlink(source, target)
	if err != nil {
		t.Fatalf("expected overwrite symlink success, got error: %v", err)
	}

	link, err := os.Readlink(target)
	if err != nil || link != source {
		t.Errorf("expected target symlink to %s, got %s (err: %v)", source, link, err)
	}
}

func TestSetSymlink_TargetExistsButNotSymlink(t *testing.T) {
	if runtime.GOOS == windowsOS {
		t.Skip("symlink test skipped on Windows")
	}
	dir := t.TempDir()
	source := filepath.Join(dir, "source")
	target := filepath.Join(dir, "target")

	_ = os.WriteFile(source, []byte("data"), 0644)
	_ = os.WriteFile(target, []byte("not a symlink"), 0644) // regular file

	err := gvmpath.SetSymlink(source, target)
	if err != nil {
		t.Fatalf("expected to create symlink (overwrite file), got error: %v", err)
	}

	_, err = os.Readlink(target)
	if err != nil {
		t.Errorf("expected symlink, got error: %v", err)
	}
}

func TestIsPathExist(t *testing.T) {
	existing := t.TempDir()
	if !gvmpath.IsPathExist(existing) {
		t.Errorf("expected path to exist: %s", existing)
	}

	nonexistent := filepath.Join(existing, "nope")
	if gvmpath.IsPathExist(nonexistent) {
		t.Errorf("expected path to not exist: %s", nonexistent)
	}
}

func TestIsPathSafe(t *testing.T) {
	tests := []struct {
		name     string
		basePath string
		target   string
		safe     bool
		wantErr  bool
	}{
		{
			name:     "normal subdirectory",
			basePath: "/tmp/gvm",
			target:   "subdir/file.txt",
			safe:     true,
			wantErr:  false,
		},
		{
			name:     "path traversal with ..",
			basePath: "/tmp/gvm",
			target:   "../etc/passwd",
			safe:     false,
			wantErr:  true,
		},
		{
			name:     "deep path traversal",
			basePath: "/tmp/gvm",
			target:   "subdir/../../etc/passwd",
			safe:     false,
			wantErr:  true,
		},
		{
			name:     "absolute path escape",
			basePath: "/tmp/gvm",
			target:   "/etc/passwd",
			safe:     false,
			wantErr:  true,
		},
		{
			name:     "symlink-like path",
			basePath: "/tmp/gvm",
			target:   "subdir/../../../etc/passwd",
			safe:     false,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			safe, err := gvmpath.IsPathSafe(tt.basePath, tt.target)

			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error but got none")
				}
				if safe {
					t.Errorf("expected path to be unsafe, got safe=true")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if safe != tt.safe {
					t.Errorf("expected safe=%v, got safe=%v", tt.safe, safe)
				}
			}
		})
	}
}

func TestCheckSymlinkSafety(t *testing.T) {
	if runtime.GOOS == windowsOS {
		t.Skip("symlink test skipped on Windows")
	}

	tests := []struct {
		name        string
		setup       func(baseDir string) (linkPath string)
		wantErr     bool
		errContains string
	}{
		{
			name: "safe symlink within base",
			setup: func(baseDir string) string {
				target := filepath.Join(baseDir, "safe-target.txt")
				link := filepath.Join(baseDir, "safe-link")
				_ = os.WriteFile(target, []byte("safe"), 0644)
				_ = os.Symlink(target, link)
				return link
			},
			wantErr: false,
		},
		{
			name: "symlink to parent directory escape",
			setup: func(baseDir string) string {
				parentDir := filepath.Dir(baseDir)
				target := filepath.Join(parentDir, "escape.txt")
				link := filepath.Join(baseDir, "escape-link")
				_ = os.WriteFile(target, []byte("escaped"), 0644)
				_ = os.Symlink(target, link)
				return link
			},
			wantErr:     true,
			errContains: "symlink target escapes base path",
		},
		{
			name: "relative symlink with .. escape",
			setup: func(baseDir string) string {
				link := filepath.Join(baseDir, "relative-escape-link")
				_ = os.Symlink("../outside.txt", link)
				return link
			},
			wantErr:     true,
			errContains: "symlink target escapes base path",
		},
		{
			name: "absolute symlink to system path",
			setup: func(baseDir string) string {
				link := filepath.Join(baseDir, "absolute-escape-link")
				_ = os.Symlink("/etc/passwd", link)
				return link
			},
			wantErr:     true,
			errContains: "symlink target escapes base path",
		},
		{
			name: "safe relative symlink within base",
			setup: func(baseDir string) string {
				target := filepath.Join(baseDir, "subdir", "file.txt")
				link := filepath.Join(baseDir, "safe-relative-link")
				_ = os.MkdirAll(filepath.Join(baseDir, "subdir"), 0755)
				_ = os.WriteFile(target, []byte("safe"), 0644)
				_ = os.Symlink("subdir/file.txt", link)
				return link
			},
			wantErr: false,
		},
		{
			name: "complex relative symlink escape",
			setup: func(baseDir string) string {
				link := filepath.Join(baseDir, "complex-escape-link")
				_ = os.Symlink("subdir/../../../etc/passwd", link)
				return link
			},
			wantErr:     true,
			errContains: "symlink target escapes base path",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			baseDir := t.TempDir()
			linkPath := tt.setup(baseDir)

			err := gvmpath.CheckSymlinkSafety(baseDir, linkPath)

			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error containing %q, got nil", tt.errContains)
				} else if tt.errContains != "" && !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("expected error containing %q, got %q", tt.errContains, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}
		})
	}
}

func TestIsPathSafe_SymlinkWithinBase(t *testing.T) {
	if runtime.GOOS == windowsOS {
		t.Skip("symlink test skipped on Windows")
	}

	baseDir := t.TempDir()
	targetFile := filepath.Join(baseDir, "target.txt")
	linkPath := filepath.Join(baseDir, "link")

	_ = os.WriteFile(targetFile, []byte("content"), 0644)
	_ = os.Symlink(targetFile, linkPath)

	safe, err := gvmpath.IsPathSafe(baseDir, "link")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !safe {
		t.Errorf("expected symlink within base to be safe")
	}
}

func TestIsPathSafe_SymlinkEscape(t *testing.T) {
	if runtime.GOOS == windowsOS {
		t.Skip("symlink test skipped on Windows")
	}

	baseDir := t.TempDir()
	parentDir := filepath.Dir(baseDir)
	escapeTarget := filepath.Join(parentDir, "escape.txt")
	linkPath := filepath.Join(baseDir, "escape-link")

	_ = os.WriteFile(escapeTarget, []byte("escaped"), 0644)
	_ = os.Symlink(escapeTarget, linkPath)

	safe, err := gvmpath.IsPathSafe(baseDir, "escape-link")
	if err == nil {
		t.Errorf("expected error for escaping symlink, got nil")
	}
	if safe {
		t.Errorf("expected escaping symlink to be unsafe")
	}
}
