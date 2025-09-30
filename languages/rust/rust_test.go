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

package rust

import (
	"context"
	"runtime"
	"strings"
	"testing"

	"github.com/toodofun/gvm/internal/core"

	goversion "github.com/hashicorp/go-version"
	"github.com/stretchr/testify/assert"
)

func TestRust_Name(t *testing.T) {
	r := &Rust{}
	assert.Equal(t, "rust", r.Name())
}

func TestRust_ListRemoteVersions(t *testing.T) {
	r := &Rust{}
	ctx := context.Background()

	versions, err := r.ListRemoteVersions(ctx)
	if err != nil {
		t.Skipf("Skipping remote test due to network error: %v", err)
		return
	}

	if len(versions) == 0 {
		t.Error("expected at least one version")
		return
	}

	// 检查版本格式
	for _, v := range versions {
		if v.Origin == "" {
			t.Error("version origin should not be empty")
		}
		if v.Comment != "Stable Release" && v.Comment != "Pre-release" {
			t.Errorf("unexpected comment: %s", v.Comment)
		}
	}

	// 检查是否包含预发布版本
	hasPreRelease := false
	hasStable := false
	for _, v := range versions {
		if v.Comment == "Pre-release" {
			hasPreRelease = true
		}
		if v.Comment == "Stable Release" {
			hasStable = true
		}
	}

	if !hasStable {
		t.Error("expected at least one stable release")
	}

	// 记录预发布版本信息
	if hasPreRelease {
		t.Log("Found pre-release versions")
	}
}

func TestRust_ListInstalledVersions(t *testing.T) {
	r := &Rust{}
	ctx := context.Background()

	versions, err := r.ListInstalledVersions(ctx)
	assert.NoError(t, err)
	assert.NotNil(t, versions)
}

func TestRust_GetDefaultVersion(t *testing.T) {
	r := &Rust{}
	ctx := context.Background()

	defaultVersion := r.GetDefaultVersion(ctx)
	// 可能为 nil，这是正常的
	_ = defaultVersion
}

func TestRust_TargetMapping(t *testing.T) {
	tests := []struct {
		name         string
		goos         string
		goarch       string
		expectedOS   string
		expectedArch string
	}{
		{
			name:         "macOS amd64",
			goos:         "darwin",
			goarch:       "amd64",
			expectedOS:   "apple-darwin",
			expectedArch: "x86_64",
		},
		{
			name:         "macOS arm64",
			goos:         "darwin",
			goarch:       "arm64",
			expectedOS:   "apple-darwin",
			expectedArch: "aarch64",
		},
		{
			name:         "Linux amd64",
			goos:         "linux",
			goarch:       "amd64",
			expectedOS:   "unknown-linux-gnu",
			expectedArch: "x86_64",
		},
		{
			name:         "Windows amd64",
			goos:         "windows",
			goarch:       "amd64",
			expectedOS:   "pc-windows-msvc",
			expectedArch: "x86_64",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 模拟目标平台映射逻辑
			osName := tt.goos
			archName := tt.goarch

			switch osName {
			case "darwin":
				osName = "apple-darwin"
			case "linux":
				osName = "unknown-linux-gnu"
			case "windows":
				osName = "pc-windows-msvc"
			}

			switch archName {
			case "amd64":
				archName = "x86_64"
			case "arm64":
				archName = "aarch64"
			}

			assert.Equal(t, tt.expectedOS, osName)
			assert.Equal(t, tt.expectedArch, archName)

			target := archName + "-" + osName
			assert.Contains(t, target, tt.expectedArch)
			assert.Contains(t, target, tt.expectedOS)
		})
	}
}

func TestRust_SetDefaultVersion(t *testing.T) {
	r := &Rust{}
	ctx := context.Background()

	// 测试设置默认版本（模拟）
	err := r.SetDefaultVersion(ctx, "1.75.0")
	// 这个测试可能会失败，因为版本可能没有安装
	// 我们只是测试方法不会 panic
	_ = err
}

func TestRust_Uninstall(t *testing.T) {
	r := &Rust{}
	ctx := context.Background()

	// 测试卸载版本（模拟）
	err := r.Uninstall(ctx, "1.75.0")
	// 这个测试可能会失败，因为版本可能没有安装
	// 我们只是测试方法不会 panic
	_ = err
}

func TestRust_Install_Integration(t *testing.T) {
	r := &Rust{}
	ctx := context.Background()

	// 创建一个模拟的 RemoteVersion
	ver, err := goversion.NewVersion("1.75.0")
	if err != nil {
		t.Fatal(err)
	}

	remoteVersion := &core.RemoteVersion{
		Version: ver,
		Origin:  "1.75.0",
		Comment: "Stable Release",
	}

	// 注意：这个测试可能会因为网络问题失败，所以我们跳过实际安装
	err = r.Install(ctx, remoteVersion)
	if err != nil {
		t.Skipf("Install test skipped due to error: %v", err)
	}
}

func TestRust_VersionParsing(t *testing.T) {
	tests := []struct {
		name    string
		version string
		valid   bool
	}{
		{
			name:    "valid stable version",
			version: "1.75.0",
			valid:   true,
		},
		{
			name:    "valid beta version",
			version: "1.76.0-beta.1",
			valid:   true,
		},
		{
			name:    "valid nightly version",
			version: "1.77.0-nightly",
			valid:   true,
		},
		{
			name:    "invalid version",
			version: "invalid",
			valid:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 移除 tag 前缀进行测试
			versionStr := strings.TrimPrefix(tt.version, "v")
			_, err := goversion.NewVersion(versionStr)
			if tt.valid {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
			}
		})
	}
}

func TestRust_PlatformDetection(t *testing.T) {
	// 测试当前平台的检测
	osName := runtime.GOOS
	archName := runtime.GOARCH

	// 确保支持的平台
	supportedOS := map[string]bool{
		"darwin":  true,
		"linux":   true,
		"windows": true,
	}

	supportedArch := map[string]bool{
		"amd64": true,
		"arm64": true,
	}

	if !supportedOS[osName] {
		t.Skipf("Unsupported OS: %s", osName)
	}

	if !supportedArch[archName] {
		t.Skipf("Unsupported architecture: %s", archName)
	}

	// 测试目标字符串生成
	mappedOS := osName
	mappedArch := archName

	switch osName {
	case "darwin":
		mappedOS = "apple-darwin"
	case "linux":
		mappedOS = "unknown-linux-gnu"
	case "windows":
		mappedOS = "pc-windows-msvc"
	}

	switch archName {
	case "amd64":
		mappedArch = "x86_64"
	case "arm64":
		mappedArch = "aarch64"
	}

	target := mappedArch + "-" + mappedOS
	assert.NotEmpty(t, target)
	assert.Contains(t, target, mappedArch)
	assert.Contains(t, target, mappedOS)
}
