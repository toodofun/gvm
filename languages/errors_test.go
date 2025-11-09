// Copyright 2025 The Toodofun Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package languages

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPreReleaseError_Error(t *testing.T) {
	tests := []struct {
		name     string
		err      *PreReleaseError
		expected string
	}{
		{
			name: "with available versions",
			err: &PreReleaseError{
				Language:          "python",
				RequestedVersion:  "3.14.0",
				AvailableVersions: []string{"3.14.0a1", "3.14.0rc1", "3.14.0rc2"},
			},
			expected: "版本 3.14.0 尚未正式发布。可用的候选版本：3.14.0a1, 3.14.0rc1, 3.14.0rc2\n请使用完整版本号安装，例如：gvm install python 3.14.0rc2",
		},
		{
			name: "without available versions",
			err: &PreReleaseError{
				Language:          "python",
				RequestedVersion:  "3.14.0",
				AvailableVersions: []string{},
			},
			expected: "版本 3.14.0 尚未正式发布",
		},
		{
			name: "nil available versions",
			err: &PreReleaseError{
				Language:          "python",
				RequestedVersion:  "3.14.0",
				AvailableVersions: nil,
			},
			expected: "版本 3.14.0 尚未正式发布",
		},
		{
			name: "single available version",
			err: &PreReleaseError{
				Language:          "node",
				RequestedVersion:  "21.0.0",
				AvailableVersions: []string{"21.0.0-rc1"},
			},
			expected: "版本 21.0.0 尚未正式发布。可用的候选版本：21.0.0-rc1\n请使用完整版本号安装，例如：gvm install node 21.0.0-rc1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.err.Error()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestPreReleaseError_GetRecommendedVersion(t *testing.T) {
	tests := []struct {
		name     string
		err      *PreReleaseError
		expected string
	}{
		{
			name: "with multiple versions",
			err: &PreReleaseError{
				AvailableVersions: []string{"3.14.0a1", "3.14.0rc1", "3.14.0rc2"},
			},
			expected: "3.14.0rc2",
		},
		{
			name: "with single version",
			err: &PreReleaseError{
				AvailableVersions: []string{"3.14.0rc1"},
			},
			expected: "3.14.0rc1",
		},
		{
			name: "with empty versions",
			err: &PreReleaseError{
				AvailableVersions: []string{},
			},
			expected: "",
		},
		{
			name: "with nil versions",
			err: &PreReleaseError{
				AvailableVersions: nil,
			},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.err.GetRecommendedVersion()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestPreReleaseError_AsError(t *testing.T) {
	originalErr := &PreReleaseError{
		Language:          "python",
		RequestedVersion:  "3.14.0",
		AvailableVersions: []string{"3.14.0rc1", "3.14.0rc2"},
	}

	// Test that it can be unwrapped using errors.As
	var preReleaseErr *PreReleaseError
	assert.True(t, errors.As(originalErr, &preReleaseErr))
	assert.Equal(t, originalErr, preReleaseErr)

	// Test with wrapped error
	wrappedErr := errors.New("installation failed: " + originalErr.Error())
	assert.False(t, errors.As(wrappedErr, &preReleaseErr))

	// Test with different error type
	otherErr := errors.New("some other error")
	assert.False(t, errors.As(otherErr, &preReleaseErr))
}

func TestPreReleaseError_Implementation(t *testing.T) {
	err := &PreReleaseError{
		Language:          "python",
		RequestedVersion:  "3.14.0",
		AvailableVersions: []string{"3.14.0rc2"},
	}

	// Test that it implements error interface
	var _ error = err

	// Test error message is not empty
	assert.NotEmpty(t, err.Error())

	// Test fields are accessible
	assert.Equal(t, "python", err.Language)
	assert.Equal(t, "3.14.0", err.RequestedVersion)
	assert.Equal(t, []string{"3.14.0rc2"}, err.AvailableVersions)
}
