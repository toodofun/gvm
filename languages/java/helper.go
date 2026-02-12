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
	"net/url"
	"runtime"
	"strconv"
	"strings"

	goversion "github.com/hashicorp/go-version"

	"github.com/toodofun/gvm/internal/core"
	"github.com/toodofun/gvm/internal/http"
	"github.com/toodofun/gvm/internal/log"
	"github.com/toodofun/gvm/internal/util/env"
)

func currentSystemInfo() (os, arch, hwBitness string) {
	switch strings.ToLower(runtime.GOOS) {
	case env.RuntimeFromLinux:
		os = env.RuntimeFromLinux
	case env.RuntimeFromWindows:
		os = env.RuntimeFromWindows
	case env.RuntimeFromDarwin:
		os = env.RuntimeFromMacos
	}

	switch strings.ToLower(runtime.GOARCH) {
	case env.ArchAMD64:
		arch = env.ArchX86
		hwBitness = env.Bitness64

	case env.ArchARM64:
		arch = env.ArchARMGeneric
		hwBitness = env.Bitness64
		if runtime.GOOS == env.RuntimeFromDarwin {
			arch = ""
			hwBitness = ""
		}

	case env.Arch386:
		arch = env.ArchX86
		hwBitness = env.Bitness32

	case env.ArchARM:
		arch = env.ArchARMGeneric
		hwBitness = env.Bitness32
	}
	return os, arch, hwBitness
}

func fetchRemote(
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

	osStr, arch, hwBitness := currentSystemInfo()
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
		if v.Os == env.RuntimeFromLinux && v.LibCType != "glibc" {
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
