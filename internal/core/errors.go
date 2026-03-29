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
