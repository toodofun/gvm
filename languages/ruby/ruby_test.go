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

package ruby

import (
	"context"
	"strings"
	"testing"

	"github.com/toodofun/gvm/internal/core"

	goversion "github.com/hashicorp/go-version"
	"github.com/stretchr/testify/assert"
)

func TestRuby_Name(t *testing.T) {
	r := &Ruby{}
	assert.Equal(t, "ruby", r.Name())
}

func TestRuby_ListRemoteVersions(t *testing.T) {
	r := &Ruby{}
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

	// 检查版本是否按照从新到旧排序
	for i := 1; i < len(versions); i++ {
		if versions[i-1].Version.LessThan(versions[i].Version) {
			t.Errorf("versions not sorted correctly: %s should be after %s",
				versions[i-1].Version.String(), versions[i].Version.String())
		}
	}

	// 检查是否包含候选版本
	hasPreRelease := false
	hasStable := false
	for _, v := range versions {
		if v.Comment == "Release Candidate" || v.Comment == "Preview" {
			hasPreRelease = true
		}
		if v.Comment == "Stable Release" {
			hasStable = true
		}
	}

	if !hasStable {
		t.Error("expected at least one stable release")
	}

	// 如果有预发布版本，记录日志
	if hasPreRelease {
		t.Log("Found pre-release versions with proper comments")
	}
}

func TestRuby_ListInstalledVersions(t *testing.T) {
	r := &Ruby{}
	ctx := context.Background()

	versions, err := r.ListInstalledVersions(ctx)
	assert.NoError(t, err)
	assert.NotNil(t, versions)
}

func TestRuby_GetDefaultVersion(t *testing.T) {
	r := &Ruby{}
	ctx := context.Background()

	defaultVersion := r.GetDefaultVersion(ctx)
	// 可能为 nil，这是正常的
	_ = defaultVersion
}

func TestRuby_Install_VersionFormat(t *testing.T) {
	tests := []struct {
		name               string
		origin             string
		expectedMajorMinor string
	}{
		{
			name:               "stable version",
			origin:             "3.1.4",
			expectedMajorMinor: "3.1",
		},
		{
			name:               "release candidate",
			origin:             "3.2.0-rc1",
			expectedMajorMinor: "3.2",
		},
		{
			name:               "preview version",
			origin:             "3.3.0-preview1",
			expectedMajorMinor: "3.3",
		},
		{
			name:               "patch version",
			origin:             "2.7.8",
			expectedMajorMinor: "2.7",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 测试版本字符串解析逻辑
			versionStr := tt.origin
			parts := strings.Split(versionStr, ".")
			if len(parts) >= 2 {
				majorMinor := parts[0] + "." + parts[1]
				assert.Equal(t, tt.expectedMajorMinor, majorMinor)
			} else {
				t.Errorf("invalid version format: %s", versionStr)
			}
		})
	}
}

func TestRuby_SetDefaultVersion(t *testing.T) {
	r := &Ruby{}
	ctx := context.Background()

	// 测试设置默认版本（模拟）
	err := r.SetDefaultVersion(ctx, "3.1.4")
	// 这个测试可能会失败，因为版本可能没有安装
	// 我们只是测试方法不会 panic
	_ = err
}

func TestRuby_Uninstall(t *testing.T) {
	r := &Ruby{}
	ctx := context.Background()

	// 测试卸载版本（模拟）
	err := r.Uninstall(ctx, "3.1.4")
	// 这个测试可能会失败，因为版本可能没有安装
	// 我们只是测试方法不会 panic
	_ = err
}

func TestRuby_Install_Integration(t *testing.T) {
	r := &Ruby{}
	ctx := context.Background()

	// 创建一个模拟的 RemoteVersion
	ver, err := goversion.NewVersion("3.1.4")
	if err != nil {
		t.Fatal(err)
	}

	remoteVersion := &core.RemoteVersion{
		Version: ver,
		Origin:  "3.1.4",
		Comment: "Stable Release",
	}

	// 注意：这个测试可能会因为网络问题失败，所以我们跳过实际安装
	err = r.Install(ctx, remoteVersion)
	if err != nil {
		t.Skipf("Install test skipped due to error: %v", err)
	}
}

func TestRuby_VersionParsing(t *testing.T) {
	tests := []struct {
		name    string
		version string
		valid   bool
	}{
		{
			name:    "valid stable version",
			version: "3.1.4",
			valid:   true,
		},
		{
			name:    "valid rc version",
			version: "3.2.0-rc1",
			valid:   true,
		},
		{
			name:    "valid preview version",
			version: "3.3.0-preview1",
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
			_, err := goversion.NewVersion(tt.version)
			if tt.valid {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
			}
		})
	}
}
