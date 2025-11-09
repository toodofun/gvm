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
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/toodofun/gvm/internal/log"
)

func UnTarGz(ctx context.Context, tarGzName string, dest string) error {
	logger := log.GetLogger(ctx)
	gzReader, err := os.Open(tarGzName)
	defer func(gzReader *os.File) {
		err := gzReader.Close()
		if err != nil {
			logger.Warnf("Close body error: %s", err)
		}
	}(gzReader)

	if err != nil {
		return err
	}

	unGzStream, err := gzip.NewReader(gzReader)
	if err != nil {
		return err
	}

	tarStream := tar.NewReader(unGzStream)
	absPath, err := filepath.Abs(dest)
	if err != nil {
		return err
	}

	for {
		hdr, err := tarStream.Next()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return err
		}
		fInfo := hdr.FileInfo()
		fileName := hdr.Name
		cleanName := filepath.Clean(fileName)
		if strings.HasPrefix(cleanName, "..") || strings.Contains(cleanName, "../") {
			return fmt.Errorf("invalid archive path: %s", cleanName)
		}
		absFileName := filepath.Join(absPath, fileName)
		if !strings.HasPrefix(absFileName, filepath.Clean(absPath)+string(os.PathSeparator)) {
			return fmt.Errorf("illegal file path: %s", absFileName)
		}
		logger.Debugf("%s", absFileName)

		if fInfo.Mode().IsDir() {
			if err := os.MkdirAll(absFileName, fInfo.Mode().Perm()); err != nil {
				return err
			}
			continue
		}
		dir := filepath.Dir(absFileName)
		if err = os.MkdirAll(dir, 0755); err != nil {
			return err
		}
		switch hdr.Typeflag {
		case tar.TypeDir:
			if err = os.MkdirAll(absFileName, os.FileMode(hdr.Mode)); err != nil {
				return err
			}
		case tar.TypeReg:
			file, err := os.OpenFile(absFileName, os.O_RDWR|os.O_CREATE|os.O_TRUNC, fInfo.Mode().Perm())
			if err != nil {
				return err
			}
			n, err := io.Copy(file, tarStream)
			if closeErr := file.Close(); closeErr != nil {
				return closeErr
			}
			if err != nil {
				return err
			}
			if n != fInfo.Size() {
				return fmt.Errorf("file size mismatch, wrote %d, want %d", n, fInfo.Size())
			}
		case tar.TypeSymlink:
			if err = os.Symlink(hdr.Linkname, absFileName); err != nil {
				return fmt.Errorf("failed to create symlink %s -> %s: %w", absFileName, hdr.Linkname, err)
			}
		default:
			logger.Warnf("Unsupported tar entry type: %v (%s)", hdr.Typeflag, hdr.Name)
		}
	}
	return nil
}

func UnZip(ctx context.Context, zipFile string, dest string) error {
	logger := log.GetLogger(ctx)
	reader, err := zip.OpenReader(zipFile)
	if err != nil {
		return err
	}
	defer func() {
		if e := reader.Close(); e != nil {
			logger.Warnf("Close zip reader error: %s", e)
		}
	}()

	absPath, err := filepath.Abs(dest)
	if err != nil {
		return err
	}

	for _, f := range reader.File {
		fpath := filepath.Join(absPath, f.Name)
		logger.Debugf("%s", fpath)

		// 防止 Zip Slip 漏洞
		if !strings.HasPrefix(fpath, absPath+string(os.PathSeparator)) {
			return fmt.Errorf("illegal file path: %s", fpath)
		}

		if f.FileInfo().IsDir() {
			if err := os.MkdirAll(fpath, f.Mode()); err != nil {
				return err
			}
			continue
		}

		if err := os.MkdirAll(filepath.Dir(fpath), os.ModePerm); err != nil {
			return err
		}

		dstFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return err
		}

		rc, err := f.Open()
		if err != nil {
			_ = dstFile.Close()
			return err
		}

		n, err := io.Copy(dstFile, rc)
		_ = rc.Close()
		if e := dstFile.Close(); e != nil {
			return e
		}
		if err != nil {
			return err
		}
		if n != f.FileInfo().Size() {
			return fmt.Errorf("file size mismatch, wrote %d, want %d", n, f.FileInfo().Size())
		}
	}
	return nil
}

func UnPkg(pkgPath, destDir string) error {
	if _, err := exec.LookPath("xar"); err != nil {
		return errors.New("xar command not found, please install it (macOS only)")
	}
	if _, err := exec.LookPath("pax"); err != nil {
		return errors.New("pax command not found, please ensure it's available (macOS only)")
	}

	absDest, err := filepath.Abs(destDir)
	if err != nil {
		return fmt.Errorf("failed to resolve dest path: %w", err)
	}
	if err := os.MkdirAll(absDest, 0755); err != nil {
		return fmt.Errorf("failed to create dest dir: %w", err)
	}

	tmpDir, err := os.MkdirTemp("", "pkg-extract-*")
	if err != nil {
		return fmt.Errorf("failed to create temp dir: %w", err)
	}
	defer func() {
		_ = os.RemoveAll(tmpDir)
	}()

	cmd1 := exec.Command("xar", "-xf", pkgPath)
	cmd1.Dir = tmpDir
	cmd1.Stdout = os.Stdout
	cmd1.Stderr = os.Stderr
	if err := cmd1.Run(); err != nil {
		return fmt.Errorf("xar extraction failed: %w", err)
	}

	payloadPath := filepath.Join(tmpDir, "Payload")
	if _, err := os.Stat(payloadPath); os.IsNotExist(err) {
		payloadPath = filepath.Join(tmpDir, "Content")
		if _, err := os.Stat(payloadPath); err != nil {
			return errors.New("no Payload or Content found in pkg")
		}
	}
	cmd2 := exec.Command("pax", "-rzf", payloadPath)
	cmd2.Dir = absDest
	cmd2.Stdout = os.Stdout
	cmd2.Stderr = os.Stderr
	if err := cmd2.Run(); err != nil {
		return fmt.Errorf("pax unpack failed: %w", err)
	}
	return nil
}

// UnTarXz 解压 tar.xz 文件
func UnTarXz(ctx context.Context, tarXzName string, dest string) error {
	logger := log.GetLogger(ctx)

	// 确保目标目录存在
	if err := os.MkdirAll(dest, 0755); err != nil {
		return fmt.Errorf("failed to create destination directory: %w", err)
	}

	// 在不同平台上使用不同的命令
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		// Windows 上可能需要使用 7zip 或其他工具
		// 这里假设有 tar 命令可用（Git Bash 或 WSL）
		cmd = exec.CommandContext(ctx, "tar", "-xJf", tarXzName, "-C", dest)
	} else {
		// Unix-like 系统（Linux, macOS）
		cmd = exec.CommandContext(ctx, "tar", "-xJf", tarXzName, "-C", dest)
	}

	cmd.Stdout = log.GetStdout(ctx)
	cmd.Stderr = log.GetStderr(ctx)

	logger.Infof("Extracting %s to %s", tarXzName, dest)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to extract tar.xz: %w", err)
	}

	return nil
}
