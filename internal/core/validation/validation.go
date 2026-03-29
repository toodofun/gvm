package validation

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"
)

var versionRegex = regexp.MustCompile(`^v?[0-9]+\.[0-9]+(\.[0-9]+)?$`)

// ValidateVersion validates that a version string follows semantic versioning format
func ValidateVersion(version string) error {
	if version == "" {
		return fmt.Errorf("version cannot be empty")
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
	path = strings.TrimSpace(path)
	if strings.Contains(path, "\x00") {
		return fmt.Errorf("path contains null byte")
	}
	// Check for path traversal before cleaning
	if strings.Contains(path, "..") {
		return fmt.Errorf("path traversal detected: %s", path)
	}
	return nil
}

// ValidateURL validates that a URL is properly formatted and uses HTTPS
func ValidateURL(urlStr string) error {
	if urlStr == "" {
		return fmt.Errorf("URL cannot be empty")
	}
	urlStr = strings.TrimSpace(urlStr)
	// Check for spaces before parsing
	if strings.Contains(urlStr, " ") {
		return fmt.Errorf("URL contains spaces: %s", urlStr)
	}
	parsed, err := url.Parse(urlStr)
	if err != nil {
		return fmt.Errorf("invalid URL: %w", err)
	}
	if parsed.Scheme != "https" {
		return fmt.Errorf("URL must use HTTPS: %s", urlStr)
	}
	if parsed.Host == "" {
		return fmt.Errorf("URL missing host: %s", urlStr)
	}
	return nil
}
