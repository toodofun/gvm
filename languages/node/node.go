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

package node

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"

	"github.com/toodofun/gvm/internal/core"
	"github.com/toodofun/gvm/internal/http"
	"github.com/toodofun/gvm/internal/log"
	common "github.com/toodofun/gvm/internal/util/compress"
	"github.com/toodofun/gvm/internal/util/env"
	"github.com/toodofun/gvm/internal/util/path"
	"github.com/toodofun/gvm/internal/util/slice"

	"os"
	"runtime"
	"strings"

	"github.com/toodofun/gvm/languages"

	goversion "github.com/hashicorp/go-version"
)

const (
	lang           = "node"
	defaultBaseURL = "https://nodejs.org/"
)

type Node struct {
	baseURL    string
	versionMap map[string]*Version
	baseDir    string
}

func (n *Node) Name() string {
	return lang
}

func NewNode(baseURL, baseDir string) core.Language {
	return &Node{
		baseURL:    baseURL,
		baseDir:    baseDir,
		versionMap: make(map[string]*Version),
	}
}

type Version struct {
	Version string   `json:"version"`
	Date    string   `json:"date"`
	Npm     string   `json:"npm"`
	LTS     any      `json:"lts"`
	Files   []string `json:"files"`
}

func (v *Version) ConvertToLTS() string {
	switch val := v.LTS.(type) {
	case bool:
		if val {
			return "LTS"
		}
		return ""
	case string:
		return fmt.Sprintf("LTS: %s", val)
	default:
		return ""
	}
}

func (n *Node) ListRemoteVersions(ctx context.Context) ([]*core.RemoteVersion, error) {
	logger := log.GetLogger(ctx)
	res := make([]*core.RemoteVersion, 0)
	body, err := http.Default().Get(ctx, fmt.Sprintf("%sdist/index.json", n.baseURL))
	if err != nil {
		return nil, err
	}

	versions := make([]*Version, 0)
	if err = json.Unmarshal(body, &versions); err != nil {
		return nil, err
	}

	for _, v := range versions {
		rv := &core.RemoteVersion{
			Origin:  v.Version,
			Comment: v.ConvertToLTS(),
		}
		if rv.Version, err = goversion.NewVersion(strings.TrimPrefix(v.Version, "v")); err != nil {
			logger.Warnf("Failed to parse version %s: %s", v.Version, err)
			continue
		}
		n.versionMap[v.Version] = v
		res = append(res, rv)
	}
	slice.ReverseSlice(res)
	return res, nil
}

func (n *Node) ListInstalledVersions(ctx context.Context) ([]*core.InstalledVersion, error) {
	if runtime.GOOS == "windows" {
		return languages.NewLanguage(n).ListInstalledVersions(ctx, filepath.Join(lang))
	}
	return languages.NewLanguage(n).ListInstalledVersions(ctx, filepath.Join(lang, "bin"))
}

func (n *Node) SetDefaultVersion(ctx context.Context, version string) error {
	binPath := filepath.Join(path.GetLangRoot(n.Name()), path.Current, "node", "bin")
	if runtime.GOOS == "windows" {
		binPath = filepath.Join(path.GetLangRoot(n.Name()), path.Current, "node")
	}
	envs := []env.KV{
		{
			Key:    "PATH",
			Value:  binPath,
			Append: true,
		},
	}
	return languages.NewLanguage(n).SetDefaultVersion(ctx, version, envs)
}

func (n *Node) GetDefaultVersion(ctx context.Context) *core.InstalledVersion {
	return languages.NewLanguage(n).GetDefaultVersion()
}

func (n *Node) Uninstall(ctx context.Context, version string) error {
	return languages.NewLanguage(n).Uninstall(version)
}

func (n *Node) Install(ctx context.Context, version *core.RemoteVersion) error {
	logger := log.GetLogger(ctx)
	logger.Debugf("Install remote version: %s", version.Origin)
	if err, exist := languages.HasInstall(ctx, n, *version.Version); err != nil || exist {
		return err
	}
	nodeInfo, ok := n.versionMap[version.Origin]
	if !ok {
		return fmt.Errorf("%s version not found", version.Origin)
	}
	logger.Infof("Installing version %s", version.Version.String())
	name, err := getPackageName(nodeInfo, version)
	if err != nil {
		return err
	}
	url := fmt.Sprintf("%sdist/%s/%s", n.baseURL, version.Origin, name)
	head, code, err := http.Default().Head(ctx, url)
	if err != nil {
		return err
	}
	if code != 200 {
		if runtime.GOOS == "darwin" && code == 404 {
			url = strings.ReplaceAll(url, runtime.GOARCH, "x64")
			name = strings.ReplaceAll(name, runtime.GOARCH, "x64")
			logger.Infof(
				"Version %s not found for %s/%s, trying %s/x64",
				version.Version.String(),
				runtime.GOOS,
				runtime.GOARCH,
				runtime.GOOS,
			)
			head, code, err = http.Default().Head(ctx, url)
			if err != nil {
				return err
			}
			if code != 200 {
				return fmt.Errorf("version %s not found at %s, status code: %d", version, url, code)
			}
		} else {
			return fmt.Errorf("version %s not found at %s, status code: %d", version, url, code)
		}
	}

	logger.Infof("Downloading: %s, size: %s", url, head.Get("Content-Length"))
	file, err := http.Default().
		Download(ctx, url, filepath.Join(core.GetRootDir(), lang, version.Version.String()), name)
	if err != nil {
		return fmt.Errorf("failed to download version %s: %w", version, err)
	}

	logger.Infof("Extracting: %s, size: %s", url, head.Get("Content-Length"))
	if err = unPackage(ctx, file, name, version.Version.String()); err != nil {
		return err
	}
	logger.Infof(
		"Version %s was successfully installed in %s",
		version.Version.String(),
		filepath.Join(core.GetRootDir(), lang, version.Version.String(), lang, "bin"),
	)
	return nil
}

func unPackage(ctx context.Context, file, packageName, version string) error {
	logger := log.GetLogger(ctx)
	dest := filepath.Join(core.GetRootDir(), lang, version)
	var (
		err      error
		fileInfo os.FileInfo
	)
	tagetPath := fmt.Sprintf("%s/%s", dest, languages.AllSuffix.Trim(packageName))

	//nolint:errcheck,ineffassign,staticcheck
	switch languages.AllSuffix.GetSuffix(packageName) {
	case languages.Tar:
		err = common.UnTarGz(ctx, file, dest)
	case languages.Zip:
		err = common.UnZip(ctx, file, dest)
	case languages.Pkg:
		err = common.UnPkg(file, dest)
	}
	if fileInfo, err = os.Stat(tagetPath); os.IsNotExist(err) || !fileInfo.IsDir() {
		logger.Warnf("Failed to untar version %s: %s", version, err)
		return fmt.Errorf("failed to extract version %s: %w", version, err)
	}
	newPath := fmt.Sprintf("%s/%s", dest, lang)
	err = os.Rename(tagetPath, newPath)
	if err != nil {
		logger.Warnf("Failed to untar version %s: %s", version, err)
		return fmt.Errorf("failed to extract version %s: %w", version, err)
	}

	return nil
}

func getPackageName(nodeInfo *Version, version *core.RemoteVersion) (string, error) {
	arch := runtime.GOARCH
	switch arch {
	case "amd64":
		arch = "x64"
	case "386":
		arch = "x86"
	case "arm":
		arch = "armv7l"
	default:
		break
	}
	var packageType string
	switch runtime.GOOS {
	case "linux":
		packageType = fmt.Sprintf("%s-%s.tar.gz", runtime.GOOS, arch)
	case "windows":
		packageType = fmt.Sprintf("win-%s.zip", arch)
	case "darwin":
		packageType = fmt.Sprintf("darwin-%s.tar.gz", arch)
	default:
		return "", fmt.Errorf("no supported architectures and platforms")
	}
	return fmt.Sprintf("node-%s-%s", version.Origin, packageType), nil
}

func init() {
	core.RegisterLanguage(NewNode(defaultBaseURL, core.GetRootDir()))
}
