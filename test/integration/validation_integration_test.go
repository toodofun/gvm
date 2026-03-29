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

package integration_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/toodofun/gvm/internal/core/validation"
)

// TestValidationLayerIntegration tests the integration of all validation functions
func TestValidationLayerIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	t.Run("VersionValidationIntegration", func(t *testing.T) {
		// Test valid versions
		validVersions := []string{
			"1.0.0",
			"v1.2.3",
			"2.1",
			"v10.20.30",
		}

		for _, version := range validVersions {
			err := validation.ValidateVersion(version)
			assert.NoError(t, err, "Valid version should pass: %s", version)
		}

		// Test invalid versions
		invalidVersions := []string{
			"",
			"invalid",
			"1",
			"abc.def.ghi",
			"1.2.3.4",
		}

		for _, version := range invalidVersions {
			err := validation.ValidateVersion(version)
			assert.Error(t, err, "Invalid version should fail: %s", version)
		}
	})

	t.Run("PathValidationIntegration", func(t *testing.T) {
		// Test safe paths
		safePaths := []string{
			"/usr/local/bin",
			"/home/user/documents",
			"./relative/path",
			"../parent/dir",
		}

		for _, path := range safePaths {
			err := validation.ValidatePath(path)
			// Note: ../parent/dir should fail due to path traversal check
			if path == "../parent/dir" {
				assert.Error(t, err, "Path traversal should fail: %s", path)
			} else {
				assert.NoError(t, err, "Safe path should pass: %s", path)
			}
		}

		// Test dangerous paths
		dangerousPaths := []string{
			"",
			"/../../../etc/passwd",
			"path\x00with\x00nulls",
			"path%2ewith%2eencoding",
			string(make([]byte, 5000)), // Too long
		}

		for _, path := range dangerousPaths {
			err := validation.ValidatePath(path)
			assert.Error(t, err, "Dangerous path should fail: %s", path)
		}
	})

	t.Run("URLValidationIntegration", func(t *testing.T) {
		// Test safe URLs
		safeURLs := []string{
			"https://example.com",
			"https://api.github.com/releases",
			"https://golang.org/dl/go1.21.0.darwin-amd64.tar.gz",
		}

		for _, url := range safeURLs {
			err := validation.ValidateURL(url)
			assert.NoError(t, err, "Safe URL should pass: %s", url)
		}

		// Test dangerous URLs
		dangerousURLs := []string{
			"",
			"http://example.com",              // Not HTTPS
			"ftp://example.com",               // Wrong scheme
			"https://localhost:8080",          // Localhost
			"https://127.0.0.1",               // Loopback
			"https://192.168.1.1",             // Private IP
			"not-a-url",
			"javascript:alert('xss')",
		}

		for _, url := range dangerousURLs {
			err := validation.ValidateURL(url)
			assert.Error(t, err, "Dangerous URL should fail: %s", url)
		}
	})

	t.Run("CombinedValidationScenarios", func(t *testing.T) {
		// Simulate real-world validation scenarios
		scenarios := []struct {
			name     string
			version  string
			path     string
			url      string
			wantErr  bool
		}{
			{
				name:    "Valid Go installation",
				version: "1.21.0",
				path:    "/usr/local/go",
				url:     "https://go.dev/dl/go1.21.0.darwin-arm64.tar.gz",
				wantErr: false,
			},
			{
				name:    "Malicious version download",
				version: "1.0.0",
				path:    "/etc/passwd",
				url:     "https://localhost/malware.tar.gz",
				wantErr: true,
			},
			{
				name:    "SSRF attempt",
				version: "2.0.0",
				path:    "/app/bin",
				url:     "https://192.168.1.100/internal.tar.gz",
				wantErr: true,
			},
		}

		for _, scenario := range scenarios {
			t.Run(scenario.name, func(t *testing.T) {
				hasError := false

				if err := validation.ValidateVersion(scenario.version); err != nil {
					hasError = true
				}

				if err := validation.ValidatePath(scenario.path); err != nil {
					hasError = true
				}

				if err := validation.ValidateURL(scenario.url); err != nil {
					hasError = true
				}

				if scenario.wantErr {
					assert.True(t, hasError, "Scenario should fail validation: %s", scenario.name)
				} else {
					assert.False(t, hasError, "Scenario should pass validation: %s", scenario.name)
				}
			})
		}
	})
}
