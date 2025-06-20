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

package common

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"syscall"
)

// PathManager PATH 环境变量管理器
type PathManager struct {
	shellType     string
	configFile    string
	pathSeparator string
	cachedPaths   []string // 缓存的路径列表
	cacheValid    bool     // 缓存是否有效
}

// PathInfo 路径信息
type PathInfo struct {
	Path   string // 路径
	Exists bool   // 路径是否存在
	Index  int    // 在 PATH 中的索引位置
}

// 支持的 shell 类型
const (
	BASH = "bash"
	ZSH  = "zsh"
	FISH = "fish"
)

// 位置常量
const (
	PositionAppend  = "append"
	PositionPrepend = "prepend"
)

// NewPathManager 创建新的 PATH 管理器
func NewPathManager() (*PathManager, error) {
	pm := &PathManager{
		cacheValid: false,
	}

	// 设置路径分隔符
	if runtime.GOOS == "windows" {
		pm.pathSeparator = ";"
	} else {
		pm.pathSeparator = ":"
	}

	if runtime.GOOS == "windows" {
		// Windows 使用注册表，不需要检测 shell
		return pm, nil
	}

	// 检测当前使用的 shell
	shell, err := pm.detectShell()
	if err != nil {
		return nil, fmt.Errorf("无法检测 shell 类型: %v", err)
	}

	pm.shellType = shell
	pm.configFile = pm.getConfigFile()

	return pm, nil
}

// detectShell 检测当前使用的 shell
func (pm *PathManager) detectShell() (string, error) {
	// 首先尝试从 SHELL 环境变量获取
	shellPath := os.Getenv("SHELL")
	if shellPath != "" {
		shellName := filepath.Base(shellPath)
		switch shellName {
		case "bash":
			return BASH, nil
		case "zsh":
			return ZSH, nil
		case "fish":
			return FISH, nil
		}
	}

	// 如果无法从环境变量获取，尝试检测常见的配置文件
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	// 检查 zsh 配置文件
	if _, err := os.Stat(filepath.Join(homeDir, ".zshrc")); err == nil {
		return ZSH, nil
	}

	// 检查 fish 配置文件
	if _, err := os.Stat(filepath.Join(homeDir, ".config/fish/config.fish")); err == nil {
		return FISH, nil
	}

	// 默认使用 bash
	return BASH, nil
}

// getConfigFile 获取配置文件路径
func (pm *PathManager) getConfigFile() string {
	homeDir, _ := os.UserHomeDir()

	switch pm.shellType {
	case ZSH:
		return filepath.Join(homeDir, ".zshrc")
	case FISH:
		return filepath.Join(homeDir, ".config/fish/config.fish")
	default: // BASH
		// 尝试 .bashrc，如果不存在则使用 .bash_profile
		bashrc := filepath.Join(homeDir, ".bashrc")
		if _, err := os.Stat(bashrc); err == nil {
			return bashrc
		}
		return filepath.Join(homeDir, ".bash_profile")
	}
}

// getCurrentPATH 获取当前实际的PATH值
func (pm *PathManager) getCurrentPATH() string {
	if runtime.GOOS == "windows" {
		return pm.getWindowsPATH()
	}
	return pm.getUnixPATH()
}

// getWindowsPATH 获取Windows系统的PATH
func (pm *PathManager) getWindowsPATH() string {
	// 先尝试从注册表获取用户PATH
	userPath := pm.getUserPathFromRegistry()

	// 如果用户PATH不为空，使用用户PATH
	if userPath != "" {
		return userPath
	}

	// 否则回退到环境变量
	return os.Getenv("PATH")
}

// getUserPathFromRegistry 从注册表获取用户PATH
func (pm *PathManager) getUserPathFromRegistry() string {
	cmd := exec.Command("reg", "query", "HKCU\\Environment", "/v", "PATH")
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true} // Windows下隐藏命令行窗口

	output, err := cmd.Output()
	if err != nil {
		return ""
	}

	return pm.parseWindowsRegOutput(string(output))
}

// parseWindowsRegOutput 解析Windows注册表输出
func (pm *PathManager) parseWindowsRegOutput(output string) string {
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		if strings.Contains(line, "PATH") && strings.Contains(line, "REG_") {
			parts := strings.Fields(line)
			if len(parts) >= 3 {
				return strings.Join(parts[2:], " ")
			}
		}
	}
	return ""
}

// getUnixPATH 获取Unix系统的PATH
func (pm *PathManager) getUnixPATH() string {
	// 方法1: 解析配置文件获取PATH设置
	if pm.configFile != "" {
		if pathFromConfig := pm.getPathFromConfigFile(); pathFromConfig != "" {
			return pathFromConfig
		}
	}

	// 方法2: 回退到当前进程的环境变量
	return os.Getenv("PATH")
}

// getPathFromConfigFile 从配置文件中解析PATH设置
func (pm *PathManager) getPathFromConfigFile() string {
	content, err := os.ReadFile(pm.configFile)
	if err != nil {
		return ""
	}

	lines := strings.Split(string(content), "\n")
	var pathParts []string
	basePath := os.Getenv("PATH") // 获取基础PATH

	// 按行解析配置文件中的PATH设置
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		if pm.shellType == FISH {
			// 解析 fish 的 PATH 设置: set -gx PATH /new/path $PATH
			if strings.Contains(line, "set") && strings.Contains(line, "PATH") {
				pathPart := pm.extractPathFromFishLine(line)
				if pathPart != "" {
					pathParts = append(pathParts, pathPart)
				}
			}
		} else {
			// 解析 bash/zsh 的 PATH 设置: export PATH=/new/path:$PATH
			if strings.Contains(line, "export") && strings.Contains(line, "PATH") {
				pathPart := pm.extractPathFromExportLine(line)
				if pathPart != "" {
					pathParts = append(pathParts, pathPart)
				}
			}
		}
	}

	// 如果没有找到PATH设置，返回空字符串
	if len(pathParts) == 0 {
		return ""
	}

	// 构建完整的PATH
	// 这里简化处理，将所有找到的路径添加到基础PATH前面
	allPaths := strings.Join(pathParts, pm.pathSeparator)
	if basePath != "" {
		return allPaths + pm.pathSeparator + basePath
	}
	return allPaths
}

// extractPathFromExportLine 从export行中提取路径
func (pm *PathManager) extractPathFromExportLine(line string) string {
	// 简化的解析逻辑，提取 export PATH=... 中的新路径部分
	if !strings.Contains(line, "export PATH=") {
		return ""
	}

	// 提取等号后的部分
	parts := strings.SplitN(line, "export PATH=", 2)
	if len(parts) != 2 {
		return ""
	}

	pathValue := strings.TrimSpace(parts[1])

	// 如果包含 $PATH，提取新增的部分
	if strings.Contains(pathValue, "$PATH") {
		// 处理 /new/path:$PATH 或 $PATH:/new/path 的情况
		pathValue = strings.ReplaceAll(pathValue, "$PATH", "")
		pathValue = strings.Trim(pathValue, ":")
		pathValue = strings.Trim(pathValue, `"'`)
		return pathValue
	}

	// 如果不包含$PATH，返回整个值（但这种情况下会覆盖原PATH，需要谨慎）
	return strings.Trim(pathValue, `"'`)
}

// extractPathFromFishLine 从fish行中提取路径
func (pm *PathManager) extractPathFromFishLine(line string) string {
	// 简化的解析逻辑，提取 set -gx PATH ... 中的新路径部分
	if !strings.Contains(line, "set -gx PATH") {
		return ""
	}

	// 移除 set -gx PATH 部分
	pathPart := strings.Replace(line, "set -gx PATH", "", 1)
	pathPart = strings.TrimSpace(pathPart)

	// 如果包含 $PATH，提取新增的部分
	if strings.Contains(pathPart, "$PATH") {
		pathPart = strings.ReplaceAll(pathPart, "$PATH", "")
		pathPart = strings.TrimSpace(pathPart)
		return pathPart
	}

	return pathPart
}

// invalidateCache 使缓存失效
func (pm *PathManager) invalidateCache() {
	pm.cacheValid = false
	pm.cachedPaths = nil
}

// GetPaths 获取当前 PATH 的所有路径信息
func (pm *PathManager) GetPaths() []PathInfo {
	pathValue := pm.getCurrentPATH()
	if pathValue == "" {
		return []PathInfo{}
	}

	paths := strings.Split(pathValue, pm.pathSeparator)
	var pathInfos []PathInfo

	for i, path := range paths {
		trimmed := strings.TrimSpace(path)
		if trimmed == "" {
			continue
		}

		_, err := os.Stat(trimmed)
		pathInfos = append(pathInfos, PathInfo{
			Path:   trimmed,
			Exists: err == nil,
			Index:  i,
		})
	}

	return pathInfos
}

// GetPathStrings 获取当前 PATH 的所有路径字符串
func (pm *PathManager) GetPathStrings() []string {
	pathInfos := pm.GetPaths()
	paths := make([]string, len(pathInfos))
	for i, info := range pathInfos {
		paths[i] = info.Path
	}
	return paths
}

// Contains 检查 PATH 中是否包含指定路径
func (pm *PathManager) Contains(targetPath string) bool {
	paths := pm.GetPathStrings()
	targetPath = pm.normalizePath(targetPath)

	for _, path := range paths {
		if pm.normalizePath(path) == targetPath {
			return true
		}
	}
	return false
}

// IndexOf 获取指定路径在 PATH 中的索引位置，如果不存在返回 -1
func (pm *PathManager) IndexOf(targetPath string) int {
	paths := pm.GetPathStrings()
	targetPath = pm.normalizePath(targetPath)

	for i, path := range paths {
		if pm.normalizePath(path) == targetPath {
			return i
		}
	}
	return -1
}

// Search 搜索包含指定字符串的路径
func (pm *PathManager) Search(partialPath string) []PathInfo {
	allPaths := pm.GetPaths()
	var matches []PathInfo

	partialPath = strings.ToLower(partialPath)

	for _, pathInfo := range allPaths {
		if strings.Contains(strings.ToLower(pathInfo.Path), partialPath) {
			matches = append(matches, pathInfo)
		}
	}
	return matches
}

// Add 添加路径到 PATH
func (pm *PathManager) Add(newPath string, position string) error {
	// 检查路径是否已存在
	if pm.Contains(newPath) {
		return fmt.Errorf("路径 '%s' 已存在于 PATH 中", newPath)
	}

	// 验证位置参数
	if position != PositionAppend && position != PositionPrepend {
		return fmt.Errorf("无效的位置参数: %s，必须是 'append' 或 'prepend'", position)
	}

	var err error
	if runtime.GOOS == "windows" {
		err = pm.addWindows(newPath, position)
	} else {
		err = pm.addUnix(newPath, position)
	}

	if err == nil {
		pm.invalidateCache() // 添加成功后使缓存失效
	}

	return err
}

// AddIfNotExists 如果路径不存在则添加到 PATH
func (pm *PathManager) AddIfNotExists(newPath string, position string) error {
	if pm.Contains(newPath) {
		return nil // 已存在，不需要添加
	}
	return pm.Add(newPath, position)
}

// Remove 从 PATH 中移除指定路径
func (pm *PathManager) Remove(targetPath string) error {
	if !pm.Contains(targetPath) {
		return fmt.Errorf("路径 '%s' 不存在于 PATH 中", targetPath)
	}

	var err error
	if runtime.GOOS == "windows" {
		err = pm.removeWindows(targetPath)
	} else {
		err = pm.removeUnix(targetPath)
	}

	if err == nil {
		pm.invalidateCache() // 删除成功后使缓存失效
	}

	return err
}

// RemoveIfExists 如果路径存在则从 PATH 中移除
func (pm *PathManager) RemoveIfExists(targetPath string) error {
	if !pm.Contains(targetPath) {
		return nil // 不存在，不需要移除
	}
	return pm.Remove(targetPath)
}

// RemoveByIndex 根据索引移除路径
func (pm *PathManager) RemoveByIndex(index int) error {
	paths := pm.GetPathStrings()
	if index < 0 || index >= len(paths) {
		return fmt.Errorf("索引 %d 超出范围，PATH 中有 %d 个路径", index, len(paths))
	}

	return pm.Remove(paths[index])
}

// RemoveMatching 移除所有匹配的路径
func (pm *PathManager) RemoveMatching(partialPath string) ([]string, error) {
	matches := pm.Search(partialPath)
	if len(matches) == 0 {
		return []string{}, nil
	}

	var removed []string
	for _, match := range matches {
		if err := pm.Remove(match.Path); err != nil {
			return removed, err
		}
		removed = append(removed, match.Path)
	}

	return removed, nil
}

// Clean 清理 PATH 中的无效路径（不存在的目录）
func (pm *PathManager) Clean() ([]string, error) {
	pathInfos := pm.GetPaths()
	var invalidPaths []string

	// 收集无效路径
	for _, info := range pathInfos {
		if !info.Exists {
			invalidPaths = append(invalidPaths, info.Path)
		}
	}

	// 移除无效路径
	for _, invalidPath := range invalidPaths {
		if err := pm.Remove(invalidPath); err != nil {
			return invalidPaths, err
		}
	}

	return invalidPaths, nil
}

// GetValidPaths 获取所有有效的路径
func (pm *PathManager) GetValidPaths() []string {
	pathInfos := pm.GetPaths()
	var validPaths []string

	for _, info := range pathInfos {
		if info.Exists {
			validPaths = append(validPaths, info.Path)
		}
	}

	return validPaths
}

// GetInvalidPaths 获取所有无效的路径
func (pm *PathManager) GetInvalidPaths() []string {
	pathInfos := pm.GetPaths()
	var invalidPaths []string

	for _, info := range pathInfos {
		if !info.Exists {
			invalidPaths = append(invalidPaths, info.Path)
		}
	}

	return invalidPaths
}

// Move 移动路径到新位置
func (pm *PathManager) Move(targetPath string, newIndex int) error {
	paths := pm.GetPathStrings()
	currentIndex := pm.IndexOf(targetPath)

	if currentIndex == -1 {
		return fmt.Errorf("路径 '%s' 不存在于 PATH 中", targetPath)
	}

	if newIndex < 0 || newIndex >= len(paths) {
		return fmt.Errorf("索引 %d 超出范围，PATH 中有 %d 个路径", newIndex, len(paths))
	}

	if currentIndex == newIndex {
		return nil // 已经在目标位置
	}

	// 移除原路径
	newPaths := make([]string, 0, len(paths))
	for i, path := range paths {
		if i != currentIndex {
			newPaths = append(newPaths, path)
		}
	}

	// 在新位置插入
	if newIndex > currentIndex {
		newIndex-- // 因为删除了一个元素，索引需要调整
	}

	result := make([]string, 0, len(paths))
	result = append(result, newPaths[:newIndex]...)
	result = append(result, targetPath)
	result = append(result, newPaths[newIndex:]...)

	err := pm.setPATH(strings.Join(result, pm.pathSeparator))
	if err == nil {
		pm.invalidateCache()
	}
	return err
}

// Swap 交换两个路径的位置
func (pm *PathManager) Swap(index1, index2 int) error {
	paths := pm.GetPathStrings()

	if index1 < 0 || index1 >= len(paths) || index2 < 0 || index2 >= len(paths) {
		return fmt.Errorf("索引超出范围，PATH 中有 %d 个路径", len(paths))
	}

	if index1 == index2 {
		return nil // 相同位置，不需要交换
	}

	// 交换位置
	paths[index1], paths[index2] = paths[index2], paths[index1]

	err := pm.setPATH(strings.Join(paths, pm.pathSeparator))
	if err == nil {
		pm.invalidateCache()
	}
	return err
}

// RefreshCache 刷新缓存，强制重新读取PATH
func (pm *PathManager) RefreshCache() {
	pm.invalidateCache()
}

// GetShellType 获取检测到的 shell 类型
func (pm *PathManager) GetShellType() string {
	return pm.shellType
}

// GetConfigFile 获取配置文件路径
func (pm *PathManager) GetConfigFile() string {
	return pm.configFile
}

// GetPathSeparator 获取路径分隔符
func (pm *PathManager) GetPathSeparator() string {
	return pm.pathSeparator
}

// normalizePath 标准化路径以便比较
func (pm *PathManager) normalizePath(path string) string {
	// 展开 ~ 为用户家目录
	if strings.HasPrefix(path, "~/") {
		homeDir, _ := os.UserHomeDir()
		path = filepath.Join(homeDir, path[2:])
	}

	// 清理路径
	path = filepath.Clean(path)

	// Windows 下转换为小写以进行大小写不敏感的比较
	if runtime.GOOS == "windows" {
		path = strings.ToLower(path)
	}

	return path
}

// addWindows 在 Windows 上添加路径到 PATH
func (pm *PathManager) addWindows(newPath, position string) error {
	// 获取当前用户的 PATH
	currentPath := pm.getUserPathFromRegistry()

	var newPathValue string
	if position == PositionPrepend {
		if currentPath != "" {
			newPathValue = newPath + pm.pathSeparator + currentPath
		} else {
			newPathValue = newPath
		}
	} else {
		if currentPath != "" {
			newPathValue = currentPath + pm.pathSeparator + newPath
		} else {
			newPathValue = newPath
		}
	}

	return pm.setWindowsPATH(newPathValue)
}

// setWindowsPATH 设置 Windows PATH
func (pm *PathManager) setWindowsPATH(pathValue string) error {
	cmd := exec.Command("setx", "PATH", pathValue)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	return cmd.Run()
}

// removeWindows 在 Windows 上从 PATH 移除路径
func (pm *PathManager) removeWindows(targetPath string) error {
	// 获取当前路径列表
	paths := pm.GetPathStrings()
	targetPath = pm.normalizePath(targetPath)

	var newPaths []string
	for _, path := range paths {
		if pm.normalizePath(path) != targetPath {
			newPaths = append(newPaths, path)
		}
	}

	newPathValue := strings.Join(newPaths, pm.pathSeparator)
	return pm.setWindowsPATH(newPathValue)
}

// addUnix 在 Unix 系统上添加路径到 PATH
func (pm *PathManager) addUnix(newPath, position string) error {
	// 创建导出语句
	var exportLine string
	switch pm.shellType {
	case FISH:
		if position == PositionPrepend {
			exportLine = fmt.Sprintf("set -gx PATH %s $PATH", pm.quoteValue(newPath))
		} else {
			exportLine = fmt.Sprintf("set -gx PATH $PATH %s", pm.quoteValue(newPath))
		}
	default: // BASH, ZSH
		if position == PositionPrepend {
			exportLine = fmt.Sprintf("export PATH=%s:$PATH", pm.quoteValue(newPath))
		} else {
			exportLine = fmt.Sprintf("export PATH=$PATH:%s", pm.quoteValue(newPath))
		}
	}

	return pm.appendToConfigFile(exportLine)
}

// removeUnix 在 Unix 系统上从 PATH 移除路径
func (pm *PathManager) removeUnix(targetPath string) error {
	content, err := os.ReadFile(pm.configFile)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // 文件不存在，认为删除成功
		}
		return err
	}

	lines := strings.Split(string(content), "\n")
	var newLines []string
	targetPath = regexp.QuoteMeta(targetPath)

	// 创建匹配模式
	var patterns []*regexp.Regexp
	switch pm.shellType {
	case FISH:
		// 匹配 fish 的 PATH 设置
		pattern := regexp.MustCompile(fmt.Sprintf(`.*PATH.*%s.*`, targetPath))
		patterns = append(patterns, pattern)
	default: // BASH, ZSH
		// 匹配包含目标路径的 PATH 导出语句
		pattern := regexp.MustCompile(fmt.Sprintf(`.*PATH.*%s.*`, targetPath))
		patterns = append(patterns, pattern)
	}

	// 过滤掉匹配的行
	for _, line := range lines {
		shouldKeep := true
		for _, pattern := range patterns {
			if pattern.MatchString(line) {
				shouldKeep = false
				break
			}
		}
		if shouldKeep {
			newLines = append(newLines, line)
		}
	}

	// 写回文件
	return os.WriteFile(pm.configFile, []byte(strings.Join(newLines, "\n")), 0644)
}

// setPATH 直接设置整个 PATH 值
func (pm *PathManager) setPATH(pathValue string) error {
	if runtime.GOOS == "windows" {
		return pm.setWindowsPATH(pathValue)
	}

	// Unix 系统：这里需要更复杂的逻辑来替换现有的 PATH 设置
	// 为简化，我们先删除所有现有的 PATH 设置，然后添加新的
	pm.removeAllPathExports()

	var exportLine string
	switch pm.shellType {
	case FISH:
		exportLine = fmt.Sprintf("set -gx PATH %s", pm.quoteValue(pathValue))
	default: // BASH, ZSH
		exportLine = fmt.Sprintf("export PATH=%s", pm.quoteValue(pathValue))
	}

	return pm.appendToConfigFile(exportLine)
}

// removeAllPathExports 移除配置文件中所有的 PATH 导出语句
func (pm *PathManager) removeAllPathExports() error {
	content, err := os.ReadFile(pm.configFile)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	lines := strings.Split(string(content), "\n")
	var newLines []string

	for _, line := range lines {
		// 跳过包含 PATH 设置的行
		trimmed := strings.TrimSpace(line)
		if pm.shellType == FISH {
			if strings.Contains(trimmed, "set") && strings.Contains(trimmed, "PATH") {
				continue
			}
		} else {
			if strings.Contains(trimmed, "export") && strings.Contains(trimmed, "PATH") {
				continue
			}
		}
		newLines = append(newLines, line)
	}

	return os.WriteFile(pm.configFile, []byte(strings.Join(newLines, "\n")), 0644)
}

// appendToConfigFile 向配置文件追加内容
func (pm *PathManager) appendToConfigFile(line string) error {
	// 确保目录存在
	dir := filepath.Dir(pm.configFile)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	file, err := os.OpenFile(pm.configFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	// 检查文件是否以换行符结尾
	stat, err := file.Stat()
	if err != nil {
		return err
	}

	if stat.Size() > 0 {
		// 读取最后一个字符
		file.Seek(-1, 2)
		lastChar := make([]byte, 1)
		file.Read(lastChar)
		if lastChar[0] != '\n' {
			if _, err := file.WriteString("\n"); err != nil {
				return err
			}
		}
	}

	_, err = file.WriteString(line + "\n")
	return err
}

// quoteValue 为值添加引号（如果需要）
func (pm *PathManager) quoteValue(value string) string {
	// 如果值包含空格或特殊字符，需要加引号
	if strings.ContainsAny(value, " \t\n\r\"'\\$`") {
		return fmt.Sprintf("\"%s\"", strings.ReplaceAll(value, "\"", "\\\""))
	}
	return value
}
