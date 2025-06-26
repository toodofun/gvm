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
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestManager_quoteValue(t *testing.T) {
	m := NewEnvManager()

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"no special characters", "abc", "abc"},
		{"contains space", "abc def", "\"abc def\""},
		{"already quoted", "\"abc def\"", "\"\\\"abc def\\\"\""}, // 会被再次加引号
		{"contains dollar", "value$HOME", "\"value$HOME\""},
		{"contains quote", `a"b`, `"a\"b"`},
		{"contains path separator", "foo:bar", "foo:bar"}, // 不加引号
	}

	if runtime.GOOS == "windows" {
		t.Skip("quoteValue does nothing on Windows")
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := m.quoteValue(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}
