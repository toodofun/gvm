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
	"bytes"
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// mockLogger 用于测试的模拟日志器
type mockLogger struct {
	messages []string
}

func (m *mockLogger) Infof(format string, args ...interface{}) {
	m.messages = append(m.messages, fmt.Sprintf(format, args...))
}

func (m *mockLogger) Tracef(format string, args ...interface{}) {}
func (m *mockLogger) Debugf(format string, args ...interface{}) {}
func (m *mockLogger) Warnf(format string, args ...interface{})  {}
func (m *mockLogger) Errorf(format string, args ...interface{}) {}
func (m *mockLogger) Fatalf(format string, args ...interface{}) {}
func (m *mockLogger) Panicf(format string, args ...interface{}) {}
func (m *mockLogger) Trace(args ...interface{})                 {}
func (m *mockLogger) Debug(args ...interface{})                 {}
func (m *mockLogger) Info(args ...interface{})                  {}
func (m *mockLogger) Warn(args ...interface{})                  {}
func (m *mockLogger) Error(args ...interface{})                 {}
func (m *mockLogger) Fatal(args ...interface{})                 {}
func (m *mockLogger) Panic(args ...interface{})                 {}

func TestGUIFilterWriter_shouldDisplayInGUI(t *testing.T) {
	buf := &bytes.Buffer{}
	mockLog := &mockLogger{}
	filter := &GUIFilterWriter{
		underlying:   buf,
		logger:       mockLog,
		isGUI:        true,
		messageCount: make(map[string]int),
	}

	tests := []struct {
		name     string
		line     string
		expected bool
	}{
		{
			name:     "configure message",
			line:     "configure: creating Makefile",
			expected: true,
		},
		{
			name:     "checking message",
			line:     "checking for gcc... gcc",
			expected: true,
		},
		{
			name:     "compiling message",
			line:     "compiling main.c",
			expected: true,
		},
		{
			name:     "error message",
			line:     "error: undefined symbol",
			expected: true,
		},
		{
			name:     "clang version (should skip)",
			line:     "Apple clang version 14.0.0",
			expected: false,
		},
		{
			name:     "rbconfig warning (should skip)",
			line:     "rbconfig.rb:21: warning: Insecure world writable dir",
			expected: false,
		},
		{
			name:     "empty line",
			line:     "",
			expected: false,
		},
		{
			name:     "whitespace only",
			line:     "   ",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := filter.shouldDisplayInGUI(tt.line)
			assert.Equal(t, tt.expected, result, "Line: %s", tt.line)
		})
	}
}

func TestGUIFilterWriter_convertToFriendlyMessage(t *testing.T) {
	buf := &bytes.Buffer{}
	mockLog := &mockLogger{}
	filter := &GUIFilterWriter{
		underlying:   buf,
		logger:       mockLog,
		isGUI:        true,
		messageCount: make(map[string]int),
	}

	tests := []struct {
		name     string
		line     string
		expected string
	}{
		{
			name:     "configure creating",
			line:     "configure: creating Makefile",
			expected: "⚙️ 正在生成配置文件...",
		},
		{
			name:     "compiling",
			line:     "compiling main.c",
			expected: "🔨 正在编译源代码...",
		},
		{
			name:     "linking",
			line:     "linking ruby",
			expected: "🔗 正在链接程序...",
		},
		{
			name:     "error",
			line:     "error: compilation failed",
			expected: "❌ error: compilation failed",
		},
		{
			name:     "warning",
			line:     "warning: deprecated function",
			expected: "⚠️ warning: deprecated function",
		},
		{
			name:     "success",
			line:     "installation completed successfully",
			expected: "✅ installation completed successfully",
		},
		{
			name:     "irrelevant line",
			line:     "some random output",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := filter.convertToFriendlyMessage(tt.line)
			assert.Equal(t, tt.expected, result, "Line: %s", tt.line)
		})
	}
}

func TestGUIFilterWriter_Write(t *testing.T) {
	buf := &bytes.Buffer{}
	mockLog := &mockLogger{}
	filter := &GUIFilterWriter{
		underlying:   buf,
		logger:       mockLog,
		isGUI:        true,
		messageCount: make(map[string]int),
	}

	testInput := `configure: creating Makefile
Apple clang version 14.0.0
compiling main.c
rbconfig.rb:21: warning: Insecure world writable dir
linking ruby
installation completed successfully`

	n, err := filter.Write([]byte(testInput))
	assert.NoError(t, err)
	assert.Equal(t, len(testInput), n)

	// 在GUI模式下，底层writer不应该收到原始输出
	assert.Empty(t, buf.String())

	// 检查mock logger只收到了友好信息（过滤掉了rbconfig警告）
	assert.Len(t, mockLog.messages, 4) // configure, compiling, linking, success
	expectedMessages := []string{
		"⚙️ 正在生成配置文件...",
		"🔨 正在编译源代码...",
		"🔗 正在链接程序...",
		"✅ installation completed successfully",
	}

	for i, expected := range expectedMessages {
		if i < len(mockLog.messages) {
			assert.Contains(t, mockLog.messages[i], strings.Split(expected, " ")[0]) // 检查emoji部分
		}
	}
}

func TestNewGUIFilterWriter(t *testing.T) {
	ctx := context.Background()
	logger := &mockLogger{}

	writer := NewGUIFilterWriter(ctx, logger, true)
	assert.NotNil(t, writer)

	// 测试写入
	_, err := writer.Write([]byte("configure: test"))
	assert.NoError(t, err)
}

func TestGetFilteredStdout(t *testing.T) {
	ctx := context.Background()
	writer := GetFilteredStdout(ctx)
	assert.NotNil(t, writer)
}

func TestGetFilteredStderr(t *testing.T) {
	ctx := context.Background()
	writer := GetFilteredStderr(ctx)
	assert.NotNil(t, writer)
}
