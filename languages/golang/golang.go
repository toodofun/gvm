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

package golang

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/toodofun/gvm/internal/core"
	"github.com/toodofun/gvm/internal/http"
	"github.com/toodofun/gvm/internal/log"
	"github.com/toodofun/gvm/internal/util/compress"
	"github.com/toodofun/gvm/internal/util/env"
	"github.com/toodofun/gvm/internal/util/path"
	"github.com/toodofun/gvm/internal/util/slice"
	"github.com/toodofun/gvm/languages"

	goversion "github.com/hashicorp/go-version"
)

const (
	lang    = "go"
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

func (g *Golang) ListRemoteVersions(ctx context.Context) ([]*core.RemoteVersion, error) {
	logger := log.GetLogger(ctx)
	res := make([]*core.RemoteVersion, 0)
	body, err := http.Default().Get(ctx, fmt.Sprintf("%s?mode=json&include=all", baseUrl))
	if err != nil {
		logger.Errorf("Get remote versions error: %s", err.Error())
		return nil, err
	}

	versions := make([]Version, 0)
	if err = json.Unmarshal(body, &versions); err != nil {
		return nil, err
	}

	for _, v := range versions {
		comment := ""
		if v.Stable {
			comment = "Stable Release"
		}
		ver, err := goversion.NewVersion(strings.TrimPrefix(v.Version, "go"))
		if err != nil {
			logger.Warnf("Failed to parse version %s: %s", v.Version, err)
			continue
		}
		res = append(res, &core.RemoteVersion{
			Version: ver,
			Origin:  v.Version,
			Comment: comment,
		})
	}

	slice.ReverseSlice(res)

	return res, nil
}

func (g *Golang) ListInstalledVersions(ctx context.Context) ([]*core.InstalledVersion, error) {
	return languages.NewLanguage(g).ListInstalledVersions(ctx, filepath.Join("go", "bin"))
}

func (g *Golang) SetDefaultVersion(ctx context.Context, version string) error {
	gopath := filepath.Join(path.GetLangRoot(g.Name()), "gopath")
	_ = os.MkdirAll(gopath, os.ModePerm)
	envs := []env.KV{
		{
			Key:    "PATH",
			Value:  filepath.Join(path.GetLangRoot(g.Name()), path.Current, "go", "bin"),
			Append: true,
		},
		{
			Key:   "GOROOT",
			Value: filepath.Join(path.GetLangRoot(g.Name()), path.Current, "go"),
		},
		{
			Key:   "GOPATH",
			Value: gopath,
		},
	}
	return languages.NewLanguage(g).SetDefaultVersion(ctx, version, envs)
}

func (g *Golang) GetDefaultVersion(ctx context.Context) *core.InstalledVersion {
	return languages.NewLanguage(g).GetDefaultVersion()
}

func (g *Golang) Uninstall(ctx context.Context, version string) error {
	return languages.NewLanguage(g).Uninstall(version)
}

func (g *Golang) Install(ctx context.Context, version *core.RemoteVersion) error {
	logger := log.GetLogger(ctx)
	logger.Debugf("Install remote version: %s", version.Origin)
	if err, exist := languages.HasInstall(ctx, g, *version.Version); err != nil || exist {
		return err
	}
	logger.Infof("Installing version %s", version.Version.String())
	// 检查版本是否存在
	url := fmt.Sprintf("%s%s.%s-%s.tar.gz", baseUrl, version.Origin, runtime.GOOS, runtime.GOARCH)
	if runtime.GOOS == "windows" {
		url = fmt.Sprintf("%s%s.%s-%s.zip", baseUrl, version.Origin, runtime.GOOS, runtime.GOARCH)
	}
	head, code, err := http.Default().Head(ctx, url)
	if err != nil {
		return err
	}
	if runtime.GOOS == "darwin" && code == 404 {
		logger.Infof(
			"Version %s not found for %s/%s, trying %s/amd64",
			version.Version.String(),
			runtime.GOOS,
			runtime.GOARCH,
			runtime.GOOS,
		)
		// macOS 上的版本可能需要特殊处理
		url = fmt.Sprintf("%s%s.%s-%s.tar.gz", baseUrl, version.Origin, runtime.GOOS, "amd64")
		head, code, err = http.Default().Head(ctx, url)
		if err != nil {
			return err
		}
	}

	if code != 200 {
		return fmt.Errorf("version %s not found at %s, status code: %d", version, url, code)
	}

	logger.Infof("Downloading: %s, size: %s", url, head.Get("Content-Length"))
	file, err := http.Default().
		Download(ctx, url, filepath.Join(path.GetLangRoot(lang), version.Version.String()), fmt.Sprintf("%s.%s-%s.tar.gz", version.Origin, runtime.GOOS, "amd64"))
	logger.Infof("")
	if err != nil {
		return fmt.Errorf("failed to download version %s: %w", version.Version.String(), err)
	}
	logger.Infof("Extracting: %s, size: %s", url, head.Get("Content-Length"))
	if strings.HasSuffix(url, ".tar.gz") {
		if err := compress.UnTarGz(ctx, file, filepath.Join(core.GetRootDir(), "go", version.Version.String())); err != nil {
			logger.Warnf("Failed to untar version %s: %s", version.Version.String(), err)
			return fmt.Errorf("failed to extract version %s: %w", version.Version.String(), err)
		}
	} else if strings.HasSuffix(url, ".zip") {
		if err := compress.UnZip(ctx, file, filepath.Join(core.GetRootDir(), "go", version.Version.String())); err != nil {
			logger.Warnf("Failed to untar version %s: %s", version.Version.String(), err)
			return fmt.Errorf("failed to extract version %s: %w", version.Version.String(), err)
		}
	}

	if err = os.RemoveAll(file); err != nil {
		logger.Warnf("Failed to clean %s: %v", file, err)
	}

	logger.Infof(
		"Version %s was successfully installed in %s",
		version.Version.String(),
		filepath.Join(path.GetLangRoot(lang), version.Version.String(), "go", "bin"),
	)
	return nil
}

func init() {
	core.RegisterLanguage(&Golang{})
}
