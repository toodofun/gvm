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
	"fmt"
	"net"
	"net/url"
	"path/filepath"
	"regexp"
	"strings"
)

const (
	MaxVersionLength = 64
	MaxPathLength    = 4096
	MaxURLLength     = 2048
)

var versionRegex = regexp.MustCompile(`^v?[0-9]+\.[0-9]+(\.[0-9]+)?$`)

// ValidateVersion validates that a version string follows semantic versioning format
func ValidateVersion(version string) error {
	if version == "" {
		return fmt.Errorf("version cannot be empty")
	}
	if len(version) > MaxVersionLength {
		return fmt.Errorf("version too long (max %d characters)", MaxVersionLength)
	}
	version = strings.TrimSpace(version)
	if !versionRegex.MatchString(version) {
		return fmt.Errorf("invalid version format: %s (expected format: 1.2.3 or 1.2)", version)
	}
	return nil
}

// ValidatePath validates that a path is safe and doesn't contain path traversal attempts
func ValidatePath(path string) error {
	if path == "" {
		return fmt.Errorf("path cannot be empty")
	}
	if len(path) > MaxPathLength {
		return fmt.Errorf("path too long (max %d characters)", MaxPathLength)
	}
	path = strings.TrimSpace(path)

	// Check for URL encoding attempts
	if strings.Contains(path, "%") {
		return fmt.Errorf("path contains URL encoding")
	}

	if strings.Contains(path, "\x00") {
		return fmt.Errorf("path contains null byte")
	}

	// Check for path traversal attempts
	// First, clean the path to normalize it and resolve ".." components
	cleaned := filepath.Clean(path)

	// If ".." still exists after cleaning, it indicates a real traversal attempt
	// This is more accurate than checking the original path, as it allows
	// valid relative paths like "foo/../bar" while blocking real traversal
	if strings.Contains(cleaned, "..") {
		return fmt.Errorf("path traversal detected")
	}
	return nil
}

// ValidateURL validates that a URL is properly formatted and uses HTTPS
func ValidateURL(urlStr string) error {
	if urlStr == "" {
		return fmt.Errorf("URL cannot be empty")
	}
	if len(urlStr) > MaxURLLength {
		return fmt.Errorf("URL too long (max %d characters)", MaxURLLength)
	}
	urlStr = strings.TrimSpace(urlStr)

	parsed, err := url.Parse(urlStr)
	if err != nil {
		return fmt.Errorf("invalid URL: %w", err)
	}

	// Check for missing or empty scheme before checking host
	if parsed.Scheme == "" {
		return fmt.Errorf("invalid URL: missing scheme")
	}

	if parsed.Scheme != "https" {
		return fmt.Errorf("URL must use HTTPS: %s", urlStr)
	}

	if parsed.Host == "" {
		return fmt.Errorf("URL missing host")
	}

	// SSRF protection - block localhost and private IPs
	host := parsed.Hostname()
	if host == "localhost" || host == "127.0.0.1" || host == "::1" {
		return fmt.Errorf("URL cannot point to localhost")
	}

	ip := net.ParseIP(host)
	if ip != nil && (ip.IsLoopback() || ip.IsPrivate() || ip.IsLinkLocalUnicast()) {
		return fmt.Errorf("URL cannot point to private IP: %s", host)
	}

	return nil
}
