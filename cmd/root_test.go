package cmd

import (
	"context"
	"github.com/spf13/pflag"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"gvm/internal/core"
	"gvm/internal/log"
)

func TestNewRootCmd(t *testing.T) {
	cmd := NewRootCmd()

	// 基本信息检查
	if cmd.Use != "gvm" {
		t.Errorf("expected Use to be 'gvm', got %q", cmd.Use)
	}
	if cmd.Short == "" {
		t.Error("expected Short description to be set")
	}
	if cmd.Long == "" {
		t.Error("expected Long description to be set")
	}

	// 检查 debug flag 是否注册
	foundDebugFlag := false
	cmd.PersistentFlags().VisitAll(func(f *pflag.Flag) {
		if f.Name == "debug" {
			foundDebugFlag = true
		}
	})
	if !foundDebugFlag {
		t.Error("debug flag not found in PersistentFlags")
	}

	// 测试 PersistentPreRun 对 Context 的设置
	// 模拟调用 PersistentPreRun
	debug = false // 重置全局变量
	ctx := context.Background()
	cmd.SetContext(ctx)

	// 模拟 cmd.Name() != "ui" 时的情况
	cmd.PersistentPreRun(cmd, []string{})
	newCtx := cmd.Context()
	if newCtx.Value(core.ContextLogWriterKey) == nil {
		t.Error("expected ContextLogWriterKey to be set to non-nil (os.Stdout) when command name != 'ui'")
	}

	// 模拟 cmd.Name() == "ui" 时的情况
	uiCmd := &cobra.Command{
		Use:              "ui",
		PersistentPreRun: cmd.PersistentPreRun,
	}
	uiCmd.SetContext(ctx)
	uiCmd.PersistentPreRun(uiCmd, []string{})
	uiCtx := uiCmd.Context()
	if uiCtx.Value(core.ContextLogWriterKey) != nil {
		t.Error("expected ContextLogWriterKey to be nil when command name == 'ui'")
	}

	// 测试 debug 模式下日志级别设置
	debug = true
	cmd.PersistentPreRun(cmd, []string{})
	if log.GetLevel() != logrus.DebugLevel.String() {
		t.Errorf("expected log level DebugLevel when debug is true, got %v", log.GetLevel())
	}

	expectedSubCommands := []string{
		"ls-remote",
		"ls",
		"use",
		"install",
		"uninstall",
		"current",
		"ui",
	}

	if len(cmd.Commands()) != len(expectedSubCommands) {
		t.Errorf("expected %d subcommands, got %d", len(expectedSubCommands), len(cmd.Commands()))
	}

	for _, name := range expectedSubCommands {
		found := false
		for _, c := range cmd.Commands() {
			if c.Name() == name {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected subcommand %q to be added", name)
		}
	}
}
