package core

import (
	"os"
	"path/filepath"
)

const (
	defaultHomeDir = "/opt"
	defaultDir     = ".gvm"
)

var Version = "1.0.0-dev"

func GetRootDir() string {
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
