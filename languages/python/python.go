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

package python

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/toodofun/gvm/i18n"

	"github.com/toodofun/gvm/internal/core"
	gvmhttp "github.com/toodofun/gvm/internal/http"
	"github.com/toodofun/gvm/internal/log"
	"github.com/toodofun/gvm/internal/util/compress"
	"github.com/toodofun/gvm/internal/util/env"
	"github.com/toodofun/gvm/internal/util/path"
	"github.com/toodofun/gvm/languages"
	langenv "github.com/toodofun/gvm/languages/env"

	"os/exec"
	"runtime"

	goversion "github.com/hashicorp/go-version"
)

const (
	lang    = "python"
	baseUrl = "https://www.python.org/ftp/python/"
)

type Python struct{}

type Version struct {
	Version string `json:"version"`
	Files   []File `json:"files"`
}

type File struct {
	Filename string `json:"filename"`
	OS       string `json:"os"`
	Arch     string `json:"arch"`
	Version  string `json:"version"`
	SHA256   string `json:"sha256"`
	Size     int64  `json:"size"`
	Kind     string `json:"kind"`
}

func (p *Python) Name() string {
	return lang
}

// isGUIContext 检查是否在GUI环境中
func (p *Python) isGUIContext(ctx context.Context) bool {
	// 检查是否有NotifyBuffer，这表明是GUI环境
	if ctx.Value(core.ContextLogWriterKey) != nil {
		return true
	}
	return false
}

// getCleanEnvironment 获取清理后的环境变量，避免Python编译冲突
func (p *Python) getCleanEnvironment() []string {
	env := os.Environ()
	cleanEnv := make([]string, 0, len(env))

	for _, e := range env {
		// 过滤掉可能导致冲突的 Python 环境变量
		if strings.HasPrefix(e, "PYTHONHOME=") ||
			strings.HasPrefix(e, "PYTHONPATH=") ||
			strings.HasPrefix(e, "PYTHON_CONFIGURE_OPTS=") ||
			strings.HasPrefix(e, "PYTHONSTARTUP=") ||
			strings.HasPrefix(e, "PYTHONOPTIMIZE=") {
			continue
		}
		cleanEnv = append(cleanEnv, e)
	}

	// 添加编译时需要的环境变量，使用UTF-8编码
	cleanEnv = append(cleanEnv, "LC_ALL=en_US.UTF-8")
	cleanEnv = append(cleanEnv, "LANG=en_US.UTF-8")
	cleanEnv = append(cleanEnv, "LC_CTYPE=en_US.UTF-8")

	// 在 macOS 上添加额外的环境变量
	if runtime.GOOS == "darwin" {
		cleanEnv = append(cleanEnv, "MACOSX_DEPLOYMENT_TARGET=10.9")
	}

	return cleanEnv
}

// ListRemoteVersions 获取远程Python版本列表
func (p *Python) ListRemoteVersions(ctx context.Context) ([]*core.RemoteVersion, error) {
	logger := log.GetLogger(ctx)
	res := make([]*core.RemoteVersion, 0)

	body, err := gvmhttp.Default().Get(ctx, baseUrl)
	if err != nil {
		logger.Warnf("Failed to fetch python versions: %v", err)
		return res, err
	}

	// 匹配如 <a href="3.8.19/">3.8.19/</a> 这样的目录
	re := regexp.MustCompile(`<a href="([0-9]+\.[0-9]+\.[0-9]+)/">`)
	matches := re.FindAllStringSubmatch(string(body), -1)
	versionSet := make(map[string]struct{})

	// 收集所有版本
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
		v1, _ := goversion.NewVersion(versions[i])
		v2, _ := goversion.NewVersion(versions[j])
		return v1.GreaterThan(v2)
	})

	// 只对最新的5个主版本检查候选版本
	const checkLimit = 5
	checkedCount := 0

	for _, verStr := range versions {
		ver, err := goversion.NewVersion(verStr)
		if err != nil {
			continue
		}

		// 对于最新的几个版本，检查是否有候选版本
		if checkedCount < checkLimit {
			versionURL := fmt.Sprintf("%s%s/", baseUrl, verStr)
			versionBody, err := gvmhttp.Default().Get(ctx, versionURL)
			if err == nil {
				// 检查是否有稳定版本文件
				hasStableRelease := strings.Contains(string(versionBody), fmt.Sprintf("Python-%s.tgz", verStr)) ||
					strings.Contains(string(versionBody), fmt.Sprintf("Python-%s.tar.xz", verStr))

				if hasStableRelease {
					res = append(res, &core.RemoteVersion{
						Version: ver,
						Origin:  verStr,
						Comment: "Stable Release",
					})
				} else {
					// 查找候选版本
					rcPattern := fmt.Sprintf(`Python-%s(a[0-9]+|b[0-9]+|rc[0-9]+)\.tar\.(gz|xz)`, regexp.QuoteMeta(verStr))
					rcRe := regexp.MustCompile(rcPattern)
					rcMatches := rcRe.FindAllStringSubmatch(string(versionBody), -1)

					// 收集唯一的候选版本
					rcVersions := make(map[string]bool)
					for _, rcMatch := range rcMatches {
						fullVersion := verStr + rcMatch[1]
						rcVersions[fullVersion] = true
					}

					// 添加候选版本
					for fullVersion := range rcVersions {
						rcVer, err := goversion.NewVersion(fullVersion)
						if err == nil {
							comment := ""
							if strings.Contains(fullVersion, "rc") {
								comment = "Release Candidate"
							} else if strings.Contains(fullVersion, "b") {
								comment = "Beta"
							} else if strings.Contains(fullVersion, "a") {
								comment = "Alpha"
							}
							res = append(res, &core.RemoteVersion{
								Version: rcVer,
								Origin:  fullVersion,
								Comment: comment,
							})
						}
					}
				}
				checkedCount++
			} else {
				// 如果无法获取目录内容，假定是稳定版本
				res = append(res, &core.RemoteVersion{
					Version: ver,
					Origin:  verStr,
					Comment: "Stable Release",
				})
			}
		} else {
			// 对于较旧的版本，直接假定是稳定版本
			res = append(res, &core.RemoteVersion{
				Version: ver,
				Origin:  verStr,
				Comment: "Stable Release",
			})
		}
	}

	// 重新排序结果（从新到旧）
	sort.Slice(res, func(i, j int) bool {
		return res[i].Version.GreaterThan(res[j].Version)
	})

	return res, nil
}

func (p *Python) ListInstalledVersions(ctx context.Context) ([]*core.InstalledVersion, error) {
	return languages.NewLanguage(p).ListInstalledVersions(ctx, filepath.Join("bin", "python3"))
}

func (p *Python) SetDefaultVersion(ctx context.Context, version string) error {
	installPath := filepath.Join(path.GetLangRoot(p.Name()), path.Current)

	// Get critical environment variables from env package
	pythonEnvVars := langenv.GetPythonEnvVars(installPath)

	envs := []env.KV{
		{
			Key:    "PATH",
			Value:  filepath.Join(installPath, "bin"),
			Append: true,
		},
	}

	// Add PYTHONHOME (P0 critical - was commented out)
	if pythonhome, ok := pythonEnvVars["PYTHONHOME"]; ok {
		envs = append(envs, env.KV{
			Key:   "PYTHONHOME",
			Value: pythonhome,
		})
	}

	// Add PYTHONPATH
	if pythonpath, ok := pythonEnvVars["PYTHONPATH"]; ok {
		envs = append(envs, env.KV{
			Key:   "PYTHONPATH",
			Value: pythonpath,
		})
	}

	// Add PYTHONDONTWRITEBYTECODE
	if dontwrite, ok := pythonEnvVars["PYTHONDONTWRITEBYTECODE"]; ok {
		envs = append(envs, env.KV{
			Key:   "PYTHONDONTWRITEBYTECODE",
			Value: dontwrite,
		})
	}

	// Add PYTHONUNBUFFERED
	if unbuffered, ok := pythonEnvVars["PYTHONUNBUFFERED"]; ok {
		envs = append(envs, env.KV{
			Key:   "PYTHONUNBUFFERED",
			Value: unbuffered,
		})
	}

	// Add PYTHONIOENCODING
	if ioencoding, ok := pythonEnvVars["PYTHONIOENCODING"]; ok {
		envs = append(envs, env.KV{
			Key:   "PYTHONIOENCODING",
			Value: ioencoding,
		})
	}

	// Add LD_LIBRARY_PATH (platform-specific)
	{
		envs = append(envs, env.KV{
			Key:    "LD_LIBRARY_PATH",
			Value:  filepath.Join(installPath, "lib"),
			Append: true,
		})
	}

	return languages.NewLanguage(p).SetDefaultVersion(ctx, version, envs)
}

func (p *Python) GetDefaultVersion(ctx context.Context) *core.InstalledVersion {
	return languages.NewLanguage(p).GetDefaultVersion()
}

func (p *Python) Uninstall(ctx context.Context, version string) error {
	return languages.NewLanguage(p).Uninstall(version)
}

// 检查指定版本目录下的可用文件（包括候选版本）
func (p *Python) checkAvailableVersions(ctx context.Context, baseVersion string) ([]string, error) {
	logger := log.GetLogger(ctx)
	url := fmt.Sprintf("%s%s/", baseUrl, baseVersion)

	body, err := gvmhttp.Default().Get(ctx, url)
	if err != nil {
		logger.Debugf("Failed to fetch directory listing for %s: %v", baseVersion, err)
		return nil, err
	}

	// 匹配所有 Python-X.Y.Z*.tgz 文件
	pattern := fmt.Sprintf(`<a href="(Python-%s[^"]*\.tgz)"`, regexp.QuoteMeta(baseVersion))
	re := regexp.MustCompile(pattern)
	matches := re.FindAllStringSubmatch(string(body), -1)

	var versions []string
	for _, match := range matches {
		filename := match[1]
		// 提取版本号部分（去掉 Python- 前缀和 .tgz 后缀）
		version := strings.TrimPrefix(filename, "Python-")
		version = strings.TrimSuffix(version, ".tgz")
		versions = append(versions, version)
	}

	return versions, nil
}

func (p *Python) Install(ctx context.Context, version *core.RemoteVersion) error {
	logger := log.GetLogger(ctx)
	logger.Debugf("Install remote version: %s", version.Origin)
	if err, exist := languages.HasInstall(ctx, p, *version.Version); err != nil || exist {
		return err
	}
	logger.Infof("Installing version %s", version.Version.String())

	// 处理版本字符串格式（3.14.0-rc2 -> 3.14.0rc2）
	versionStr := version.Origin
	// 移除版本号中的连字符（用于alpha/beta/rc版本）
	versionStr = strings.ReplaceAll(versionStr, "-rc", "rc")
	versionStr = strings.ReplaceAll(versionStr, "-b", "b")
	versionStr = strings.ReplaceAll(versionStr, "-a", "a")

	// 尝试不同的文件格式
	possibleFiles := []string{
		fmt.Sprintf("Python-%s.tgz", versionStr),
		fmt.Sprintf("Python-%s.tar.xz", versionStr),
	}

	var downloadURL, filename string
	var foundFile bool

	installRoot := filepath.Join(path.GetLangRoot(lang), version.Version.String())
	if strings.Contains(installRoot, " ") {
		return fmt.Errorf("Python 源码包不支持带空格的安装路径，请将 gvm 根目录迁移到无空格路径（如 ~/.gvm）后重试")
	}

	// 获取基础版本号（去掉 rc/beta/alpha 后缀）用于确定目录
	baseVersion := versionStr
	if idx := strings.IndexAny(baseVersion, "abr"); idx > 0 {
		baseVersion = baseVersion[:idx]
	}

	// 尝试找到可用的文件
	for _, file := range possibleFiles {
		testURL := fmt.Sprintf("%s%s/%s", baseUrl, baseVersion, file)
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
		// 如果标准版本不存在，尝试查找可用的候选版本
		logger.Warnf("Version %s not found, checking for pre-release versions", versionStr)
		availableVersions, checkErr := p.checkAvailableVersions(ctx, baseVersion)
		if checkErr == nil && len(availableVersions) > 0 {
			return &languages.PreReleaseError{
				Language:          lang,
				RequestedVersion:  version.Origin,
				AvailableVersions: availableVersions,
			}
		}
		return fmt.Errorf("版本 %s 未找到", version.Origin)
	}
	logger.Infof("Downloading: %s", downloadURL)
	file, err := gvmhttp.Default().
		Download(ctx, downloadURL, filepath.Join(path.GetLangRoot(lang), version.Version.String()), filename)
	if err != nil {
		return fmt.Errorf("failed to download version %s: %w", version.Version.String(), err)
	}
	logger.Infof("Extracting: %s", file)
	srcDir := filepath.Join(installRoot, fmt.Sprintf("Python-%s", versionStr))
	if strings.HasSuffix(filename, ".tgz") || strings.HasSuffix(filename, ".tar.gz") {
		if err := compress.UnTarGz(ctx, file, installRoot); err != nil {
			logger.Warnf("Failed to untar version %s: %s", version.Version.String(), err)
			return fmt.Errorf("failed to extract version %s: %w", version.Version.String(), err)
		}
	} else if strings.HasSuffix(filename, ".tar.xz") {
		if err := compress.UnTarXz(ctx, file, installRoot); err != nil {
			logger.Warnf("Failed to untar xz version %s: %s", version.Version.String(), err)
			return fmt.Errorf("failed to extract version %s: %w", version.Version.String(), err)
		}
	} else if strings.HasSuffix(filename, ".zip") {
		if err := compress.UnZip(ctx, file, installRoot); err != nil {
			logger.Warnf("Failed to unzip version %s: %s", version.Version.String(), err)
			return fmt.Errorf("failed to extract version %s: %w", version.Version.String(), err)
		}
	}
	if err = os.RemoveAll(file); err != nil {
		logger.Warnf("Failed to clean %s: %v", file, err)
	}
	// 自动编译源码并安装到gvm管理目录
	logger.Infof("🔨 准备编译 Python %s 源码...", version.Version.String())
	logger.Infof("📍 编译目录: %s", srcDir)
	logger.Infof("⚠️  注意: Python 源码编译可能需要 10-30 分钟，请耐心等待...")

	if _, err := os.Stat(filepath.Join(srcDir, "configure")); err != nil {
		return fmt.Errorf("configure not found in %s, cannot build python", srcDir)
	}

	// 检查关键编译工具是否存在
	buildTools := []string{"gcc", "make"}
	for _, tool := range buildTools {
		if _, err := exec.LookPath(tool); err != nil {
			return fmt.Errorf("%s 未安装，请先安装编译工具链后再试", tool)
		}
	}

	// 构建 configure 命令，使用简单可靠的配置
	configureCmd := fmt.Sprintf("./configure --prefix=\"%s\"", installRoot)
	if runtime.GOOS == "darwin" {
		// macOS 特定配置，避免编码和依赖问题
		configureCmd += " --without-ensurepip --disable-ipv6"
		// 检查并添加 OpenSSL 路径
		if _, err := os.Stat("/opt/homebrew/opt/openssl@3"); err == nil {
			configureCmd += " --with-openssl=/opt/homebrew/opt/openssl@3"
		} else if _, err := os.Stat("/usr/local/opt/openssl@3"); err == nil {
			configureCmd += " --with-openssl=/usr/local/opt/openssl@3"
		}
	}

	cmds := []struct {
		cmd         string
		description string
		duration    string
	}{
		{
			cmd:         configureCmd,
			description: "配置编译环境",
			duration:    "1-3 分钟",
		},
		{
			cmd:         "make -j4",
			description: "编译 Python 源码",
			duration:    "10-25 分钟",
		},
		{
			cmd:         "make install",
			description: "安装编译结果",
			duration:    "1-2 分钟",
		},
	}

	for i, cmdInfo := range cmds {
		logger.Infof("📝 步骤 %d/3: %s (预计耗时: %s)", i+1, cmdInfo.description, cmdInfo.duration)
		logger.Infof("🚀 执行: %s", cmdInfo.cmd)

		var cmd *exec.Cmd
		if runtime.GOOS == env.RuntimeFromWindows {
			cmd = exec.CommandContext(ctx, "cmd", "/C", cmdInfo.cmd)
		} else {
			cmd = exec.CommandContext(ctx, "sh", "-c", cmdInfo.cmd)
		}
		cmd.Dir = srcDir

		// 清理可能冲突的 Python 环境变量
		cmd.Env = p.getCleanEnvironment()

		// 在GUI环境下使用过滤输出，命令行环境下显示完整输出
		if p.isGUIContext(ctx) {
			cmd.Stdout = log.GetFilteredStdout(ctx)
			cmd.Stderr = log.GetFilteredStderr(ctx)
		} else {
			cmd.Stdout = log.GetStdout(ctx)
			cmd.Stderr = log.GetStderr(ctx)
		}

		if err := cmd.Run(); err != nil {
			// 提供更友好的错误信息
			if strings.Contains(cmdInfo.cmd, "configure") {
				return fmt.Errorf("❌ Python 配置失败。请确保已安装必要的依赖:\n"+
					"macOS: xcode-select --install && brew install openssl readline sqlite3 xz zlib\n"+
					"错误详情: %w", err)
			} else if strings.Contains(cmdInfo.cmd, "make install") {
				return fmt.Errorf("❌ Python 安装失败。可能是权限或编码问题。\n"+
					"建议: 1) 检查安装目录权限 2) 重新运行安装 3) 使用系统包管理器: brew install python\n"+
					"错误详情: %w", err)
			} else if strings.Contains(cmdInfo.cmd, "make") {
				return fmt.Errorf("❌ Python 编译失败。这可能是由于:\n"+
					"1. 缺少系统依赖库\n"+
					"2. 编译器版本不兼容\n"+
					"3. 内存不足\n"+
					"建议使用系统包管理器: brew install python@%s\n"+
					"错误详情: %w", strings.Split(version.Version.String(), ".")[0]+"."+strings.Split(version.Version.String(), ".")[1], err)
			}
			return fmt.Errorf("failed to run %s: %w", cmdInfo.cmd, err)
		}
		logger.Infof("✅ 步骤 %d/3 完成: %s", i+1, cmdInfo.description)
	}
	logger.Infof(
		"✅ %s",
		i18n.GetTranslate("languages.installComplete", map[string]any{
			"lang":     lang,
			"version":  version.Version.String(),
			"location": filepath.Join(installRoot, "bin"),
		}),
	)
	return nil
}

func init() {
	core.RegisterLanguage(&Python{})
}
