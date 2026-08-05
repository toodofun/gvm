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
	"archive/tar"
	"compress/gzip"
	"os"
)

// CreateTarGzFixture creates a test tar.gz file
func CreateTarGzFixture(destPath string, files map[string]string) error {
	destFile, err := os.Create(destPath)
	if err != nil {
		return err
	}
	defer destFile.Close()

	gw := gzip.NewWriter(destFile)
	defer gw.Close()

	tw := tar.NewWriter(gw)
	defer tw.Close()

	for name, content := range files {
		header := &tar.Header{
			Name: name,
			Mode: 0644,
			Size: int64(len(content)),
		}

		if err := tw.WriteHeader(header); err != nil {
			return err
		}

		if _, err := tw.Write([]byte(content)); err != nil {
			return err
		}
	}

	return nil
}

// CreateGzipFixture creates a test gzip file
func CreateGzipFixture(destPath, content string) error {
	file, err := os.Create(destPath)
	if err != nil {
		return err
	}
	defer file.Close()

	gw := gzip.NewWriter(file)
	defer gw.Close()

	_, err = gw.Write([]byte(content))
	return err
}

// HTTPTestServer HTTP test server configuration
type HTTPTestServer struct {
	BaseURL    string
	Responses  map[string]string
	StatusCode map[string]int
}

// NewHTTPTestServer creates a new test HTTP server configuration
// Note: Actual test servers should use httptest.NewServer in tests
func NewHTTPTestServer() *HTTPTestServer {
	return &HTTPTestServer{
		Responses:  make(map[string]string),
		StatusCode: make(map[string]int),
	}
}
