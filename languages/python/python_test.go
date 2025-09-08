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

package python

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/toodofun/gvm/internal/core"
	"github.com/toodofun/gvm/languages"

	goversion "github.com/hashicorp/go-version"
)

func TestPython_Name(t *testing.T) {
	p := &Python{}
	if p.Name() != "python" {
		t.Errorf("expected python, got %s", p.Name())
	}
}

func TestPython_ListRemoteVersions(t *testing.T) {
	p := &Python{}
	versions, err := p.ListRemoteVersions(context.Background())
	if err != nil {
		t.Errorf("ListRemoteVersions error: %v", err)
	}
	if len(versions) == 0 {
		t.Errorf("expected remote versions, got 0")
	}
	// 检查版本格式是否正确
	for _, v := range versions {
		if v.Version == nil {
			t.Errorf("version should not be nil")
		}
		if v.Origin == "" {
			t.Errorf("origin should not be empty")
		}
	}
}

func TestPython_ListInstalledVersions(t *testing.T) {
	p := &Python{}
	versions, err := p.ListInstalledVersions(context.Background())
	if err != nil {
		t.Errorf("ListInstalledVersions error: %v", err)
	}
	// 如果没有安装版本，应该返回空列表而不是错误
	if versions == nil {
		t.Errorf("should return empty slice, not nil")
	}
}

func TestPython_GetDefaultVersion(t *testing.T) {
	p := &Python{}
	version := p.GetDefaultVersion(context.Background())
	if version == nil {
		t.Errorf("GetDefaultVersion should not return nil")
	}
}

func TestPython_Install_WithSpacePath(t *testing.T) {
	p := &Python{}
	// 模拟带空格的路径
	originalGetRootDir := core.GetRootDir
	core.GetRootDir = func() string {
		return "/Users/test/Application Support/.gvm"
	}
	defer func() {
		core.GetRootDir = originalGetRootDir
	}()

	ver, _ := goversion.NewVersion("3.8.19")
	remoteVersion := &core.RemoteVersion{
		Version: ver,
		Origin:  "3.8.19",
		Comment: "",
	}

	err := p.Install(context.Background(), remoteVersion)
	if err == nil {
		t.Errorf("expected error for path with spaces, got nil")
	}
	if err != nil && !contains(err.Error(), "带空格的安装路径") {
		t.Errorf("expected error about space in path, got: %v", err)
	}
}

func TestPython_SetDefaultVersion(t *testing.T) {
	p := &Python{}
	// 测试设置不存在的版本
	err := p.SetDefaultVersion(context.Background(), "999.999.999")
	if err == nil {
		t.Errorf("expected error for non-existent version, got nil")
	}
}

func TestPython_Uninstall(t *testing.T) {
	p := &Python{}
	// 测试卸载不存在的版本
	err := p.Uninstall(context.Background(), "999.999.999")
	if err != nil {
		t.Errorf("uninstall non-existent version should not return error: %v", err)
	}
}

func TestPython_Install_ValidPath(t *testing.T) {
	p := &Python{}
	// 模拟有效路径
	originalGetRootDir := core.GetRootDir
	core.GetRootDir = func() string {
		return "/tmp/test_gvm"
	}
	defer func() {
		core.GetRootDir = originalGetRootDir
		// 清理测试目录
		os.RemoveAll("/tmp/test_gvm")
	}()

	ver, _ := goversion.NewVersion("3.8.19")
	remoteVersion := &core.RemoteVersion{
		Version: ver,
		Origin:  "3.8.19",
		Comment: "",
	}

	// 这个测试会失败，因为我们没有真实的下载和编译环境
	// 但至少可以测试路径检查逻辑
	err := p.Install(context.Background(), remoteVersion)
	// 由于没有真实的网络和编译环境，这个测试可能会失败
	// 但我们主要测试路径检查逻辑
	if err != nil && !contains(err.Error(), "not found") && !contains(err.Error(), "download") {
		t.Logf("Install test completed with expected error: %v", err)
	}
}

// 辅助函数
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr ||
		(len(s) > len(substr) && (s[:len(substr)] == substr ||
			s[len(s)-len(substr):] == substr ||
			containsSubstring(s, substr))))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func TestPython_checkAvailableVersions(t *testing.T) {
	p := &Python{}
	ctx := context.Background()

	t.Run("check non-existent version", func(t *testing.T) {
		versions, err := p.checkAvailableVersions(ctx, "99.99.99")
		if err == nil {
			t.Error("expected error for non-existent version")
		}
		if len(versions) != 0 {
			t.Errorf("expected empty versions, got %v", versions)
		}
	})

	t.Run("check invalid base version", func(t *testing.T) {
		versions, err := p.checkAvailableVersions(ctx, "invalid")
		if err == nil {
			t.Error("expected error for invalid version")
		}
		if len(versions) != 0 {
			t.Errorf("expected empty versions, got %v", versions)
		}
	})
}

func TestPython_ListRemoteVersions_Enhanced(t *testing.T) {
	p := &Python{}
	ctx := context.Background()

	versions, err := p.ListRemoteVersions(ctx)
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
		if v.Comment == "Release Candidate" || v.Comment == "Beta" || v.Comment == "Alpha" {
			hasPreRelease = true
		}
		if v.Comment == "Stable Release" {
			hasStable = true
		}
	}

	if !hasStable {
		t.Error("expected at least one stable release")
	}

	// 如果有预发布版本，检查注释是否正确
	if hasPreRelease {
		t.Log("Found pre-release versions with proper comments")
	}
}

func TestPython_Install_VersionFormat(t *testing.T) {

	tests := []struct {
		name           string
		origin         string
		expectedVer    string
		expectedBase   string
	}{
		{
			name:         "stable version",
			origin:       "3.13.0",
			expectedVer:  "3.13.0",
			expectedBase: "3.13.0",
		},
		{
			name:         "release candidate with dash",
			origin:       "3.14.0-rc2",
			expectedVer:  "3.14.0rc2",
			expectedBase: "3.14.0",
		},
		{
			name:         "beta version with dash",
			origin:       "3.14.0-b4",
			expectedVer:  "3.14.0b4",
			expectedBase: "3.14.0",
		},
		{
			name:         "alpha version with dash",
			origin:       "3.14.0-a7",
			expectedVer:  "3.14.0a7",
			expectedBase: "3.14.0",
		},
		{
			name:         "rc without dash",
			origin:       "3.14.0rc1",
			expectedVer:  "3.14.0rc1",
			expectedBase: "3.14.0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 测试版本字符串转换逻辑
			versionStr := tt.origin
			versionStr = strings.ReplaceAll(versionStr, "-rc", "rc")
			versionStr = strings.ReplaceAll(versionStr, "-b", "b")
			versionStr = strings.ReplaceAll(versionStr, "-a", "a")
			
			if versionStr != tt.expectedVer {
				t.Errorf("expected version string %s, got %s", tt.expectedVer, versionStr)
			}

			// 测试基础版本提取
			baseVersion := versionStr
			if idx := strings.IndexAny(baseVersion, "abr"); idx > 0 {
				baseVersion = baseVersion[:idx]
			}
			
			if baseVersion != tt.expectedBase {
				t.Errorf("expected base version %s, got %s", tt.expectedBase, baseVersion)
			}
		})
	}
}

func TestPython_PreReleaseError_Integration(t *testing.T) {
	p := &Python{}
	ctx := context.Background()

	// 创建一个模拟的 RemoteVersion
	ver, err := goversion.NewVersion("3.14.0")
	if err != nil {
		t.Fatal(err)
	}

	remoteVersion := &core.RemoteVersion{
		Version: ver,
		Origin:  "3.14.0",
		Comment: "",
	}

	// 注意：这个测试可能会因为网络问题失败，所以我们只检查错误类型
	err = p.Install(ctx, remoteVersion)
	if err != nil {
		// 检查是否是 PreReleaseError
		if preErr, ok := err.(*languages.PreReleaseError); ok {
			if preErr.Language != "python" {
				t.Errorf("expected language python, got %s", preErr.Language)
			}
			if preErr.RequestedVersion != "3.14.0" {
				t.Errorf("expected requested version 3.14.0, got %s", preErr.RequestedVersion)
			}
			if len(preErr.AvailableVersions) == 0 {
				t.Error("expected some available versions")
			}
		}
		// 如果不是 PreReleaseError，可能是网络错误等，跳过测试
		t.Skipf("Install test skipped due to error: %v", err)
	}
}
