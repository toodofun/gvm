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
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
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
	lang    = "java"
	zuluUrl = "https://api.azul.com/metadata/v1/zulu/packages"
)

type Java struct{}

func (j *Java) Name() string {
	return lang
}

func (j *Java) getCurrentSystemInfo() (os, arch, hwBitness string) {
	switch strings.ToLower(runtime.GOOS) {
	case "linux":
		os = "linux"
	case "windows":
		os = "windows"
	case "darwin":
		os = "macos"
	}

	switch strings.ToLower(runtime.GOARCH) {
	case "amd64":
		arch = "x86"
		hwBitness = "64"
	case "arm64":
		arch = "arm"
		hwBitness = "64"
		if runtime.GOOS == "darwin" {
			arch = ""
			hwBitness = ""
		}
	case "386":
		arch = "x86"
		hwBitness = "32"
	case "arm":
		arch = "arm"
		hwBitness = "32"
	}
	return
}

func (j *Java) fetchRemote(
	ctx context.Context,
	page, size int,
	callback func(version *core.RemoteVersion),
) (more bool, err error) {
	logger := log.GetLogger(ctx)
	params := url.Values{}
	params.Set("page", strconv.Itoa(page))
	params.Set("page_size", strconv.Itoa(size))
	params.Set("availability_types", "ca")
	params.Set("release_status", "both")
	params.Set(
		"include_fields",
		"java_package_features,release_status,support_term,os,arch,hw_bitness,abi,java_package_type,javafx_bundled,sha256_hash,cpu_gen,size,archive_type,certifications,lib_c_type,crac_supported",
	)
	params.Set("azul_com", "true")
	params.Set("archive_type", "tar.gz")
	params.Set("lib_c_type", "glibc")

	osStr, arch, hwBitness := j.getCurrentSystemInfo()
	params.Set("os", osStr)
	if len(arch) > 0 {
		params.Set("arch", arch)
	}
	if len(hwBitness) > 0 {
		params.Set("hw_bitness", hwBitness)
	}

	targetUrl := zuluUrl + "?" + params.Encode()

	logger.Infof("Fetching %s", targetUrl)
	body, err := http.Default().Get(ctx, targetUrl)
	if err != nil {
		logger.Errorf("Failed to fetch %s: %s", targetUrl, err)
		return false, err
	}

	versions := make([]Version, 0)
	if err := json.Unmarshal(body, &versions); err != nil {
		logger.Errorf("Failed to unmarshal %s: %s", targetUrl, err)
		return false, err
	}

	for _, v := range versions {
		if v.Os == "linux" && v.LibCType != "glibc" {
			continue
		}
		if v.JavaPackageType != "jdk" {
			continue
		}
		vs := make([]string, len(v.JavaVersion))
		for i, num := range v.JavaVersion {
			vs[i] = strconv.Itoa(num)
		}

		ver, err := goversion.NewVersion(strings.Join(vs, ".") + "-zulu-" + v.Sha256Hash[:4])
		if err != nil {
			logger.Errorf("Failed to parse version %s: %s", v.Name, err)
			return false, err
		}

		comment := strings.ReplaceAll(v.Name, ".tar.gz", "")

		callback(&core.RemoteVersion{
			Version: ver,
			Origin:  v.DownloadUrl,
			Comment: comment,
		})
	}
	return len(versions) == 1000, nil
}

func (j *Java) ListRemoteVersions(ctx context.Context) ([]*core.RemoteVersion, error) {
	logger := log.GetLogger(ctx)
	res := make([]*core.RemoteVersion, 0)

	page := 1
	pageSize := 1000
	for {
		more, err := j.fetchRemote(ctx, page, pageSize, func(version *core.RemoteVersion) {
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
	logger.Infof("Installing %s(%s)", version.Version.String(), version.Comment)

	file, err := http.Default().
		Download(ctx, version.Origin, filepath.Join(path.GetLangRoot(lang), version.Version.String()), fmt.Sprintf("%s.%s-%s.tar.gz", version.Version.String(), runtime.GOOS, "amd64"))
	logger.Infof("")
	if err != nil {
		return fmt.Errorf("failed to download version: %s(%s): %w", version.Version.String(), version.Comment, err)
	}

	installDir := filepath.Join(path.GetLangRoot(lang), version.Version.String())

	if err := compress.UnTarGz(ctx, file, installDir); err != nil {
		return fmt.Errorf("failed to unTarGz: %s(%s): %w", version.Version.String(), version.Comment, err)
	}

	if err = os.RemoveAll(file); err != nil {
		logger.Warnf("failed to clean %s: %v", file, err)
	}

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
		"Version %s(%s) was successfully installed in %s",
		version.Version.String(),
		version.Comment,
		installDir,
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
