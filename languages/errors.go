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
	"fmt"
	"strings"
)

// PreReleaseError 预发布版本错误类型
type PreReleaseError struct {
	Language          string   // 语言名称
	RequestedVersion  string   // 用户请求的版本
	AvailableVersions []string // 可用的候选版本
}

func (e *PreReleaseError) Error() string {
	if len(e.AvailableVersions) > 0 {
		return fmt.Sprintf("版本 %s 尚未正式发布。可用的候选版本：%s\n请使用完整版本号安装，例如：gvm install %s %s",
			e.RequestedVersion, strings.Join(e.AvailableVersions, ", "), e.Language, e.AvailableVersions[len(e.AvailableVersions)-1])
	}
	return fmt.Sprintf("版本 %s 尚未正式发布", e.RequestedVersion)
}

// GetRecommendedVersion 获取推荐的候选版本（通常是最新的）
func (e *PreReleaseError) GetRecommendedVersion() string {
	if len(e.AvailableVersions) > 0 {
		return e.AvailableVersions[len(e.AvailableVersions)-1]
	}
	return ""
}
