package core

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

type mockLanguage struct {
	name string
}

func (m *mockLanguage) ListRemoteVersions(ctx context.Context) ([]*RemoteVersion, error) {
	//TODO implement me
	panic("implement me")
}

func (m *mockLanguage) ListInstalledVersions(ctx context.Context) ([]*InstalledVersion, error) {
	//TODO implement me
	panic("implement me")
}

func (m *mockLanguage) SetDefaultVersion(ctx context.Context, version string) error {
	//TODO implement me
	panic("implement me")
}

func (m *mockLanguage) GetDefaultVersion(ctx context.Context) *InstalledVersion {
	//TODO implement me
	panic("implement me")
}

func (m *mockLanguage) Install(ctx context.Context, remoteVersion *RemoteVersion) error {
	//TODO implement me
	panic("implement me")
}

func (m *mockLanguage) Uninstall(ctx context.Context, version string) error {
	//TODO implement me
	panic("implement me")
}

func (m *mockLanguage) Name() string {
	return m.name
}

func TestRegisterAndGetLanguage(t *testing.T) {
	// 清空全局变量 languages，避免测试间污染
	languages = make(map[string]Language)

	lang := &mockLanguage{name: "Go"}

	// 测试注册语言
	RegisterLanguage(lang)

	// 测试通过名称获取语言，应该成功
	gotLang, exists := GetLanguage("Go")
	assert.True(t, exists, "language should exist")
	assert.Equal(t, lang, gotLang, "got language should be the registered one")

	// 测试获取不存在的语言，应该失败
	_, exists = GetLanguage("Python")
	assert.False(t, exists, "language should not exist")
}

func TestGetAllLanguage(t *testing.T) {
	languages = make(map[string]Language)

	RegisterLanguage(&mockLanguage{name: "Java"})
	RegisterLanguage(&mockLanguage{name: "Go"})
	RegisterLanguage(&mockLanguage{name: "C++"})

	allLangs := GetAllLanguage()

	// 结果应该是排序后的语言名称切片
	expected := []string{"C++", "Go", "Java"}
	assert.Equal(t, expected, allLangs)
}
