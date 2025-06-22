package cmd_test

import (
	"bytes"
	"context"
	"errors"
	"gvm/cmd"
	"strings"
	"testing"

	"gvm/internal/core"
)

// 模拟 core.GetLanguage 返回的语言接口
type fakeLanguage struct {
	setDefaultVersionCalled bool
	versionPassed           string
	setDefaultVersionErr    error
}

func (f *fakeLanguage) Name() string {
	return "fake"
}

func (f *fakeLanguage) ListRemoteVersions(ctx context.Context) ([]*core.RemoteVersion, error) {
	//TODO implement me
	panic("implement me")
}

func (f *fakeLanguage) ListInstalledVersions(ctx context.Context) ([]*core.InstalledVersion, error) {
	//TODO implement me
	panic("implement me")
}

func (f *fakeLanguage) GetDefaultVersion(ctx context.Context) *core.InstalledVersion {
	//TODO implement me
	panic("implement me")
}

func (f *fakeLanguage) Install(ctx context.Context, remoteVersion *core.RemoteVersion) error {
	//TODO implement me
	panic("implement me")
}

func (f *fakeLanguage) Uninstall(ctx context.Context, version string) error {
	//TODO implement me
	panic("implement me")
}

func (f *fakeLanguage) SetDefaultVersion(ctx context.Context, version string) error {
	f.setDefaultVersionCalled = true
	f.versionPassed = version
	return f.setDefaultVersionErr
}

func TestNewUseCmd(t *testing.T) {
	origGetLanguage := core.GetLanguage
	defer func() { core.GetLanguage = origGetLanguage }()

	tests := []struct {
		name        string
		args        []string
		langExists  bool
		fakeLang    *fakeLanguage
		wantErr     bool
		wantErrMsg  string
		wantVersion string
	}{
		{
			name:       "参数不足返回错误",
			args:       []string{"golang"},
			wantErr:    true,
			wantErrMsg: "you need to provide language information",
		},
		{
			name:       "语言不存在显示帮助",
			args:       []string{"python", "3.10"},
			langExists: false,
			wantErr:    false, // Help() 返回 cobra.ErrHelp
		},
		{
			name:        "设置默认版本成功",
			args:        []string{"golang", "1.20"},
			langExists:  true,
			fakeLang:    &fakeLanguage{},
			wantErr:     false,
			wantVersion: "1.20",
		},
		{
			name:       "设置默认版本失败返回错误",
			args:       []string{"node", "18.0"},
			langExists: true,
			fakeLang: &fakeLanguage{
				setDefaultVersionErr: errors.New("set version failed"),
			},
			wantErr:    true,
			wantErrMsg: "set version failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// mock core.GetLanguage
			core.GetLanguage = func(name string) (core.Language, bool) {
				if !tt.langExists {
					return nil, false
				}
				return tt.fakeLang, true
			}

			cmdUse := cmd.NewUseCmd()
			cmdUse.SetArgs(tt.args)

			// 捕获命令输出，避免污染测试日志
			buf := new(bytes.Buffer)
			cmdUse.SetOut(buf)
			cmdUse.SetErr(buf)

			err := cmdUse.Execute()

			if tt.wantErr {
				if err == nil {
					t.Fatalf("预期错误，实际无错误")
				}
				if tt.wantErrMsg != "" && !strings.Contains(err.Error(), tt.wantErrMsg) {
					t.Fatalf("错误消息不匹配，期望包含: %q，实际: %q", tt.wantErrMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Fatalf("预期无错误，实际错误: %v", err)
				}
			}

			if tt.langExists && tt.fakeLang != nil && tt.wantVersion != "" {
				if !tt.fakeLang.setDefaultVersionCalled {
					t.Fatalf("预期调用 SetDefaultVersion，但未调用")
				}
				if tt.fakeLang.versionPassed != tt.wantVersion {
					t.Fatalf("传入版本不匹配，期望: %q，实际: %q", tt.wantVersion, tt.fakeLang.versionPassed)
				}
			}
		})
	}
}
