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
	"encoding/json"
	"fmt"
	"gvm/core"
	"gvm/internal/common"
	"gvm/internal/http"
	"gvm/languages"
	"path"
	"runtime"
	"strings"

	goversion "github.com/hashicorp/go-version"
	"github.com/sirupsen/logrus"
)

const (
	lang    = "node"
	baseUrl = "https://go.dev/dl/"
)

type Golang struct {
}

func (g *Golang) Name() string {
	return lang
}

type Version struct {
	Version string `json:"version"`
	Stable  bool   `json:"stable"`
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

func (g *Golang) ListRemoteVersions() ([]*core.RemoteVersion, error) {
	res := make([]*core.RemoteVersion, 0)
	body, err := http.Default().Get(fmt.Sprintf("%s?mode=json&include=all", baseUrl))
	if err != nil {
		return nil, err
	}

	versions := make([]Version, 0)
	if err = json.Unmarshal(body, &versions); err != nil {
		return nil, err
	}

	for _, v := range versions {
		comment := "Stable Release"
		if !v.Stable {
			comment = "Unstable Release"
		}
		vers, err := goversion.NewVersion(strings.TrimPrefix(v.Version, "go"))
		if err != nil {
			logrus.Warnf("Failed to parse version %s: %s", v.Version, err)
			continue
		}
		res = append(res, &core.RemoteVersion{
			Version: vers,
			Origin:  v.Version,
			Comment: comment,
		})
	}

	common.ReverseSlice(res)

	return res, nil
}

func (g *Golang) ListInstalledVersions() ([]*core.InstalledVersion, error) {
	installedVersions, err := common.GetInstalledVersion(lang, path.Join("go", "bin"))
	if err != nil {
		return nil, err
	}

	res := make([]*core.InstalledVersion, 0)
	for _, v := range installedVersions {
		version, err := goversion.NewVersion(strings.TrimPrefix(v, "go"))
		if err != nil {
			logrus.Warnf("Failed to parse version %s: %s", v, err)
			continue
		}
		res = append(res, &core.InstalledVersion{
			Version:  version,
			Origin:   v,
			Location: path.Join(common.GetLangRoot(lang), v),
		})
	}
	return res, nil
}

func (g *Golang) SetDefaultVersion(version string) error {
	// 检查是否已经安装
	source := path.Join(common.GetLangRoot(lang), version)
	target := path.Join(common.GetLangRoot(lang), common.Current)
	if !common.IsPathExist(source) {
		return fmt.Errorf("%s is not installed", version)
	}
	return common.SetSymlink(source, target)
}

func (g *Golang) GetDefaultVersion() *core.InstalledVersion {
	return languages.NewLanguage(g).GetDefaultVersion()
}

func (g *Golang) Uninstall(version string) error {
	return languages.NewLanguage(g).Uninstall(version)
}

func (g *Golang) Install(version *core.RemoteVersion) error {
	// 检查是否已经安装
	if common.IsPathExist(path.Join(common.GetLangRoot(lang), version.Version.String(), "go", "bin")) {
		logrus.Infof("Already installed")
		return nil
	}
	logrus.Infof("Installing version %s", version.Version.String())
	// 检查版本是否存在
	url := fmt.Sprintf("%s%s.%s-%s.tar.gz", baseUrl, version.Origin, runtime.GOOS, runtime.GOARCH)
	head, code, err := http.Default().Head(url)
	if err != nil {
		return err
	}
	if runtime.GOOS == "darwin" && code == 404 {
		logrus.Infof(
			"Version %s not found for %s/%s, trying %s/amd64",
			version.Version.String(),
			runtime.GOOS,
			runtime.GOARCH,
			runtime.GOOS,
		)
		// macOS 上的版本可能需要特殊处理
		url = fmt.Sprintf("%s%s.%s-%s.tar.gz", baseUrl, version.Origin, runtime.GOOS, "amd64")
		head, code, err = http.Default().Head(url)
		if err != nil {
			return err
		}
	}

	if code != 200 {
		return fmt.Errorf("version %s not found at %s, status code: %d", version, url, code)
	}

	logrus.Infof("Downloading: %s, size: %s", url, head.Get("Content-Length"))
	file, err := http.Default().
		Download(url, path.Join(core.GetRootDir(), "go", version.Version.String()), fmt.Sprintf("%s.%s-%s.tar.gz", version.Origin, runtime.GOOS, "amd64"))
	if err != nil {
		return fmt.Errorf("failed to download version %s: %w", version, err)
	}
	logrus.Infof("Extracting: %s, size: %s", url, head.Get("Content-Length"))
	if err := common.UnTarGz(file, path.Join(core.GetRootDir(), "go", version.Version.String())); err != nil {
		logrus.Warnf("Failed to untar version %s: %s", version, err)
		return fmt.Errorf("failed to extract version %s: %w", version, err)
	}
	logrus.Infof(
		"Version %s was successfully installed in %s",
		version.Version.String(),
		path.Join(core.GetRootDir(), "go", version.Version.String(), "go", "bin"),
	)
	return nil
}

func init() {
	core.RegisterLanguage(&Golang{})
}
