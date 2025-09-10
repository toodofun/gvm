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
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strings"

	"github.com/toodofun/gvm/internal/core"
	gvmhttp "github.com/toodofun/gvm/internal/http"
	"github.com/toodofun/gvm/internal/log"
	"github.com/toodofun/gvm/internal/util/compress"
	"github.com/toodofun/gvm/internal/util/env"
	"github.com/toodofun/gvm/internal/util/path"
	"github.com/toodofun/gvm/languages"

	goversion "github.com/hashicorp/go-version"
)

const (
	lang    = "ruby"
	baseUrl = "https://cache.ruby-lang.org/pub/ruby/"
)

type Ruby struct{}

func (r *Ruby) Name() string {
	return lang
}

// isGUIContext 检查是否在GUI环境中
func (r *Ruby) isGUIContext(ctx context.Context) bool {
	// 检查是否有NotifyBuffer，这表明是GUI环境
	if ctx.Value(core.ContextLogWriterKey) != nil {
		return true
	}
	return false
}

// getCleanEnvironment 获取清理后的环境变量，避免Ruby编译冲突
func (r *Ruby) getCleanEnvironment() []string {
	env := os.Environ()
	cleanEnv := make([]string, 0, len(env))

	for _, e := range env {
		// 过滤掉可能导致冲突的 Ruby 环境变量
		if strings.HasPrefix(e, "RUBY_HOME=") ||
			strings.HasPrefix(e, "RUBY_ROOT=") ||
			strings.HasPrefix(e, "GEM_HOME=") ||
			strings.HasPrefix(e, "GEM_PATH=") ||
			strings.HasPrefix(e, "BUNDLE_PATH=") {
			continue
		}
		cleanEnv = append(cleanEnv, e)
	}

	// 添加编译时需要的环境变量
	cleanEnv = append(cleanEnv, "LC_ALL=C")
	cleanEnv = append(cleanEnv, "LANG=C")

	return cleanEnv
}

// 获取远程Ruby版本列表
func (r *Ruby) ListRemoteVersions(ctx context.Context) ([]*core.RemoteVersion, error) {
	logger := log.GetLogger(ctx)
	res := make([]*core.RemoteVersion, 0)

	body, err := gvmhttp.Default().Get(ctx, baseUrl)
	if err != nil {
		logger.Warnf("Failed to fetch ruby versions: %v", err)
		return res, err
	}

	// 匹配版本目录，如 <a href="/pub/ruby/3.1/">3.1/</a>
	re := regexp.MustCompile(`<a href="/pub/ruby/([0-9]+\.[0-9]+(?:[a-z])?)/?">[0-9]+\.[0-9]+(?:[a-z])?/?</a>`)
	matches := re.FindAllStringSubmatch(string(body), -1)
	versionSet := make(map[string]struct{})

	// 收集所有主版本号
	versions := make([]string, 0)
	for _, m := range matches {
		verStr := m[1]
		if _, ok := versionSet[verStr]; !ok {
			versions = append(versions, verStr)
			versionSet[verStr] = struct{}{}
		}
	}

	// 对版本进行排序（从新到旧）
	sort.Slice(versions, func(i, j int) bool {
		// 处理特殊版本号如 "1.1a", "1.1b" 等
		v1Str := strings.TrimRight(versions[i], "abcdefghijklmnopqrstuvwxyz") + ".0"
		v2Str := strings.TrimRight(versions[j], "abcdefghijklmnopqrstuvwxyz") + ".0"
		v1, err1 := goversion.NewVersion(v1Str)
		v2, err2 := goversion.NewVersion(v2Str)
		if err1 != nil || err2 != nil {
			return versions[i] > versions[j] // 字符串比较作为后备
		}
		return v1.GreaterThan(v2)
	})

	// 对每个主版本，获取详细版本列表
	for _, verStr := range versions {
		versionURL := fmt.Sprintf("%s%s/", baseUrl, verStr)
		versionBody, err := gvmhttp.Default().Get(ctx, versionURL)
		if err != nil {
			logger.Debugf("Failed to fetch version directory %s: %v", verStr, err)
			continue
		}

		// 匹配具体版本文件，如 ruby-3.1.4.tar.gz
		detailPattern := `ruby-([0-9]+\.[0-9]+\.[0-9]+(?:-[a-z0-9]+)?)\.tar\.gz`
		detailRe := regexp.MustCompile(detailPattern)
		detailMatches := detailRe.FindAllStringSubmatch(string(versionBody), -1)

		// 收集唯一的详细版本
		detailVersions := make(map[string]bool)
		for _, detailMatch := range detailMatches {
			fullVersion := detailMatch[1]
			detailVersions[fullVersion] = true
		}

		// 添加详细版本
		for fullVersion := range detailVersions {
			ver, err := goversion.NewVersion(fullVersion)
			if err != nil {
				logger.Debugf("Failed to parse version %s: %v", fullVersion, err)
				continue
			}

			comment := "Stable Release"
			if strings.Contains(fullVersion, "-rc") {
				comment = "Release Candidate"
			} else if strings.Contains(fullVersion, "-preview") {
				comment = "Preview"
			} else if strings.Contains(fullVersion, "-alpha") {
				comment = "Alpha"
			} else if strings.Contains(fullVersion, "-beta") {
				comment = "Beta"
			}

			res = append(res, &core.RemoteVersion{
				Version: ver,
				Origin:  fullVersion,
				Comment: comment,
			})
		}
	}

	// 按版本号排序（从新到旧）
	sort.Slice(res, func(i, j int) bool {
		return res[i].Version.GreaterThan(res[j].Version)
	})

	return res, nil
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
	logger.Infof("Installing version %s", version.Version.String())

	// 在 macOS 上提供友好的安装提示
	if runtime.GOOS == "darwin" {
		logger.Infof("在 macOS 上安装 Ruby 需要编译工具。如果遇到问题，建议:")
		logger.Infof("1. 安装 Xcode Command Line Tools: xcode-select --install")
		logger.Infof("2. 安装 Homebrew 依赖: brew install openssl readline")
		logger.Infof("3. 或者直接使用 Homebrew: brew install ruby")
	}

	// 处理版本字符串格式
	versionStr := version.Origin
	// 获取主版本号（如 3.1.4 -> 3.1）
	parts := strings.Split(versionStr, ".")
	if len(parts) < 2 {
		return fmt.Errorf("invalid version format: %s", versionStr)
	}
	majorMinor := parts[0] + "." + parts[1]

	installRoot := filepath.Join(path.GetLangRoot(lang), version.Version.String())
	if strings.Contains(installRoot, " ") {
		return fmt.Errorf("Ruby 源码包不支持带空格的安装路径，请将 gvm 根目录迁移到无空格路径（如 ~/.gvm）后重试")
	}

	// 构建下载 URL（直接使用 cache.ruby-lang.org）
	downloadURL := fmt.Sprintf("https://cache.ruby-lang.org/pub/ruby/%s/ruby-%s.tar.gz", majorMinor, versionStr)
	filename := fmt.Sprintf("ruby-%s.tar.gz", versionStr)

	// 检查文件是否存在
	head, code, err := gvmhttp.Default().Head(ctx, downloadURL)
	if err != nil || code != 200 {
		return fmt.Errorf("版本 %s 未找到，状态码: %d", version.Origin, code)
	}

	logger.Infof("Found available file: %s, size: %s", filename, head.Get("Content-Length"))

	logger.Infof("Downloading: %s", downloadURL)
	file, err := gvmhttp.Default().
		Download(ctx, downloadURL, filepath.Join(path.GetLangRoot(lang), version.Version.String()), filename)
	if err != nil {
		return fmt.Errorf("failed to download version %s: %w", version.Version.String(), err)
	}

	logger.Infof("Extracting: %s", file)
	srcDir := filepath.Join(installRoot, fmt.Sprintf("ruby-%s", versionStr))
	if strings.HasSuffix(filename, ".tar.gz") {
		if err := compress.UnTarGz(ctx, file, installRoot); err != nil {
			logger.Warnf("Failed to untar version %s: %s", version.Version.String(), err)
			return fmt.Errorf("failed to extract version %s: %w", version.Version.String(), err)
		}
	} else if strings.HasSuffix(filename, ".tar.xz") {
		if err := compress.UnTarXz(ctx, file, installRoot); err != nil {
			logger.Warnf("Failed to untar xz version %s: %s", version.Version.String(), err)
			return fmt.Errorf("failed to extract version %s: %w", version.Version.String(), err)
		}
	}

	if err = os.RemoveAll(file); err != nil {
		logger.Warnf("Failed to clean %s: %v", file, err)
	}

	// 自动编译源码并安装到gvm管理目录
	logger.Infof("💎 准备编译 Ruby %s 源码...", version.Version.String())
	logger.Infof("📍 编译目录: %s", srcDir)
	logger.Infof("⚠️  注意: Ruby 源码编译可能需要 15-45 分钟，请耐心等待...")
	logger.Infof("💡 提示: 如果编译失败，建议使用系统包管理器: brew install ruby")

	if _, err := os.Stat(filepath.Join(srcDir, "configure")); err != nil {
		return fmt.Errorf("configure not found in %s, cannot build ruby", srcDir)
	}

	// 检查关键编译工具是否存在
	buildTools := []string{"make"}
	for _, tool := range buildTools {
		if _, err := exec.LookPath(tool); err != nil {
			return fmt.Errorf("%s 未安装，请先安装编译工具链后再试", tool)
		}
	}

	// 检查 C 编译器
	if _, err := exec.LookPath("clang"); err != nil {
		if _, err := exec.LookPath("gcc"); err != nil {
			return fmt.Errorf("未找到 C 编译器 (clang 或 gcc)，请先安装 Xcode Command Line Tools: xcode-select --install")
		}
	}

	// 检查是否有系统 Ruby 可以作为 baseruby
	var baseRuby string
	if _, err := exec.LookPath("ruby"); err == nil {
		baseRuby = "ruby"
	} else if _, err := exec.LookPath("ruby3"); err == nil {
		baseRuby = "ruby3"
	}

	// 构建配置选项
	configureFlags := fmt.Sprintf("--prefix=\"%s\" --disable-install-doc", installRoot)

	if baseRuby != "" {
		configureFlags += fmt.Sprintf(" --with-baseruby=%s", baseRuby)
	} else {
		// 如果没有系统 Ruby，提供友好的错误信息
		return fmt.Errorf("❌ Ruby 编译需要系统中已安装 Ruby 作为 baseruby。\n" +
			"请先安装系统 Ruby:\n" +
			"Ubuntu/Debian: apt-get install ruby\n" +
			"CentOS/RHEL: yum install ruby\n" +
			"macOS: brew install ruby")
	}

	if runtime.GOOS == "darwin" {
		// macOS 特定配置
		// 尝试自动查找 OpenSSL
		if _, err := os.Stat("/opt/homebrew/opt/openssl@3"); err == nil {
			configureFlags += " --with-openssl-dir=/opt/homebrew/opt/openssl@3"
		} else if _, err := os.Stat("/usr/local/opt/openssl@3"); err == nil {
			configureFlags += " --with-openssl-dir=/usr/local/opt/openssl@3"
		}
	} else if runtime.GOOS == "linux" {
		// Linux 特定配置，使用系统 OpenSSL
		configureFlags += " --enable-shared --disable-static"
	}

	cmds := []struct {
		cmd         string
		description string
		duration    string
	}{
		{
			cmd:         fmt.Sprintf("./configure %s", configureFlags),
			description: "配置 Ruby 编译环境",
			duration:    "2-5 分钟",
		},
		{
			cmd:         "make -j2",
			description: "编译 Ruby 源码和扩展",
			duration:    "15-35 分钟",
		},
		{
			cmd:         "make install",
			description: "安装 Ruby 到目标目录",
			duration:    "1-3 分钟",
		},
	}

	for i, cmdInfo := range cmds {
		logger.Infof("📝 步骤 %d/3: %s (预计耗时: %s)", i+1, cmdInfo.description, cmdInfo.duration)
		logger.Infof("🚀 执行: %s", cmdInfo.cmd)

		var cmd *exec.Cmd
		if runtime.GOOS == "windows" {
			cmd = exec.CommandContext(ctx, "cmd", "/C", cmdInfo.cmd)
		} else {
			cmd = exec.CommandContext(ctx, "sh", "-c", cmdInfo.cmd)
		}
		cmd.Dir = srcDir

		// 清理可能冲突的 Ruby 环境变量
		cmd.Env = r.getCleanEnvironment()

		// 在GUI环境下使用过滤输出，命令行环境下显示完整输出
		if r.isGUIContext(ctx) {
			cmd.Stdout = log.GetFilteredStdout(ctx)
			cmd.Stderr = log.GetFilteredStderr(ctx)
		} else {
			cmd.Stdout = log.GetStdout(ctx)
			cmd.Stderr = log.GetStderr(ctx)
		}

		if err := cmd.Run(); err != nil {
			// 提供更友好的错误信息
			if strings.Contains(cmdInfo.cmd, "configure") {
				return fmt.Errorf("❌ Ruby 配置失败。请确保已安装必要的依赖:\n"+
					"macOS: xcode-select --install && brew install openssl readline\n"+
					"错误详情: %w", err)
			} else if strings.Contains(cmdInfo.cmd, "make") {
				return fmt.Errorf("❌ Ruby 编译失败。这可能需要安装额外的系统依赖。\n"+
					"建议使用系统包管理器安装 Ruby: brew install ruby\n"+
					"错误详情: %w", err)
			}
			return fmt.Errorf("failed to run %s: %w", cmdInfo.cmd, err)
		}
		logger.Infof("✅ 步骤 %d/3 完成: %s", i+1, cmdInfo.description)
	}

	logger.Infof(
		"Version %s was successfully installed in %s",
		version.Version.String(),
		filepath.Join(installRoot, "bin"),
	)
	return nil
}

func init() {
	core.RegisterLanguage(&Ruby{})
}
