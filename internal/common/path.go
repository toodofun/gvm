package common

import (
	"fmt"
	"gvm/core"
	"os"
	"path"
)

const (
	Current = "Current"
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
	info, err := os.Lstat(target)
	if err == nil && info.Mode()&os.ModeSymlink != 0 {
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
