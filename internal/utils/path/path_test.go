package path_test

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"gvm/internal/core"
	gvmpath "gvm/internal/utils/path"
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
	if runtime.GOOS == "windows" {
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
	if runtime.GOOS == "windows" {
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
