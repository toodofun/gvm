// Copyright 2026 The Toodofun Authors
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

package validation

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateVersion(t *testing.T) {
	tests := []struct {
		name    string
		version string
		wantErr bool
		errMsg  string
	}{
		{name: "valid version with patch", version: "1.20.0", wantErr: false},
		{name: "valid version without patch", version: "1.20", wantErr: false},
		{name: "valid version with v prefix", version: "v1.20.0", wantErr: false},
		{name: "empty version", version: "", wantErr: true, errMsg: "cannot be empty"},
		{name: "invalid format", version: "invalid", wantErr: true, errMsg: "invalid version format"},
		{name: "negative number", version: "1.-1.0", wantErr: true, errMsg: "invalid version format"},
		{name: "too many parts", version: "1.2.3.4", wantErr: true, errMsg: "invalid version format"},
		{name: "version too long", version: strings.Repeat("1", 65), wantErr: true, errMsg: "too long"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateVersion(tt.version)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidatePath(t *testing.T) {
	tests := []struct {
		name    string
		path    string
		wantErr bool
		errMsg  string
	}{
		{name: "valid absolute path", path: "/usr/local/go", wantErr: false},
		{name: "valid relative path", path: "./go", wantErr: false},
		{name: "valid home path", path: "~/go", wantErr: false},
		{name: "empty path", path: "", wantErr: true, errMsg: "cannot be empty"},
		{name: "valid path with .. that normalizes", path: "/etc/../passwd", wantErr: false},
		{name: "path traversal with ..", path: "../../../etc/passwd", wantErr: true, errMsg: "path traversal"},
		{name: "null bytes", path: "/etc/passwd\x00", wantErr: true, errMsg: "null byte"},
		{name: "path too long", path: strings.Repeat("a", 4097), wantErr: true, errMsg: "too long"},
		{name: "URL encoded path traversal", path: "/etc%2f../passwd", wantErr: true, errMsg: "URL encoding"},
		{name: "URL encoded dot", path: "/etc/%2e%2e/passwd", wantErr: true, errMsg: "URL encoding"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePath(tt.path)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateURL(t *testing.T) {
	tests := []struct {
		name    string
		url     string
		wantErr bool
		errMsg  string
	}{
		{name: "valid HTTPS URL", url: "https://go.dev/dl/go1.20.0.linux-amd64.tar.gz", wantErr: false},
		{name: "valid HTTPS URL with query", url: "https://example.com/download?version=1.20", wantErr: false},
		{name: "empty URL", url: "", wantErr: true, errMsg: "cannot be empty"},
		{name: "HTTP not HTTPS", url: "http://example.com/file.tar.gz", wantErr: true, errMsg: "must use HTTPS"},
		{name: "invalid URL format", url: "not-a-url", wantErr: true, errMsg: "invalid URL"},
		{
			name:    "URL too long",
			url:     "https://example.com/" + strings.Repeat("a", 2048),
			wantErr: true,
			errMsg:  "too long",
		},
		{name: "localhost URL", url: "https://localhost/file.tar.gz", wantErr: true, errMsg: "localhost"},
		{name: "127.0.0.1 URL", url: "https://127.0.0.1/file.tar.gz", wantErr: true, errMsg: "localhost"},
		{name: "private IP URL", url: "https://192.168.1.1/file.tar.gz", wantErr: true, errMsg: "private IP"},
		{name: "private IP URL 10.0.0.1", url: "https://10.0.0.1/file.tar.gz", wantErr: true, errMsg: "private IP"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateURL(tt.url)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
