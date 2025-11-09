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

package core

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/toodofun/gvm/internal/util/file"
)

const (
	defaultDir                 = ".gvm"
	ContextLogWriterKey ctxKey = "context.log.writer"
)

var Version = "1.0.0-dev"

type Config struct {
	Language string         `json:"language"`
	Addon    []LanguageItem `json:"addon"`
}

type LanguageItem struct {
	Name           string `json:"name"`
	Provider       string `json:"provider"`
	DataSourceName string `json:"dsn"`
}

type ctxKey string

var GetRootDir = func() string {
	// 1. 优先使用环境变量 GVM_ROOT
	if custom := os.Getenv("GVM_ROOT"); custom != "" {
		if !strings.Contains(custom, " ") {
			ensureDir(custom)
			return custom
		}
	}

	// 2. 其次使用 XDG 配置目录
	if cfgDir, err := os.UserConfigDir(); err == nil && cfgDir != "" {
		path := filepath.Join(cfgDir, defaultDir)
		if !strings.Contains(path, " ") {
			ensureDir(path)
			return path
		}
	}

	// 3. 再退回到 $HOME/.gvm
	home := os.Getenv("HOME")
	if home != "" {
		path := filepath.Join(home, defaultDir)
		if !strings.Contains(path, " ") {
			ensureDir(path)
			return path
		}
	}

	// 4. 最后兜底到 /opt/gvm
	fallback := filepath.Join("/opt", "gvm")
	ensureDir(fallback)
	return fallback
}

// ensureDir 创建目录并在失败时 panic
func ensureDir(dir string) {
	if err := os.MkdirAll(dir, 0755); err != nil {
		panic("无法创建配置目录: " + err.Error())
	}
}

func GetConfigPath() string {
	return filepath.Join(GetRootDir(), "config.json")
}

func GetConfig() *Config {
	config := new(Config)
	if err := file.ReadJSONFile(GetConfigPath(), config); err != nil {
		return config
	}
	return config
}
