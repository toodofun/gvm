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

package path

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/toodofun/gvm/internal/core"
)

const (
	Current = "current"
)

func GetLangRoot(lang string) string {
	return path.Join(core.GetRootDir(), lang)
}

func GetInstalledVersion(lang, binPath string) ([]string, error) {
	installedDir := path.Join(core.GetRootDir(), lang)

	_, err := os.Stat(installedDir)
	if os.IsNotExist(err) {
		return make([]string, 0), nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to stat directory %s: %w", installedDir, err)
	}

	entries, err := os.ReadDir(installedDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read dir %s: %w", installedDir, err)
	}

	versions := make([]string, 0)

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		if entry.Name() == Current {
			continue
		}

		_, err = os.Stat(path.Join(installedDir, entry.Name(), binPath))
		if err == nil {
			versions = append(versions, entry.Name())
		}
	}
	return versions, nil
}

func SetSymlink(source, target string) error {
	_, err := os.Lstat(target)
	if err == nil {
		if err := os.Remove(target); err != nil {
			return fmt.Errorf("failed to remove symlink %s: %w", target, err)
		}
	}
	return os.Symlink(source, target)
}

func IsPathExist(dir string) bool {
	_, err := os.Stat(dir)
	if os.IsNotExist(err) {
		return false
	}
	if err != nil {
		return false
	}
	return true
}

// IsPathSafe checks if target path is within base path, preventing path traversal attacks
func IsPathSafe(basePath, targetPath string) (bool, error) {
	// Resolve to absolute paths
	absBase, err := filepath.Abs(basePath)
	if err != nil {
		return false, fmt.Errorf("failed to resolve base path: %w", err)
	}

	// Build full target path
	fullTarget := filepath.Join(basePath, targetPath)
	absTarget, err := filepath.Abs(fullTarget)
	if err != nil {
		return false, fmt.Errorf("failed to resolve target path: %w", err)
	}

	// Check if target path is within base path
	rel, err := filepath.Rel(absBase, absTarget)
	if err != nil {
		return false, fmt.Errorf("failed to compute relative path: %w", err)
	}

	// If relative path starts with .., it tries to escape
	if strings.HasPrefix(rel, "..") {
		return false, fmt.Errorf("path traversal attempt detected: %s tries to escape %s", targetPath, basePath)
	}

	return true, nil
}
