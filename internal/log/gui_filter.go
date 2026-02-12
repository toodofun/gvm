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

package log

import (
	"bufio"
	"context"
	"io"
	"strings"
	"sync"
	"time"
)

// GUIFilterWriter 过滤编译输出，只显示重要信息给GUI用户
type GUIFilterWriter struct {
	underlying   io.Writer
	logger       ILogger
	isGUI        bool
	lastMessage  string         // 记录上一条消息，避免重复
	lastTime     time.Time      // 记录上次消息时间
	messageCount map[string]int // 记录消息类型的计数
	mu           sync.Mutex     // 保护并发访问
}

// NewGUIFilterWriter 创建一个新的GUI过滤写入器
func NewGUIFilterWriter(ctx context.Context, logger ILogger, isGUI bool) io.Writer {
	writer := GetWriter(ctx)
	return &GUIFilterWriter{
		underlying:   writer,
		logger:       logger,
		isGUI:        isGUI,
		messageCount: make(map[string]int),
	}
}

func (w *GUIFilterWriter) Write(p []byte) (int, error) {
	content := string(p)
	scanner := bufio.NewScanner(strings.NewReader(content))

	w.mu.Lock()
	defer w.mu.Unlock()

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if w.shouldDisplayInGUI(line) {
			// 转换技术信息为用户友好的信息
			friendlyMsg := w.convertToFriendlyMessage(line)
			if friendlyMsg != "" {
				// 使用消息类型来去重，而不是完全相同的消息
				msgType := w.getMessageType(friendlyMsg)
				now := time.Now()

				// 相同类型的消息，如果在3秒内已经显示过，则跳过
				if w.lastMessage != msgType || now.Sub(w.lastTime) > 3*time.Second {
					w.logger.Infof("%s", friendlyMsg)
					w.lastMessage = msgType
					w.lastTime = now
					w.messageCount[msgType]++
				}
			}
		}
	}

	// 如果是GUI环境，不写入原始内容到底层writer，避免显示技术细节
	// 如果不是GUI环境，写入完整内容用于日志文件
	if !w.isGUI {
		return w.underlying.Write(p)
	}

	// GUI环境下只返回长度，表示"写入成功"但不实际输出原始内容
	return len(p), nil
}

// shouldDisplayInGUI 判断是否应该在GUI中显示这行内容
func (w *GUIFilterWriter) shouldDisplayInGUI(line string) bool {
	if line == "" {
		return false
	}

	lineLower := strings.ToLower(line)

	// 首先过滤掉纯技术细节（优先级最高）
	skipKeywords := []string{
		"clang version",
		"target:",
		"thread model:",
		"installeddir:",
		"/usr/bin/ruby",
		"rbconfig.rb",
		"insecure world writable",
		"apple clang version",
		"checking build system type",
		"checking host system type",
		"warning: insecure world writable",
	}

	for _, keyword := range skipKeywords {
		if strings.Contains(lineLower, keyword) {
			return false
		}
	}

	// 特殊处理：过滤掉包含insecure的警告
	if strings.Contains(lineLower, "warning:") && strings.Contains(lineLower, "insecure") {
		return false
	}

	// 显示重要的进度信息
	importantKeywords := []string{
		"configure:",
		"checking for",
		"creating",
		"installing",
		"building",
		"linking",
		"generating",
		"compiling",
		"error:",
		"warning:",
		"failed",
		"success",
		"completed",
		"finished",
	}

	for _, keyword := range importantKeywords {
		if strings.Contains(lineLower, keyword) {
			return true
		}
	}

	return false
}

// convertToFriendlyMessage 将技术信息转换为用户友好的信息
func (w *GUIFilterWriter) convertToFriendlyMessage(line string) string {
	lineLower := strings.ToLower(line)

	// 配置阶段
	if strings.Contains(lineLower, "configure:") {
		if strings.Contains(lineLower, "creating") {
			return "⚙️ 正在生成配置文件..."
		}
		if strings.Contains(lineLower, "error") {
			return "❌ 配置过程中遇到错误"
		}
		return "⚙️ 正在配置编译环境..."
	}

	// 编译阶段 - 使用通用消息避免重复
	if strings.Contains(lineLower, "compiling") {
		// 检查是否是编译进度的开始
		if strings.Contains(line, ".c") || strings.Contains(line, ".py") || strings.Contains(line, "main") {
			return "🔨 正在编译源代码..."
		}
		return "" // 其他编译信息不显示，避免重复
	}

	if strings.Contains(lineLower, "linking") {
		return "🔗 正在链接程序..."
	}

	if strings.Contains(lineLower, "building") {
		return "🏗️ 正在构建..."
	}

	if strings.Contains(lineLower, "installing") {
		return "📦 正在安装文件..."
	}

	if strings.Contains(lineLower, "generating") {
		return "📝 正在生成文件..."
	}

	// 检查重要的检查步骤
	if strings.Contains(lineLower, "checking for") {
		if strings.Contains(lineLower, "gcc") || strings.Contains(lineLower, "clang") {
			return "🔧 检查编译工具..."
		}
		if strings.Contains(lineLower, "make") {
			return "🔧 检查构建工具..."
		}
		// 其他检查步骤不显示，避免过多信息
		return ""
	}

	// 错误和警告
	if strings.Contains(lineLower, "error") {
		return "❌ " + line
	}

	if strings.Contains(lineLower, "warning") && !strings.Contains(lineLower, "insecure") {
		return "⚠️ " + line
	}

	// 成功信息
	if strings.Contains(lineLower, "success") || strings.Contains(lineLower, "completed") {
		return "✅ " + line
	}

	return ""
}

// getMessageType 获取消息类型，用于去重
func (w *GUIFilterWriter) getMessageType(msg string) string {
	if strings.Contains(msg, "🔨") {
		return "compiling"
	}
	if strings.Contains(msg, "🔗") {
		return "linking"
	}
	if strings.Contains(msg, "⚙️") {
		return "configuring"
	}
	if strings.Contains(msg, "📝") {
		return "generating"
	}
	if strings.Contains(msg, "📦") {
		return "installing"
	}
	if strings.Contains(msg, "🔧") {
		return "checking"
	}
	if strings.Contains(msg, "❌") {
		return "error"
	}
	if strings.Contains(msg, "⚠️") {
		return "warning"
	}
	if strings.Contains(msg, "✅") {
		return "success"
	}
	return "other"
}

// GetFilteredStdout 获取过滤后的标准输出（用于GUI）
func GetFilteredStdout(ctx context.Context) io.Writer {
	logger := GetLogger(ctx)
	return NewGUIFilterWriter(ctx, logger, true)
}

// GetFilteredStderr 获取过滤后的标准错误输出（用于GUI）
func GetFilteredStderr(ctx context.Context) io.Writer {
	logger := GetLogger(ctx)
	return NewGUIFilterWriter(ctx, logger, true)
}
