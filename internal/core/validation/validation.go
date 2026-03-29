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

	cleaned := filepath.Clean(path)
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
