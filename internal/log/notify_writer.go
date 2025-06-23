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
	mu       sync.RWMutex
	Updated  chan struct{}
	closed   bool
}

func NewNotifyBuffer() *NotifyBuffer {
	return &NotifyBuffer{
		Updated: make(chan struct{}, 1),
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

	// 处理回车符和换行符，只保留最后一行
	if strings.Contains(data, "\r") {
		lines := strings.Split(data, "\r")
		b.lastLine = strings.TrimSpace(lines[len(lines)-1])
	} else if strings.Contains(data, "\n") {
		lines := strings.Split(data, "\n")
		for i := len(lines) - 1; i >= 0; i-- {
			if strings.TrimSpace(lines[i]) != "" {
				b.lastLine = strings.TrimSpace(lines[i])
				break
			}
		}
	} else {
		b.lastLine = data
	}

	// 非阻塞通知
	select {
	case b.Updated <- struct{}{}:
	default:
	}

	return len(p), nil
}

func (b *NotifyBuffer) Read() string {
	b.mu.RLock()
	defer b.mu.RUnlock()

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
