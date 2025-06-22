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

package color_test

import (
	"strings"
	"testing"

	mycolor "gvm/internal/utils/color"

	"github.com/fatih/color"
)

func init() {
	color.NoColor = false // 强制启用颜色输出
}

func TestRedFont(t *testing.T) {
	result := mycolor.RedFont("hello")
	if !strings.Contains(result, "\x1b[31m") {
		t.Errorf("expected red ANSI code in output, got: %q", result)
	}
}

func TestGreenFont(t *testing.T) {
	result := mycolor.GreenFont("world")
	if !strings.Contains(result, "\x1b[32m") {
		t.Errorf("expected green ANSI code in output, got: %q", result)
	}
}

func TestBlueFont(t *testing.T) {
	result := mycolor.BlueFont("test")
	if !strings.Contains(result, "\x1b[34m") {
		t.Errorf("expected blue ANSI code in output, got: %q", result)
	}
}
