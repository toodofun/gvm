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
	"gvm/internal/log"
	"io"
	"os"
	"path/filepath"
	"strings"
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
		absFileName := filepath.Join(absPath, fileName)
		logger.Debugf("%s", absFileName)

		if fInfo.Mode().IsDir() {
			if err := os.MkdirAll(absFileName, fInfo.Mode().Perm()); err != nil {
				return err
			}
			continue
		}
		dir := filepath.Dir(absFileName)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}
		file, err := os.OpenFile(absFileName, os.O_RDWR|os.O_CREATE|os.O_TRUNC, fInfo.Mode().Perm())
		if err != nil {
			return err
		}

		n, err := io.Copy(file, tarStream)
		if e := file.Close(); e != nil {
			return e
		}
		if err != nil {
			return err
		}
		if n != fInfo.Size() {
			return fmt.Errorf("file size mismatch, wrote %d, want %d", n, fInfo.Size())
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
