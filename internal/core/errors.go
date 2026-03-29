package core

import "errors"

// Standard error code definitions
var (
	// Validation errors
	ErrInvalidVersion = errors.New("invalid version format")
	ErrInvalidPath    = errors.New("invalid path")
	ErrInvalidURL     = errors.New("invalid URL")

	// Download and installation errors
	ErrDownloadFailed = errors.New("download failed")
	ErrExtractFailed  = errors.New("extraction failed")

	// Security errors
	ErrCommandBlocked = errors.New("command not allowed")

	// Language and version errors
	ErrLanguageNotFound = errors.New("language not found")
	ErrVersionNotFound  = errors.New("version not found")
	ErrAlreadyInstalled = errors.New("version already installed")
)
