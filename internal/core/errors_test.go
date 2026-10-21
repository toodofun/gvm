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

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestErrorCodes_AreDefined(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want string
	}{
		{name: "ErrInvalidVersion", err: ErrInvalidVersion, want: "invalid version format"},
		{name: "ErrInvalidPath", err: ErrInvalidPath, want: "invalid path"},
		{name: "ErrInvalidURL", err: ErrInvalidURL, want: "invalid URL"},
		{name: "ErrDownloadFailed", err: ErrDownloadFailed, want: "download failed"},
		{name: "ErrExtractFailed", err: ErrExtractFailed, want: "extraction failed"},
		{name: "ErrCommandBlocked", err: ErrCommandBlocked, want: "command not allowed"},
		{name: "ErrLanguageNotFound", err: ErrLanguageNotFound, want: "language not found"},
		{name: "ErrVersionNotFound", err: ErrVersionNotFound, want: "version not found"},
		{name: "ErrAlreadyInstalled", err: ErrAlreadyInstalled, want: "version already installed"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.NotNil(t, tt.err)
			assert.Contains(t, tt.err.Error(), tt.want)
		})
	}
}

func TestErrorIs(t *testing.T) {
	baseErr := ErrInvalidVersion
	wrappedErr := fmt.Errorf("wrap: %w", baseErr)

	assert.True(t, errors.Is(wrappedErr, ErrInvalidVersion))
}
