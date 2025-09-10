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

package rust

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
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
	lang              = "rust"
	githubReleasesURL = "https://api.github.com/repos/rust-lang/rust/releases"
	downloadBaseURL   = "https://static.rust-lang.org/dist/"
)

type Rust struct{}

type GitHubRelease struct {
	TagName    string `json:"tag_name"`
	Name       string `json:"name"`
	Prerelease bool   `json:"prerelease"`
	Draft      bool   `json:"draft"`
}

func (r *Rust) Name() string {
	return lang
}

// 获取远程Rust版本列表
func (r *Rust) ListRemoteVersions(ctx context.Context) ([]*core.RemoteVersion, error) {
	logger := log.GetLogger(ctx)
	res := make([]*core.RemoteVersion, 0)

	// 从 GitHub API 获取发布版本
	body, err := gvmhttp.Default().Get(ctx, githubReleasesURL+"?per_page=100")
	if err != nil {
		logger.Warnf("Failed to fetch rust versions from GitHub: %v", err)
		return res, err
	}

	var releases []GitHubRelease
	if err := json.Unmarshal(body, &releases); err != nil {
		logger.Warnf("Failed to parse GitHub releases: %v", err)
		return res, err
	}

	for _, release := range releases {
		if release.Draft {
			continue
		}

		// 移除 tag 前缀，如 "1.75.0" 而不是 "v1.75.0"
		versionStr := strings.TrimPrefix(release.TagName, "v")

		// 跳过无效版本
		if versionStr == "" {
			continue
		}

		ver, err := goversion.NewVersion(versionStr)
		if err != nil {
			logger.Debugf("Failed to parse version %s: %v", versionStr, err)
			continue
		}

		comment := "Stable Release"
		if release.Prerelease {
			comment = "Pre-release"
		}

		res = append(res, &core.RemoteVersion{
			Version: ver,
			Origin:  versionStr,
			Comment: comment,
		})
	}

	return res, nil
}

func (r *Rust) ListInstalledVersions(ctx context.Context) ([]*core.InstalledVersion, error) {
	return languages.NewLanguage(r).ListInstalledVersions(ctx, filepath.Join("bin", "rustc"))
}

func (r *Rust) SetDefaultVersion(ctx context.Context, version string) error {
	envs := []env.KV{
		{
			Key:    "PATH",
			Value:  filepath.Join(path.GetLangRoot(r.Name()), path.Current, "bin"),
			Append: true,
		},
		{
			Key:   "RUSTUP_HOME",
			Value: filepath.Join(path.GetLangRoot(r.Name()), path.Current),
		},
		{
			Key:   "CARGO_HOME",
			Value: filepath.Join(path.GetLangRoot(r.Name()), path.Current, "cargo"),
		},
	}
	return languages.NewLanguage(r).SetDefaultVersion(ctx, version, envs)
}

func (r *Rust) GetDefaultVersion(ctx context.Context) *core.InstalledVersion {
	return languages.NewLanguage(r).GetDefaultVersion()
}

func (r *Rust) Uninstall(ctx context.Context, version string) error {
	return languages.NewLanguage(r).Uninstall(version)
}

func (r *Rust) Install(ctx context.Context, version *core.RemoteVersion) error {
	logger := log.GetLogger(ctx)
	logger.Debugf("Install remote version: %s", version.Origin)
	if err, exist := languages.HasInstall(ctx, r, *version.Version); err != nil || exist {
		return err
	}
	logger.Infof("Installing version %s", version.Version.String())

	versionStr := version.Origin
	installRoot := filepath.Join(path.GetLangRoot(lang), version.Version.String())

	if strings.Contains(installRoot, " ") {
		return fmt.Errorf("Rust 安装包不支持带空格的安装路径，请将 gvm 根目录迁移到无空格路径（如 ~/.gvm）后重试")
	}

	// 构建下载 URL
	var downloadURL, filename string
	var foundFile bool

	// 根据操作系统和架构确定文件名
	osName := runtime.GOOS
	archName := runtime.GOARCH

	// 映射 Go 的 GOOS/GOARCH 到 Rust 的命名
	switch osName {
	case "darwin":
		osName = "apple-darwin"
	case "linux":
		osName = "unknown-linux-gnu"
	case "windows":
		osName = "pc-windows-msvc"
	}

	switch archName {
	case "amd64":
		archName = "x86_64"
	case "arm64":
		archName = "aarch64"
	}

	target := archName + "-" + osName

	// 尝试不同的文件格式
	possibleFiles := []string{
		fmt.Sprintf("rust-%s-%s.tar.gz", versionStr, target),
		fmt.Sprintf("rust-%s-%s.tar.xz", versionStr, target),
	}

	for _, file := range possibleFiles {
		testURL := downloadBaseURL + file
		head, code, err := gvmhttp.Default().Head(ctx, testURL)
		if err == nil && code == 200 {
			downloadURL = testURL
			filename = file
			foundFile = true
			logger.Infof("Found available file: %s, size: %s", file, head.Get("Content-Length"))
			break
		}
	}

	if !foundFile {
		return fmt.Errorf("版本 %s 未找到适合 %s 的安装包", version.Origin, target)
	}

	logger.Infof("Downloading: %s", downloadURL)
	file, err := gvmhttp.Default().
		Download(ctx, downloadURL, filepath.Join(path.GetLangRoot(lang), version.Version.String()), filename)
	if err != nil {
		return fmt.Errorf("failed to download version %s: %w", version.Version.String(), err)
	}

	logger.Infof("Extracting: %s", file)
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

	// 查找解压后的目录
	rustDir := fmt.Sprintf("rust-%s-%s", versionStr, target)
	srcDir := filepath.Join(installRoot, rustDir)

	// 检查是否有安装脚本
	installScript := filepath.Join(srcDir, "install.sh")
	if runtime.GOOS == "windows" {
		installScript = filepath.Join(srcDir, "install.bat")
	}

	if _, err := os.Stat(installScript); err == nil {
		logger.Infof("🦀 准备安装 Rust %s...", version.Version.String())
		logger.Infof("📍 安装目录: %s", installRoot)
		logger.Infof("⚠️  注意: Rust 安装可能需要 2-5 分钟，正在安装组件...")

		var cmd string
		if runtime.GOOS == "windows" {
			cmd = fmt.Sprintf("cd \"%s\" && install.bat --prefix=\"%s\"", srcDir, installRoot)
		} else {
			cmd = fmt.Sprintf("cd \"%s\" && ./install.sh --prefix=\"%s\"", srcDir, installRoot)
		}

		logger.Infof("🚀 执行 Rust 安装脚本...")
		if err := r.runCommand(ctx, cmd); err != nil {
			return fmt.Errorf("❌ Rust 安装失败: %w", err)
		}
		logger.Infof("✅ Rust 安装脚本执行完成")
	} else {
		// 如果没有安装脚本，直接复制文件
		logger.Infof("📁 复制 Rust 文件到安装目录...")
		logger.Infof("⚠️  注意: 正在复制文件，请稍候...")
		if err := r.copyRustFiles(srcDir, installRoot); err != nil {
			return fmt.Errorf("❌ 复制 Rust 文件失败: %w", err)
		}
		logger.Infof("✅ Rust 文件复制完成")
	}

	// 清理源目录
	if err := os.RemoveAll(srcDir); err != nil {
		logger.Warnf("Failed to clean source directory %s: %v", srcDir, err)
	}

	logger.Infof(
		"Version %s was successfully installed in %s",
		version.Version.String(),
		filepath.Join(installRoot, "bin"),
	)
	return nil
}

func (r *Rust) runCommand(ctx context.Context, command string) error {
	logger := log.GetLogger(ctx)

	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.CommandContext(ctx, "cmd", "/C", command)
	} else {
		cmd = exec.CommandContext(ctx, "sh", "-c", command)
	}

	cmd.Stdout = log.GetStdout(ctx)
	cmd.Stderr = log.GetStderr(ctx)

	logger.Infof("Running: %s", command)
	return cmd.Run()
}

func (r *Rust) copyRustFiles(srcDir, destDir string) error {
	// 简单的文件复制实现
	// 在实际情况下，这里应该有更复杂的逻辑来处理 Rust 的目录结构
	return filepath.Walk(srcDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(srcDir, path)
		if err != nil {
			return err
		}

		destPath := filepath.Join(destDir, relPath)

		if info.IsDir() {
			return os.MkdirAll(destPath, info.Mode())
		}

		return r.copyFile(path, destPath)
	})
}

func (r *Rust) copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	return err
}

func init() {
	core.RegisterLanguage(&Rust{})
}
