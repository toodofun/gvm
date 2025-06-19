package golang

import (
	"encoding/json"
	"fmt"
	goversion "github.com/hashicorp/go-version"
	"gvm/core"
	"gvm/internal/common"
	"gvm/internal/http"
	"gvm/internal/log"
	"gvm/languages"
	"path"
	"runtime"
	"strings"
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

func (g *Golang) ListRemoteVersions() ([]*core.RemoteVersion, error) {
	res := make([]*core.RemoteVersion, 0)
	body, err := http.Default().Get(fmt.Sprintf("%s?mode=json&include=all", baseUrl))
	if err != nil {
		log.Logger.Errorf("Get remote versions error: %s", err.Error())
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
		ver, err := goversion.NewVersion(strings.TrimPrefix(v.Version, "go"))
		if err != nil {
			log.Logger.Warnf("Failed to parse version %s: %s", v.Version, err)
			continue
		}
		res = append(res, &core.RemoteVersion{
			Version: ver,
			Origin:  v.Version,
			Comment: comment,
		})
	}

	common.ReverseSlice(res)

	return res, nil
}

func (g *Golang) ListInstalledVersions() ([]*core.InstalledVersion, error) {
	return languages.NewLanguage(g).ListInstalledVersions(path.Join("go", "bin"))
}

func (g *Golang) SetDefaultVersion(version string) error {
	return languages.NewLanguage(g).SetDefaultVersion(version)
}

func (g *Golang) GetDefaultVersion() *core.InstalledVersion {
	return languages.NewLanguage(g).GetDefaultVersion()
}

func (g *Golang) Uninstall(version string) error {
	return languages.NewLanguage(g).Uninstall(version)
}

func (g *Golang) Install(version *core.RemoteVersion) error {
	log.Logger.Debugf("Install remote version: %s", version.Origin)
	// 检查是否已经安装
	installed, err := g.ListInstalledVersions()
	if err != nil {
		log.Logger.Errorf("Failed to list installed versions: %+v", err)
		return err
	}
	for _, ver := range installed {
		if ver.Version.Equal(version.Version) {
			log.Logger.Infof("Version %s already installed", version.Version.String())
			return nil
		}
	}

	log.Logger.Infof("Installing version %s", version.Version.String())
	// 检查版本是否存在
	url := fmt.Sprintf("%s%s.%s-%s.tar.gz", baseUrl, version.Origin, runtime.GOOS, runtime.GOARCH)
	if runtime.GOOS == "windows" {
		url = fmt.Sprintf("%s%s.%s-%s.zip", baseUrl, version.Origin, runtime.GOOS, runtime.GOARCH)
	}
	head, code, err := http.Default().Head(url)
	if err != nil {
		return err
	}
	if runtime.GOOS == "darwin" && code == 404 {
		log.Logger.Infof("Version %s not found for %s/%s, trying %s/amd64", version.Version.String(), runtime.GOOS, runtime.GOARCH, runtime.GOOS)
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

	log.Logger.Infof("Downloading: %s, size: %s", url, head.Get("Content-Length"))
	file, err := http.Default().Download(url, path.Join(core.GetRootDir(), "go", version.Version.String()), fmt.Sprintf("%s.%s-%s.tar.gz", version.Origin, runtime.GOOS, "amd64"))
	log.Logger.Print("\n")
	if err != nil {
		return fmt.Errorf("failed to download version %s: %w", version, err)
	}
	log.Logger.Infof("Extracting: %s, size: %s", url, head.Get("Content-Length"))
	if strings.HasSuffix(url, ".tar.gz") {
		if err := common.UnTarGz(file, path.Join(core.GetRootDir(), "go", version.Version.String())); err != nil {
			log.Logger.Warnf("Failed to untar version %s: %s", version, err)
			return fmt.Errorf("failed to extract version %s: %w", version, err)
		}
	} else if strings.HasSuffix(url, ".zip") {
		if err := common.UnZip(file, path.Join(core.GetRootDir(), "go", version.Version.String())); err != nil {
			log.Logger.Warnf("Failed to untar version %s: %s", version, err)
			return fmt.Errorf("failed to extract version %s: %w", version, err)
		}
	}

	log.Logger.Infof("Version %s was successfully installed in %s", version.Version.String(), path.Join(core.GetRootDir(), "go", version.Version.String(), "go", "bin"))
	return nil
}

func init() {
	core.RegisterLanguage(&Golang{})
}
