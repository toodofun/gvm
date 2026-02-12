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

package env

import (
	"fmt"
	"runtime"
	"strings"

	"github.com/duke-git/lancet/v2/slice"
)

const (
	RuntimeFromWindows   = "windows"
	RuntimeFromLinux     = "linux"
	RuntimeFromDarwin    = "darwin"
	RuntimeFromMacos     = "macos"
	RuntimeFromApple     = "apple-darwin"
	RuntimeUnknown       = "unknown-linux-gnu"
	RuntimeFromWindowsPC = "pc-windows-msvc"
	ArchAMD64            = "amd64"
	ArchARM64            = "arm64"
	Arch386              = "386"
	ArchARM              = "arm"
	ArchArmv7l           = "armv7l"
	ArchX86              = "x86"
	ArchX86And64         = "x86_64"
	ArchX64              = "x64"
	ArchARMGeneric       = "arm"
	Aarch64              = "aarch64"
	Bitness32            = "32"
	Bitness64            = "64"
)

type IManager interface {
	// AppendEnv 向环境变量中追加一个值
	AppendEnv(key, value string) error
	// RemoveEnv 从环境变量中移除一个值
	RemoveEnv(key, value string) error
	// GetEnv 获取环境变量的值
	GetEnv(key string) (string, error)
	// SetEnv 设置环境变量的值
	SetEnv(key, value string) error
	// DeleteEnv 删除环境变量
	DeleteEnv(key string) error
}
type Manager struct {
}

func NewEnvManager() *Manager {
	return &Manager{}
}

type KV struct {
	Key    string
	Value  string
	Append bool
}

func (m *Manager) AppendEnv(key, value string) error {
	if len(value) == 0 {
		return fmt.Errorf("value is empty")
	}

	value = m.quoteValue(value)

	values, err := m.GetEnv(key)
	if err != nil {
		return err
	}
	var valueList []string

	if len(values) > 0 {
		valueList = strings.Split(value, pathSeparator)
		valueList = append(valueList, strings.Split(values, pathSeparator)...)
	} else {
		valueList = []string{value}
	}

	if runtime.GOOS != RuntimeFromWindows {
		valueList = append(valueList, "$"+key)
	}

	// 去重
	valueList = slice.Unique(valueList)
	return m.SetEnv(key, strings.Join(valueList, pathSeparator))
}

func (m *Manager) RemoveEnv(key, value string) error {
	if len(value) == 0 {
		return fmt.Errorf("value is empty")
	}

	value = m.quoteValue(value)

	values, err := m.GetEnv(key)
	if err != nil {
		return err
	}
	if len(values) == 0 {
		return nil
	}

	valueList := strings.Split(values, pathSeparator)
	newValueList := slice.Filter(valueList, func(index int, item string) bool {
		return item != value
	})

	if len(newValueList) == 0 {
		return m.DeleteEnv(key)
	}

	if runtime.GOOS != RuntimeFromWindows {
		newValueList = append(newValueList, "$"+key)
	}
	// 去重
	newValueList = slice.Unique(newValueList)
	return m.SetEnv(key, strings.Join(newValueList, pathSeparator))
}

func (m *Manager) quoteValue(value string) string {
	if runtime.GOOS == RuntimeFromWindows {
		return value
	}
	// 如果值包含空格或特殊字符，需要加引号
	if strings.ContainsAny(value, " \t\n\r\"'\\$`") && !strings.Contains(value, pathSeparator) {
		return fmt.Sprintf("\"%s\"", strings.ReplaceAll(value, "\"", "\\\""))
	}
	return value
}
