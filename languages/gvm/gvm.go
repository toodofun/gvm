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

package gvm

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	goversion "github.com/hashicorp/go-version"

	"github.com/toodofun/gvm/internal/core"
	"github.com/toodofun/gvm/internal/http"
	"github.com/toodofun/gvm/internal/log"
	"github.com/toodofun/gvm/internal/util/compress"
	"github.com/toodofun/gvm/internal/util/env"
	"github.com/toodofun/gvm/internal/util/path"
	"github.com/toodofun/gvm/languages"
)

const (
	lang            = "gvm"
	apiBaseUrl      = "https://api.github.com/repos/toodofun/gvm/releases"
	downloadBaseUrl = "https://github.com/toodofun/gvm/releases/download/%s/gvm-%s-%s-%s.%s"
)

type GVM struct {
}

type Release struct {
	Name       string `json:"name"`
	Prerelease bool   `json:"prerelease"`
}

func (g *GVM) Name() string {
	return lang
}

func (g *GVM) ListRemoteVersions(ctx context.Context) ([]*core.RemoteVersion, error) {
	logger := log.GetLogger(ctx)
	res := make([]*core.RemoteVersion, 0)
	body, err := http.Default().Get(ctx, apiBaseUrl)
	if err != nil {
		logger.Errorf("Get remote versions error: %v", err)
		return nil, err
	}

	releases := make([]Release, 0)
	if err := json.Unmarshal(body, &releases); err != nil {
		logger.Errorf("Unmarshal remote versions error: %v", err)
		return nil, err
	}

	for _, release := range releases {
		comment := "Stable Release"
		if release.Prerelease {
			comment = "Prerelease"
		}

		ver, err := goversion.NewVersion(strings.TrimPrefix(release.Name, "v"))
		if err != nil {
			logger.Warnf("Failed to parse version %s: %v", release.Name, err)
			continue
		}
		res = append(res, &core.RemoteVersion{
			Version: ver,
			Origin:  release.Name,
			Comment: comment,
		})
	}
	return res, nil
}

func (g *GVM) ListInstalledVersions(ctx context.Context) ([]*core.InstalledVersion, error) {
	return languages.NewLanguage(g).ListInstalledVersions(ctx, filepath.Join())
}

func (g *GVM) SetDefaultVersion(ctx context.Context, version string) error {
	envs := []env.KV{
		{
			Key:    "PATH",
			Value:  filepath.Join(path.GetLangRoot(g.Name()), path.Current),
			Append: true,
		},
	}
	return languages.NewLanguage(g).SetDefaultVersion(ctx, version, envs)
}

func (g *GVM) GetDefaultVersion(ctx context.Context) *core.InstalledVersion {
	return languages.NewLanguage(g).GetDefaultVersion()
}

func (g *GVM) Install(ctx context.Context, remoteVersion *core.RemoteVersion) error {
	logger := log.GetLogger(ctx)
	logger.Infof("Install remote version %s", remoteVersion.Origin)
	tarType := "tar.gz"
	if runtime.GOOS == "windows" {
		tarType = "zip"
	}

	url := fmt.Sprintf(
		downloadBaseUrl,
		remoteVersion.Origin,
		remoteVersion.Origin,
		runtime.GOOS,
		runtime.GOARCH,
		tarType,
	)

	head, code, err := http.Default().Head(ctx, url)
	if err != nil {
		logger.Errorf("Head remote version error: %v", err)
		return err
	}
	if code != 200 {
		logger.Warnf("Head remote version code: %d", code)
		return fmt.Errorf("version %s not found", remoteVersion.Version.String())
	}

	logger.Infof("Downloading %s size: %s", url, head.Get("Content-Length"))
	file, err := http.Default().
		Download(ctx, url, filepath.Join(path.GetLangRoot(lang), remoteVersion.Version.String()), fmt.Sprintf("gvm-%s-%s-%s.%s", remoteVersion.Origin, runtime.GOOS, runtime.GOARCH, tarType))
	logger.Infof("")
	if err != nil {
		logger.Errorf("Download remote version error: %v", err)
		return fmt.Errorf("failed to download version %s: %w", remoteVersion.Version.String(), err)
	}

	if strings.HasSuffix(url, ".tar.gz") {
		if err := compress.UnTarGz(ctx, file, filepath.Join(path.GetLangRoot(lang), remoteVersion.Version.String())); err != nil {
			logger.Warnf("Failed to untar version %s: %s", remoteVersion.Version.String(), err)
			return fmt.Errorf("failed to extract version %s: %w", remoteVersion.Version.String(), err)
		}
	} else if strings.HasSuffix(url, ".zip") {
		if err := compress.UnZip(ctx, file, filepath.Join(path.GetLangRoot(lang), remoteVersion.Version.String())); err != nil {
			logger.Warnf("Failed to untar version %s: %s", remoteVersion.Version.String(), err)
			return fmt.Errorf("failed to extract version %s: %w", remoteVersion.Version.String(), err)
		}
	}

	if err = os.RemoveAll(file); err != nil {
		logger.Warnf("Failed to clean %s: %v", file, err)
	}

	logger.Infof(
		"Version %s was successfully installed in %s",
		remoteVersion.Version.String(),
		filepath.Join(path.GetLangRoot(lang), remoteVersion.Version.String()),
	)
	return nil
}

func (g *GVM) Uninstall(ctx context.Context, version string) error {
	return languages.NewLanguage(g).Uninstall(version)
}

func init() {
	core.RegisterLanguage(&GVM{})
}
