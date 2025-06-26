//go:build !windows

package env

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const (
	defaultEnvFile = ".gvmrc"
	pathSeparator  = ":"
)

const (
	ShellTypeBash ShellType = "bash"
	ShellTypeZsh  ShellType = "zsh"
	ShellTypeFish ShellType = "fish"
)

type ShellType string

func (m *Manager) GetEnv(key string) (string, error) {
	return m.getGvmEvn(key)
}

func (m *Manager) SetEnv(key, value string) error {
	if err := m.setGvmEnv(key, value); err != nil {
		return err
	}
	return m.appendToConfigFile()
}

func (m *Manager) DeleteEnv(key string) error {
	if err := m.deleteGvmEnv(key); err != nil {
		return err
	}
	return m.appendToConfigFile()
}

func (m *Manager) setGvmEnv(key, value string) error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	envFilePath := filepath.Join(homeDir, defaultEnvFile)
	data, _ := os.ReadFile(envFilePath) // 读取失败允许，可能文件不存在

	var newFileContent []string
	found := false
	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		l := strings.TrimPrefix(line, "export ")
		if strings.HasPrefix(l, key+"=") {
			newFileContent = append(newFileContent, "export "+key+"="+value)
			found = true
		} else {
			newFileContent = append(newFileContent, line)
		}
	}

	if !found {
		newFileContent = append(newFileContent, "export "+key+"="+value)
	}

	// 写回文件
	if err := os.MkdirAll(filepath.Dir(envFilePath), 0755); err != nil {
		return err
	}
	return os.WriteFile(envFilePath, []byte(strings.Join(newFileContent, "\n")), 0644)
}

func (m *Manager) getGvmEvn(key string) (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	envFilePath := filepath.Join(homeDir, defaultEnvFile)
	data, err := os.ReadFile(envFilePath)
	if err != nil {
		return "", nil // 如果文件不存在，返回空字符串
	}

	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		l := strings.TrimPrefix(line, "export ")
		if strings.HasPrefix(l, key+"=") {
			return strings.TrimSpace(strings.TrimPrefix(l, key+"=")), nil
		}
	}

	return "", nil
}

func (m *Manager) deleteGvmEnv(key string) error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	envFilePath := filepath.Join(homeDir, defaultEnvFile)
	data, err := os.ReadFile(envFilePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // 如果文件不存在，视为删除成功
		}
		return err
	}

	var newFileContent []string
	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		l := strings.TrimPrefix(line, "export ")
		if !strings.HasPrefix(l, key+"=") {
			newFileContent = append(newFileContent, line)
		}
	}

	// 写回文件
	return os.WriteFile(envFilePath, []byte(strings.Join(newFileContent, "\n")), 0644)
}

func (m *Manager) detectShell() ShellType {
	defaultShell := ShellTypeBash
	shellPath := os.Getenv("SHELL")
	if shellPath != "" {
		shellName := filepath.Base(shellPath)
		switch shellName {
		case "bash":
			return ShellTypeBash
		case "zsh":
			return ShellTypeZsh
		case "fish":
			return ShellTypeFish
		}
	}
	// 如果无法从环境变量获取，检测常见的配置
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return defaultShell
	}

	if _, err := os.Stat(filepath.Join(homeDir, ".zshrc")); err == nil {
		return ShellTypeZsh
	}

	if _, err := os.Stat(filepath.Join(homeDir, ".config/fish/config.fish")); err == nil {
		return ShellTypeFish
	}

	return defaultShell
}

func (m *Manager) getConfigFile() string {
	homeDir, _ := os.UserHomeDir()

	switch m.detectShell() {
	case ShellTypeFish:
		return filepath.Join(homeDir, ".config/fish/config.fish")
	case ShellTypeZsh:
		return filepath.Join(homeDir, ".zshrc")
	default:
		bashrc := filepath.Join(homeDir, ".bashrc")
		if _, err := os.Stat(bashrc); err == nil {
			return bashrc
		}
		return filepath.Join(homeDir, ".bash_profile")
	}
}

func (m *Manager) appendToConfigFile() error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	line := fmt.Sprintf("source %s", filepath.Join(homeDir, defaultEnvFile))
	cf := m.getConfigFile()
	dir := filepath.Dir(cf)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	// 先读取文件内容，检查是否已经包含新增内容
	existing := false
	if data, err := os.ReadFile(cf); err == nil {
		lines := strings.Split(string(data), "\n")
		for _, l := range lines {
			if strings.TrimSpace(l) == strings.TrimSpace(line) {
				existing = true
				break
			}
		}
	}

	if existing {
		return nil // 如果已经存在，则不需要追加
	}

	file, err := os.OpenFile(cf, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		return err
	}

	if stat.Size() > 0 {
		_, _ = file.Seek(-1, 2)
		lastChar := make([]byte, 1)
		_, _ = file.Read(lastChar)
		if lastChar[0] != '\n' {
			if _, err := file.WriteString("\n"); err != nil {
				return err
			}
		}
	}

	_, err = file.WriteString(line + "\n")
	return err
}
