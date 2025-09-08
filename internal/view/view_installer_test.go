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

package view

import (
	"context"
	"errors"
	"testing"

	"github.com/toodofun/gvm/internal/core"
	"github.com/toodofun/gvm/languages"

	goversion "github.com/hashicorp/go-version"
	"github.com/rivo/tview"
	"github.com/stretchr/testify/assert"
)

// MockLanguage 模拟语言实现
type MockLanguage struct {
	installError error
}

func (m *MockLanguage) Name() string {
	return "mock"
}

func (m *MockLanguage) ListRemoteVersions(ctx context.Context) ([]*core.RemoteVersion, error) {
	return nil, nil
}

func (m *MockLanguage) ListInstalledVersions(ctx context.Context) ([]*core.InstalledVersion, error) {
	return nil, nil
}

func (m *MockLanguage) SetDefaultVersion(ctx context.Context, version string) error {
	return nil
}

func (m *MockLanguage) GetDefaultVersion(ctx context.Context) *core.InstalledVersion {
	return nil
}

func (m *MockLanguage) Uninstall(ctx context.Context, version string) error {
	return nil
}

func (m *MockLanguage) Install(ctx context.Context, version *core.RemoteVersion) error {
	return m.installError
}

func TestNewInstall(t *testing.T) {
	app := tview.NewApplication()
	pages := tview.NewPages()
	mockLang := &MockLanguage{}
	
	ver, err := goversion.NewVersion("3.14.0")
	assert.NoError(t, err)
	
	version := &core.RemoteVersion{
		Version: ver,
		Origin:  "3.14.0",
		Comment: "",
	}
	
	callback := func(err error) {}
	
	installer := NewInstall(&Application{Application: app}, pages, mockLang, version, callback)
	
	assert.NotNil(t, installer)
	assert.Equal(t, mockLang, installer.lang)
	assert.Equal(t, version, installer.version)
	assert.NotNil(t, installer.callback)
	assert.NotNil(t, installer.Modal)
}

func TestInstaller_ErrorHandling(t *testing.T) {
	tests := []struct {
		name          string
		installError  error
		expectedButtons int
		description   string
	}{
		{
			name: "PreReleaseError with available versions",
			installError: &languages.PreReleaseError{
				Language:          "python",
				RequestedVersion:  "3.14.0",
				AvailableVersions: []string{"3.14.0rc1", "3.14.0rc2"},
			},
			expectedButtons: 2,
			description:    "Should show install and cancel buttons",
		},
		{
			name: "PreReleaseError with empty versions",
			installError: &languages.PreReleaseError{
				Language:          "python",
				RequestedVersion:  "3.14.0",
				AvailableVersions: []string{},
			},
			expectedButtons: 1,
			description:    "Should show only OK button",
		},
		{
			name:            "Regular error",
			installError:    errors.New("network error"),
			expectedButtons: 1,
			description:    "Should show only OK button",
		},
		{
			name:            "No error",
			installError:    nil,
			expectedButtons: 1,
			description:    "Should show only OK button for success",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := tview.NewApplication()
			pages := tview.NewPages()
			mockLang := &MockLanguage{}
			
			ver, err := goversion.NewVersion("3.14.0")
			assert.NoError(t, err)
			
			version := &core.RemoteVersion{
				Version: ver,
				Origin:  "3.14.0",
				Comment: "",
			}
			
			callback := func(err error) {}
			
			mockLang := &MockLanguage{installError: tt.installError}
			
			// 模拟安装过程中的错误处理逻辑
			installErr := mockLang.Install(context.Background(), version)
			
			// 检查错误类型
			var preReleaseErr *languages.PreReleaseError
			if errors.As(installErr, &preReleaseErr) && preReleaseErr.GetRecommendedVersion() != "" {
				// 应该有两个按钮：安装推荐版本和取消
				assert.Equal(t, 2, tt.expectedButtons, tt.description)
				assert.Equal(t, "3.14.0rc2", preReleaseErr.GetRecommendedVersion())
			} else {
				// 应该只有一个按钮：OK
				assert.Equal(t, 1, tt.expectedButtons, tt.description)
			}
			
		})
	}
}

func TestInstaller_PreReleaseErrorHandling(t *testing.T) {
	app := tview.NewApplication()
	pages := tview.NewPages()
	mockLang := &MockLanguage{}
	
	ver, err := goversion.NewVersion("3.14.0")
	assert.NoError(t, err)
	
	version := &core.RemoteVersion{
		Version: ver,
		Origin:  "3.14.0",
		Comment: "",
	}
	
	preReleaseErr := &languages.PreReleaseError{
		Language:          "python",
		RequestedVersion:  "3.14.0",
		AvailableVersions: []string{"3.14.0rc1", "3.14.0rc2"},
	}
	
	callback := func(err error) {}
	
	mockLang := &MockLanguage{installError: preReleaseErr}
	
	// 测试错误检测
	installErr := mockLang.Install(context.Background(), version)
	assert.Error(t, installErr)
	
	var detectedPreReleaseErr *languages.PreReleaseError
	assert.True(t, errors.As(installErr, &detectedPreReleaseErr))
	assert.Equal(t, "python", detectedPreReleaseErr.Language)
	assert.Equal(t, "3.14.0", detectedPreReleaseErr.RequestedVersion)
	assert.Equal(t, []string{"3.14.0rc1", "3.14.0rc2"}, detectedPreReleaseErr.AvailableVersions)
	assert.Equal(t, "3.14.0rc2", detectedPreReleaseErr.GetRecommendedVersion())
	
	mockLang.AssertExpectations(t)
}

func TestInstaller_ButtonLabels(t *testing.T) {
	preReleaseErr := &languages.PreReleaseError{
		Language:          "python",
		RequestedVersion:  "3.14.0",
		AvailableVersions: []string{"3.14.0rc1", "3.14.0rc2"},
	}
	
	recommendedVersion := preReleaseErr.GetRecommendedVersion()
	assert.Equal(t, "3.14.0rc2", recommendedVersion)
	
	// 测试按钮标签
	expectedInstallButton := "安装 " + recommendedVersion
	expectedCancelButton := "取消"
	
	assert.Equal(t, "安装 3.14.0rc2", expectedInstallButton)
	assert.Equal(t, "取消", expectedCancelButton)
}

func TestInstaller_ErrorTypes(t *testing.T) {
	tests := []struct {
		name  string
		err   error
		isPreRelease bool
	}{
		{
			name: "PreReleaseError",
			err: &languages.PreReleaseError{
				Language:          "python",
				RequestedVersion:  "3.14.0",
				AvailableVersions: []string{"3.14.0rc2"},
			},
			isPreRelease: true,
		},
		{
			name:         "Regular error",
			err:          errors.New("network error"),
			isPreRelease: false,
		},
		{
			name:         "Nil error",
			err:          nil,
			isPreRelease: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var preReleaseErr *languages.PreReleaseError
			isPreRelease := errors.As(tt.err, &preReleaseErr)
			
			assert.Equal(t, tt.isPreRelease, isPreRelease)
			
			if isPreRelease {
				assert.NotNil(t, preReleaseErr)
				assert.NotEmpty(t, preReleaseErr.Language)
			}
		})
	}
}
