package color_test

import (
	"strings"
	"testing"

	"github.com/fatih/color"
	mycolor "gvm/internal/utils/color"
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
