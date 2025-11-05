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

package log

import (
	"strings"
	"sync"
)

type NotifyBuffer struct {
	lastLine string
	lines    []string // 保存多行消息
	mu       sync.RWMutex
	Updated  chan struct{}
	closed   bool
	maxLines int // 最大保存行数
}

func NewNotifyBuffer() *NotifyBuffer {
	return &NotifyBuffer{
		Updated:  make(chan struct{}, 1),
		lines:    make([]string, 0),
		maxLines: 20, // 最多保存20行
	}
}

func (b *NotifyBuffer) Write(p []byte) (int, error) {
	if len(p) == 0 {
		return 0, nil
	}

	b.mu.Lock()
	defer b.mu.Unlock()

	if b.closed {
		return 0, nil
	}

	data := strings.TrimSpace(string(p))
	if data == "" {
		return len(p), nil
	}

	// 处理回车符和换行符
	if strings.Contains(data, "\r") {
		lines := strings.Split(data, "\r")
		b.lastLine = strings.TrimSpace(lines[len(lines)-1])
	} else if strings.Contains(data, "\n") {
		lines := strings.Split(data, "\n")
		for _, line := range lines {
			trimmed := strings.TrimSpace(line)
			if trimmed != "" {
				b.addLine(trimmed)
				b.lastLine = trimmed
			}
		}
	} else {
		b.addLine(data)
		b.lastLine = data
	}

	// 非阻塞通知
	select {
	case b.Updated <- struct{}{}:
	default:
	}

	return len(p), nil
}

// addLine 添加一行到缓冲区
func (b *NotifyBuffer) addLine(line string) {
	// 检查是否是重要的友好消息（包含emoji）
	if strings.ContainsAny(line, "🔨🔗⚙️📝📦📁🔧❌⚠️✅🐹🟢☕💎🦀") {
		// 检查是否是重复的消息类型
		messageType := b.getMessageType(line)

		// 查找并替换相同类型的消息，而不是累积
		found := false
		for i, existingLine := range b.lines {
			if b.getMessageType(existingLine) == messageType {
				b.lines[i] = line // 替换而不是追加
				found = true
				break
			}
		}

		if !found {
			b.lines = append(b.lines, line)
			// 保持最大行数限制
			if len(b.lines) > b.maxLines {
				b.lines = b.lines[1:]
			}
		}
	}
}

// getMessageType 获取消息类型
func (b *NotifyBuffer) getMessageType(msg string) string {
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
	if strings.Contains(msg, "📁") {
		return "extracting"
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

func (b *NotifyBuffer) Read() string {
	b.mu.RLock()
	defer b.mu.RUnlock()

	// 如果有多行友好消息，返回所有行
	if len(b.lines) > 0 {
		return strings.Join(b.lines, "\n")
	}

	return b.lastLine
}

func (b *NotifyBuffer) Close() {
	b.mu.Lock()
	defer b.mu.Unlock()

	if !b.closed {
		b.closed = true
		close(b.Updated)
	}
}
