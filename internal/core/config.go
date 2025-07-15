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

	"github.com/toodofun/gvm/internal/util/file"
)

const (
	defaultHomeDir = "/opt"
	defaultDir     = ".gvm"

	ContextLogWriterKey ctxKey = "context.log.writer"
)

var Version = "1.0.0-dev"

type Config struct {
	Addon []LanguageItem `json:"addon"`
}

type LanguageItem struct {
	Name           string `json:"name"`
	Provider       string `json:"provider"`
	DataSourceName string `json:"dsn"`
}

type ctxKey string

var GetRootDir = func() string {
	home, err := os.UserConfigDir()
	if err != nil {
		home = defaultHomeDir
	}

	rootDir := filepath.Join(home, defaultDir)
	if err := os.MkdirAll(rootDir, 0755); err != nil {
		panic("无法创建配置目录: " + err.Error())
	}
	return rootDir
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
