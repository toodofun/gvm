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

package ruby

import (
	"context"
	"fmt"
	"path/filepath"
	"runtime"
	"strings"

	goversion "github.com/hashicorp/go-version"

	"github.com/toodofun/gvm/internal/core"
	gvmhttp "github.com/toodofun/gvm/internal/http"
	"github.com/toodofun/gvm/internal/log"
	"github.com/toodofun/gvm/internal/util/env"
	"github.com/toodofun/gvm/internal/util/path"
	"github.com/toodofun/gvm/languages"
)

const (
	lang = "ruby"
)

type Ruby struct{}

func (r *Ruby) Name() string {
	return lang
}

// 获取远程Ruby版本列表 - 直接使用回退版本列表（支持更广泛的版本）
func (r *Ruby) ListRemoteVersions(ctx context.Context) ([]*core.RemoteVersion, error) {
	// 直接返回常见的 Ruby 稳定版本列表
	// 这些版本可以通过官方源码包或预编译包安装
	return r.getFallbackVersions(), nil
}

// 回退版本列表 - 使用常见的稳定版本
func (r *Ruby) getFallbackVersions() []*core.RemoteVersion {
	versions := []string{
		"3.3.6", "3.3.5", "3.3.4", "3.3.3", "3.3.2", "3.3.1", "3.3.0",
		"3.2.6", "3.2.5", "3.2.4", "3.2.3", "3.2.2", "3.2.1", "3.2.0",
		"3.1.7", "3.1.6", "3.1.5", "3.1.4", "3.1.3", "3.1.2", "3.1.1", "3.1.0",
		"3.0.7", "3.0.6", "3.0.5", "3.0.4", "3.0.3", "3.0.2", "3.0.1", "3.0.0",
	}

	res := make([]*core.RemoteVersion, 0, len(versions))
	for _, v := range versions {
		if ver, err := goversion.NewVersion(v); err == nil {
			res = append(res, &core.RemoteVersion{
				Version: ver,
				Origin:  v,
				Comment: "Stable Release",
			})
		}
	}

	return res
}

func (r *Ruby) ListInstalledVersions(ctx context.Context) ([]*core.InstalledVersion, error) {
	return languages.NewLanguage(r).ListInstalledVersions(ctx, filepath.Join("bin", "ruby"))
}

func (r *Ruby) SetDefaultVersion(ctx context.Context, version string) error {
	envs := []env.KV{
		{
			Key:    "PATH",
			Value:  filepath.Join(path.GetLangRoot(r.Name()), path.Current, "bin"),
			Append: true,
		},
		{
			Key:   "RUBY_HOME",
			Value: filepath.Join(path.GetLangRoot(r.Name()), path.Current),
		},
		{
			Key:    "GEM_PATH",
			Value:  filepath.Join(path.GetLangRoot(r.Name()), path.Current, "lib", "ruby", "gems"),
			Append: true,
		},
	}
	return languages.NewLanguage(r).SetDefaultVersion(ctx, version, envs)
}

func (r *Ruby) GetDefaultVersion(ctx context.Context) *core.InstalledVersion {
	return languages.NewLanguage(r).GetDefaultVersion()
}

func (r *Ruby) Uninstall(ctx context.Context, version string) error {
	return languages.NewLanguage(r).Uninstall(version)
}

func (r *Ruby) Install(ctx context.Context, version *core.RemoteVersion) error {
	logger := log.GetLogger(ctx)
	logger.Debugf("Install remote version: %s", version.Origin)
	if err, exist := languages.HasInstall(ctx, r, *version.Version); err != nil || exist {
		return err
	}

	versionStr := version.Origin

	logger.Infof("💎 开始安装 Ruby %s", versionStr)

	// 不同平台使用不同的安装策略
	switch runtime.GOOS {
	case "windows":
		return r.installWindowsRuby(ctx, version)
	default:
		// macOS 和 Linux 建议使用系统包管理器
		logger.Infof("📦 为了更好的兼容性和稳定性，建议使用系统包管理器安装 Ruby:")
		logger.Infof("")
		switch runtime.GOOS {
		case "darwin":
			logger.Infof("🍺 macOS 用户推荐:")
			logger.Infof("   brew install ruby")
			logger.Infof("   brew install ruby@3.3  # 指定版本")
		case "linux":
			logger.Infof("🐧 Linux 用户推荐:")
			logger.Infof("   # Ubuntu/Debian:")
			logger.Infof("   sudo apt-get update && sudo apt-get install ruby-full")
			logger.Infof("   # CentOS/RHEL:")
			logger.Infof("   sudo yum install ruby ruby-devel")
			logger.Infof("   # 或使用 rbenv:")
			logger.Infof("   curl -fsSL https://github.com/rbenv/rbenv-installer/raw/HEAD/bin/rbenv-installer | bash")
		}
		logger.Infof("")
		logger.Infof("💡 使用系统包管理器的优势:")
		logger.Infof("   • 预编译二进制包，安装速度快")
		logger.Infof("   • 自动处理系统依赖")
		logger.Infof("   • 更好的系统集成")
		logger.Infof("   • 定期安全更新")

		return fmt.Errorf("Ruby 源码编译复杂且耗时较长，建议使用上述系统包管理器安装方式")
	}
}

// Windows 平台的 Ruby 安装
func (r *Ruby) installWindowsRuby(ctx context.Context, version *core.RemoteVersion) error {
	logger := log.GetLogger(ctx)
	versionStr := version.Origin
	installRoot := filepath.Join(path.GetLangRoot(lang), version.Version.String())

	logger.Infof("📦 Windows Ruby 使用预编译包，安装通常需要 1-3 分钟...")

	// 构建下载 URL
	downloadURL, filename, err := r.getDownloadURL(versionStr)
	if err != nil {
		return fmt.Errorf("无法获取 Ruby %s 的下载链接: %w", versionStr, err)
	}

	logger.Infof("Downloading: %s", downloadURL)
	file, err := gvmhttp.Default().
		Download(ctx, downloadURL, installRoot, filename)
	if err != nil {
		return fmt.Errorf("failed to download version %s: %w", version.Version.String(), err)
	}

	// Windows 使用 .exe 安装包
	if strings.HasSuffix(filename, ".exe") {
		logger.Infof("🔧 运行 Ruby 安装程序...")
		logger.Infof("⚠️  注意: 请在弹出的安装向导中选择安装到: %s", installRoot)
		return fmt.Errorf("Windows Ruby 安装需要手动运行安装程序: %s", file)
	}

	logger.Infof("✅ Ruby %s 安装成功! 安装位置: %s", versionStr, filepath.Join(installRoot, "bin"))
	return nil
}

// 获取下载 URL - 优先使用预编译包，回退到源码包
func (r *Ruby) getDownloadURL(version string) (string, string, error) {
	switch runtime.GOOS {
	case "windows":
		// Windows 使用 RubyInstaller2 的预编译包
		filename := fmt.Sprintf("rubyinstaller-%s-1-x64.exe", version)
		downloadURL := fmt.Sprintf(
			"https://github.com/oneclick/rubyinstaller2/releases/download/RubyInstaller-%s/%s",
			version,
			filename,
		)
		return downloadURL, filename, nil
	default:
		// macOS 和 Linux 使用官方源码包（轻量级，编译快）
		filename := fmt.Sprintf("ruby-%s.tar.gz", version)
		majorMinor := strings.Join(strings.Split(version, ".")[0:2], ".")
		baseURL := fmt.Sprintf("https://cache.ruby-lang.org/pub/ruby/%s", majorMinor)
		downloadURL := fmt.Sprintf("%s/%s", baseURL, filename)
		return downloadURL, filename, nil
	}
}

func init() {
	core.RegisterLanguage(&Ruby{})
}
