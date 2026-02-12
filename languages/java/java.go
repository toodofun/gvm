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

package java

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/toodofun/gvm/i18n"
	"github.com/toodofun/gvm/internal/core"
	"github.com/toodofun/gvm/internal/http"
	"github.com/toodofun/gvm/internal/log"
	"github.com/toodofun/gvm/internal/util/compress"
	"github.com/toodofun/gvm/internal/util/env"
	"github.com/toodofun/gvm/internal/util/path"
	"github.com/toodofun/gvm/languages"
)

const (
	lang    = "java"
	zuluUrl = "https://api.azul.com/metadata/v1/zulu/packages"
)

type Java struct{}

func (j *Java) Name() string {
	return lang
}

func (j *Java) ListRemoteVersions(ctx context.Context) ([]*core.RemoteVersion, error) {
	logger := log.GetLogger(ctx)
	res := make([]*core.RemoteVersion, 0)

	page := 1
	pageSize := 1000
	for {
		more, err := fetchRemote(ctx, page, pageSize, func(version *core.RemoteVersion) {
			logger.Debugf("Fetching remote version %+v", version)
			res = append(res, version)
		})
		if err != nil {
			return nil, err
		}
		if !more {
			break
		}
		page++
	}
	return res, nil
}

func (j *Java) ListInstalledVersions(ctx context.Context) ([]*core.InstalledVersion, error) {
	return languages.NewLanguage(j).ListInstalledVersions(ctx, "bin")
}

func (j *Java) SetDefaultVersion(ctx context.Context, version string) error {
	envs := []env.KV{
		{
			Key:    "PATH",
			Value:  filepath.Join(path.GetLangRoot(lang), path.Current, "bin"),
			Append: true,
		},
	}
	return languages.NewLanguage(j).SetDefaultVersion(ctx, version, envs)
}

func (j *Java) GetDefaultVersion(ctx context.Context) *core.InstalledVersion {
	return languages.NewLanguage(j).GetDefaultVersion()
}

func (j *Java) Uninstall(ctx context.Context, version string) error {
	return languages.NewLanguage(j).Uninstall(version)
}

func (j *Java) Install(ctx context.Context, version *core.RemoteVersion) error {
	logger := log.GetLogger(ctx)
	logger.Debugf("Install version: %+v", version)
	logger.Infof("🐹 %s", i18n.GetTranslate("languages.startInstall", map[string]any{
		"lang":    lang,
		"version": version.Version.String(),
	}))
	//logger.Infof("📦 Java 使用预编译包，安装通常需要 1-3 分钟...")

	file, err := http.Default().
		Download(ctx, version.Origin, filepath.Join(path.GetLangRoot(lang), version.Version.String()), fmt.Sprintf("%s.%s-%s.tar.gz", version.Version.String(), runtime.GOOS, "amd64"))
	logger.Infof("")
	if err != nil {
		return fmt.Errorf("failed to download version: %s(%s): %w", version.Version.String(), version.Comment, err)
	}

	installDir := filepath.Join(path.GetLangRoot(lang), version.Version.String())
	logger.Infof("📁 解压 Java 安装包...")

	if err := compress.UnTarGz(ctx, file, installDir); err != nil {
		return fmt.Errorf("failed to unTarGz: %s(%s): %w", version.Version.String(), version.Comment, err)
	}

	if err = os.RemoveAll(file); err != nil {
		logger.Warnf("failed to clean %s: %v", file, err)
	}

	logger.Infof("🔧 整理 Java 安装文件...")
	dirs, err := filepath.Glob(filepath.Join(installDir, "/*"))
	if err != nil {
		logger.Errorf("failed to glob %s: %v", installDir, err)
		return err
	}
	for _, dir := range dirs {
		sourceDir := dir
		files, err := os.ReadDir(sourceDir)
		if err != nil {
			return err
		}

		for _, f := range files {
			sourcePath := filepath.Join(sourceDir, f.Name())
			destPath := filepath.Join(installDir, f.Name())

			if _, err := os.Stat(destPath); err == nil {
				logger.Warnf("%s already exists", destPath)
				continue
			}

			err := os.Rename(sourcePath, destPath)
			if err != nil {
				return err
			}
		}
	}

	logger.Infof(
		"✅ %s",
		i18n.GetTranslate("languages.installComplete", map[string]any{
			"lang":     lang,
			"version":  fmt.Sprintf("%s (%s)", version.Version.String(), version.Comment),
			"location": installDir,
		}),
	)

	return nil
}

type Version struct {
	Abi                 string        `json:"abi"`
	Arch                string        `json:"arch"`
	ArchiveType         string        `json:"archive_type"`
	AvailabilityType    string        `json:"availability_type"`
	Certifications      []string      `json:"certifications"`
	CpuGen              []interface{} `json:"cpu_gen"`
	CracSupported       bool          `json:"crac_supported"`
	DistroVersion       []int         `json:"distro_version"`
	DownloadUrl         string        `json:"download_url"`
	HwBitness           int           `json:"hw_bitness"`
	JavaPackageFeatures []string      `json:"java_package_features"`
	JavaPackageType     string        `json:"java_package_type"`
	JavaVersion         []int         `json:"java_version"`
	JavafxBundled       bool          `json:"javafx_bundled"`
	Latest              bool          `json:"latest"`
	LibCType            string        `json:"lib_c_type"`
	Name                string        `json:"name"`
	OpenjdkBuildNumber  int           `json:"openjdk_build_number"`
	Os                  string        `json:"os"`
	PackageUuid         string        `json:"package_uuid"`
	Product             string        `json:"product"`
	ReleaseStatus       string        `json:"release_status"`
	Sha256Hash          string        `json:"sha256_hash"`
	Size                int           `json:"size"`
	SupportTerm         string        `json:"support_term"`
}

func init() {
	core.RegisterLanguage(&Java{})
}
