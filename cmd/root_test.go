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

package cmd

import (
	"context"
	"io"
	"testing"

	"github.com/spf13/pflag"

	"gvm/internal/core"
	"gvm/internal/log"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
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
	if uiCtx.Value(core.ContextLogWriterKey) != io.Discard {
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
