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
	"strings"

	"github.com/toodofun/gvm/internal/core"
	gvmhttp "github.com/toodofun/gvm/internal/http"
	"github.com/toodofun/gvm/internal/log"
	"github.com/toodofun/gvm/internal/util/compress"
	"github.com/toodofun/gvm/internal/util/env"
	"github.com/toodofun/gvm/internal/util/path"
	"github.com/toodofun/gvm/internal/util/slice"
	"github.com/toodofun/gvm/languages"

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

// 获取远程Python版本列表
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
	for _, m := range matches {
		verStr := m[1]
		if _, ok := versionSet[verStr]; ok {
			continue
		}
		ver, err := goversion.NewVersion(verStr)
		if err != nil {
			continue
		}
		res = append(res, &core.RemoteVersion{
			Version: ver,
			Origin:  verStr,
			Comment: "",
		})
		versionSet[verStr] = struct{}{}
	}
	slice.ReverseSlice(res)
	return res, nil
}

func (p *Python) ListInstalledVersions(ctx context.Context) ([]*core.InstalledVersion, error) {
	return languages.NewLanguage(p).ListInstalledVersions(ctx, filepath.Join("bin", "python3"))
}

func (p *Python) SetDefaultVersion(ctx context.Context, version string) error {
	envs := []env.KV{
		{
			Key:    "PATH",
			Value:  filepath.Join(path.GetLangRoot(p.Name()), path.Current, "bin"),
			Append: true,
		},
		{
			Key:   "PYTHONHOME",
			Value: filepath.Join(path.GetLangRoot(p.Name()), path.Current),
		},
	}
	return languages.NewLanguage(p).SetDefaultVersion(ctx, version, envs)
}

func (p *Python) GetDefaultVersion(ctx context.Context) *core.InstalledVersion {
	return languages.NewLanguage(p).GetDefaultVersion()
}

func (p *Python) Uninstall(ctx context.Context, version string) error {
	return languages.NewLanguage(p).Uninstall(version)
}

func (p *Python) Install(ctx context.Context, version *core.RemoteVersion) error {
	logger := log.GetLogger(ctx)
	logger.Debugf("Install remote version: %s", version.Origin)
	if err, exist := languages.HasInstall(ctx, p, *version.Version); err != nil || exist {
		return err
	}
	logger.Infof("Installing version %s", version.Version.String())
	filename := fmt.Sprintf("Python-%s.tgz", version.Origin)
	url := fmt.Sprintf("%s%s/%s", baseUrl, version.Origin, filename)
	installRoot := filepath.Join(path.GetLangRoot(lang), version.Version.String())
	if strings.Contains(installRoot, " ") {
		return fmt.Errorf("Python 源码包不支持带空格的安装路径，请将 gvm 根目录迁移到无空格路径（如 ~/.gvm）后重试")
	}
	head, code, err := gvmhttp.Default().Head(ctx, url)
	if err != nil || code != 200 {
		logger.Warnf("Version %s not found at %s, status code: %d", version.Origin, url, code)
		return fmt.Errorf("version %s not found at %s, status code: %d", version.Origin, url, code)
	}
	logger.Infof("Downloading: %s, size: %s", url, head.Get("Content-Length"))
	file, err := gvmhttp.Default().Download(ctx, url, filepath.Join(path.GetLangRoot(lang), version.Version.String()), filename)
	if err != nil {
		return fmt.Errorf("failed to download version %s: %w", version.Version.String(), err)
	}
	logger.Infof("Extracting: %s", file)
	srcDir := filepath.Join(installRoot, fmt.Sprintf("Python-%s", version.Origin))
	if strings.HasSuffix(url, ".tgz") {
		if err := compress.UnTarGz(ctx, file, installRoot); err != nil {
			logger.Warnf("Failed to untar version %s: %s", version.Version.String(), err)
			return fmt.Errorf("failed to extract version %s: %w", version.Version.String(), err)
		}
	} else if strings.HasSuffix(url, ".zip") {
		if err := compress.UnZip(ctx, file, installRoot); err != nil {
			logger.Warnf("Failed to unzip version %s: %s", version.Version.String(), err)
			return fmt.Errorf("failed to extract version %s: %w", version.Version.String(), err)
		}
	}
	if err = os.RemoveAll(file); err != nil {
		logger.Warnf("Failed to clean %s: %v", file, err)
	}
	// 自动编译源码并安装到gvm管理目录
	logger.Infof("Building python source in %s", srcDir)
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

	cmds := []string{
		fmt.Sprintf("./configure --prefix=\"%s\"", installRoot),
		"make -j4",
		"make install",
	}
	for _, shellCmd := range cmds {
		var cmd *exec.Cmd
		if runtime.GOOS == "windows" {
			cmd = exec.Command("cmd", "/C", shellCmd)
		} else {
			cmd = exec.Command("sh", "-c", shellCmd)
		}
		cmd.Dir = srcDir
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		logger.Infof("Running: %s", shellCmd)
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to run %s: %w", shellCmd, err)
		}
	}
	logger.Infof("Version %s was successfully installed in %s", version.Version.String(), filepath.Join(installRoot, "bin"))
	return nil
}

func init() {
	core.RegisterLanguage(&Python{})
}
