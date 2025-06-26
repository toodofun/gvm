// Copyright 2025 The Toodofun Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http:www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package version

import (
	"encoding/json"
	"fmt"
	"runtime"
	"strings"
	"testing"
)

func TestInfo_ToJSON(t *testing.T) {
	tests := []struct {
		name string
		info Info
		want map[string]interface{}
	}{
		{
			name: "complete info",
			info: Info{
				GitVersion:   "v1.0.0",
				GitCommit:    "abc123def456",
				GitTreeState: "clean",
				BuildDate:    "2023-01-01T12:00:00Z",
				GoVersion:    "go1.21.0",
				Compiler:     "gc",
				Platform:     "linux/amd64",
			},
			want: map[string]interface{}{
				"gitVersion":   "v1.0.0",
				"gitCommit":    "abc123def456",
				"gitTreeState": "clean",
				"buildDate":    "2023-01-01T12:00:00Z",
				"goVersion":    "go1.21.0",
				"compiler":     "gc",
				"platform":     "linux/amd64",
			},
		},
		{
			name: "empty info",
			info: Info{},
			want: map[string]interface{}{
				"gitVersion":   "",
				"gitCommit":    "",
				"gitTreeState": "",
				"buildDate":    "",
				"goVersion":    "",
				"compiler":     "",
				"platform":     "",
			},
		},
		{
			name: "partial info",
			info: Info{
				GitVersion: "v2.0.0-beta",
				GitCommit:  "xyz789",
				BuildDate:  "2023-12-31T23:59:59Z",
			},
			want: map[string]interface{}{
				"gitVersion":   "v2.0.0-beta",
				"gitCommit":    "xyz789",
				"gitTreeState": "",
				"buildDate":    "2023-12-31T23:59:59Z",
				"goVersion":    "",
				"compiler":     "",
				"platform":     "",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.info.ToJSON()

			// 验证返回的是有效的JSON
			var gotMap map[string]interface{}
			if err := json.Unmarshal([]byte(got), &gotMap); err != nil {
				t.Errorf("ToJSON() returned invalid JSON: %v", err)
				return
			}

			// 比较每个字段
			for key, wantValue := range tt.want {
				if gotValue, exists := gotMap[key]; !exists {
					t.Errorf("ToJSON() missing key %s", key)
				} else if gotValue != wantValue {
					t.Errorf("ToJSON() key %s = %v, want %v", key, gotValue, wantValue)
				}
			}

			// 确保没有额外的字段
			if len(gotMap) != len(tt.want) {
				t.Errorf("ToJSON() returned %d fields, want %d", len(gotMap), len(tt.want))
			}
		})
	}
}

func TestInfo_ToJSON_ValidJSON(t *testing.T) {
	info := Info{
		GitVersion:   "v1.0.0",
		GitCommit:    "abc123",
		GitTreeState: "clean",
		BuildDate:    "2023-01-01T00:00:00Z",
		GoVersion:    runtime.Version(),
		Compiler:     runtime.Compiler,
		Platform:     fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH),
	}

	jsonStr := info.ToJSON()

	// 验证JSON格式正确
	var result map[string]interface{}
	if err := json.Unmarshal([]byte(jsonStr), &result); err != nil {
		t.Errorf("ToJSON() produced invalid JSON: %v", err)
	}

	// 验证JSON字符串不为空
	if jsonStr == "" {
		t.Error("ToJSON() returned empty string")
	}

	// 验证包含预期的JSON结构
	if !strings.Contains(jsonStr, "gitVersion") {
		t.Error("ToJSON() should contain gitVersion field")
	}
}

func TestGet(t *testing.T) {
	info := Get()

	// 测试返回的Info结构体包含所有必要字段
	if info.GitVersion == "" {
		t.Error("Get() GitVersion should not be empty")
	}

	if info.GoVersion == "" {
		t.Error("Get() GoVersion should not be empty")
	}

	if info.Compiler == "" {
		t.Error("Get() Compiler should not be empty")
	}

	if info.Platform == "" {
		t.Error("Get() Platform should not be empty")
	}

	// 验证运行时信息的正确性
	expectedGoVersion := runtime.Version()
	if info.GoVersion != expectedGoVersion {
		t.Errorf("Get() GoVersion = %s, want %s", info.GoVersion, expectedGoVersion)
	}

	expectedCompiler := runtime.Compiler
	if info.Compiler != expectedCompiler {
		t.Errorf("Get() Compiler = %s, want %s", info.Compiler, expectedCompiler)
	}

	expectedPlatform := fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH)
	if info.Platform != expectedPlatform {
		t.Errorf("Get() Platform = %s, want %s", info.Platform, expectedPlatform)
	}
}

func TestGet_DefaultValues(t *testing.T) {
	info := Get()

	// 测试默认值
	expectedGitVersion := GitVersion
	if info.GitVersion != expectedGitVersion {
		t.Errorf("Get() GitVersion = %s, want %s", info.GitVersion, expectedGitVersion)
	}

	expectedGitCommit := GitCommit
	if info.GitCommit != expectedGitCommit {
		t.Errorf("Get() GitCommit = %s, want %s", info.GitCommit, expectedGitCommit)
	}

	expectedGitTreeState := GitTreeState
	if info.GitTreeState != expectedGitTreeState {
		t.Errorf("Get() GitTreeState = %s, want %s", info.GitTreeState, expectedGitTreeState)
	}

	expectedBuildDate := BuildDate
	if info.BuildDate != expectedBuildDate {
		t.Errorf("Get() BuildDate = %s, want %s", info.BuildDate, expectedBuildDate)
	}
}

func TestGet_PlatformFormat(t *testing.T) {
	info := Get()

	// 验证Platform格式为 "OS/ARCH"
	parts := strings.Split(info.Platform, "/")
	if len(parts) != 2 {
		t.Errorf("Get() Platform format should be 'OS/ARCH', got %s", info.Platform)
	}

	if parts[0] == "" || parts[1] == "" {
		t.Errorf("Get() Platform should not have empty OS or ARCH, got %s", info.Platform)
	}
}

func TestInfo_JSONRoundTrip(t *testing.T) {
	original := Get()

	// 转换为JSON
	jsonStr := original.ToJSON()

	// 从JSON反序列化
	var restored Info
	if err := json.Unmarshal([]byte(jsonStr), &restored); err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	// 比较原始和恢复的数据
	if original.GitVersion != restored.GitVersion {
		t.Errorf("GitVersion mismatch: original=%s, restored=%s", original.GitVersion, restored.GitVersion)
	}

	if original.GitCommit != restored.GitCommit {
		t.Errorf("GitCommit mismatch: original=%s, restored=%s", original.GitCommit, restored.GitCommit)
	}

	if original.GitTreeState != restored.GitTreeState {
		t.Errorf("GitTreeState mismatch: original=%s, restored=%s", original.GitTreeState, restored.GitTreeState)
	}

	if original.BuildDate != restored.BuildDate {
		t.Errorf("BuildDate mismatch: original=%s, restored=%s", original.BuildDate, restored.BuildDate)
	}

	if original.GoVersion != restored.GoVersion {
		t.Errorf("GoVersion mismatch: original=%s, restored=%s", original.GoVersion, restored.GoVersion)
	}

	if original.Compiler != restored.Compiler {
		t.Errorf("Compiler mismatch: original=%s, restored=%s", original.Compiler, restored.Compiler)
	}

	if original.Platform != restored.Platform {
		t.Errorf("Platform mismatch: original=%s, restored=%s", original.Platform, restored.Platform)
	}
}

func TestPackageVariables(t *testing.T) {
	// 测试包级变量的默认值
	if GitVersion == "" {
		t.Error("GitVersion should have a default value")
	}

	if BuildDate == "" {
		t.Error("BuildDate should have a default value")
	}

	if GitCommit == "" {
		t.Error("GitCommit should have a default value")
	}

	// GitTreeState可以为空，这是正常的
}
