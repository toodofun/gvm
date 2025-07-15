package github

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	goversion "github.com/hashicorp/go-version"

	"github.com/toodofun/gvm/internal/http"
	"github.com/toodofun/gvm/internal/log"
	"github.com/toodofun/gvm/internal/util/compress"
	"github.com/toodofun/gvm/internal/util/env"
	"github.com/toodofun/gvm/internal/util/path"
	"github.com/toodofun/gvm/languages"

	"github.com/toodofun/gvm/internal/core"
)

const (
	apiBaseUrl = "https://api.github.com/repos/%s/%s/releases"
)

type Github struct {
	name  string
	owner string
	repo  string
}

func (g *Github) ListRemoteVersions(ctx context.Context) ([]*core.RemoteVersion, error) {
	logger := log.GetLogger(ctx)
	res := make([]*core.RemoteVersion, 0)

	releases, err := NewGitHubClient("").GetAllReleases(ctx, g.owner, g.repo)
	if err != nil {
		logger.Errorf("get releases error: %s", err)
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

func (g *Github) ListInstalledVersions(ctx context.Context) ([]*core.InstalledVersion, error) {
	return languages.NewLanguage(g).ListInstalledVersions(ctx, filepath.Join())
}

func (g *Github) SetDefaultVersion(ctx context.Context, version string) error {
	envs := []env.KV{
		{
			Key:    "PATH",
			Value:  filepath.Join(path.GetLangRoot(g.Name()), path.Current),
			Append: true,
		},
		{
			Key:    "PATH",
			Value:  filepath.Join(path.GetLangRoot(g.Name()), path.Current, "bin"),
			Append: true,
		},
	}
	return languages.NewLanguage(g).SetDefaultVersion(ctx, version, envs)
}

func (g *Github) GetDefaultVersion(ctx context.Context) *core.InstalledVersion {
	return languages.NewLanguage(g).GetDefaultVersion()
}

func (g *Github) Install(ctx context.Context, remoteVersion *core.RemoteVersion) error {
	logger := log.GetLogger(ctx)
	logger.Infof("Install remote version %s", remoteVersion.Origin)
	lang := g.Name()

	body, err := http.Default().Get(ctx, fmt.Sprintf(apiBaseUrl, g.owner, g.repo))
	if err != nil {
		logger.Errorf("Get remote versions error: %v", err)
		return err
	}

	releases := make([]Release, 0)
	if err := json.Unmarshal(body, &releases); err != nil {
		logger.Errorf("Unmarshal remote versions error: %v", err)
		return err
	}

	var record *Release
	for _, release := range releases {
		if release.Name == remoteVersion.Origin {
			record = &release
		}
	}
	if record == nil {
		logger.Errorf("Release %s not found", remoteVersion.Origin)
		return fmt.Errorf("remote version %s not found", remoteVersion.Origin)
	}

	url := ""
	name := ""
	for _, asset := range *record.Assets {
		if strings.Contains(strings.ToLower(asset.Name), runtime.GOOS) &&
			strings.Contains(strings.ToLower(asset.Name), runtime.GOARCH) {
			url = asset.DownloadURL
			name = asset.Name
			break
		}
	}
	if len(url) == 0 && runtime.GOOS == "darwin" {
		for _, asset := range *record.Assets {
			if strings.Contains(strings.ToLower(asset.Name), runtime.GOOS) &&
				strings.Contains(strings.ToLower(asset.Name), "amd64") {
				url = asset.DownloadURL
				name = asset.Name
				break
			}
		}
	}

	if len(url) == 0 {
		logger.Errorf("Release %s not found", remoteVersion.Origin)
		return fmt.Errorf("remote version %s not found", remoteVersion.Origin)
	}

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
		Download(ctx, url, filepath.Join(path.GetLangRoot(lang), remoteVersion.Version.String()), name)
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

func (g *Github) Uninstall(ctx context.Context, version string) error {
	return languages.NewLanguage(g).Uninstall(version)
}

func NewGithub(name, dsn string) (*Github, error) {
	parts := strings.Split(dsn, "/")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid data source name: %s, expected format: <owner>/<repo>", dsn)
	}
	return &Github{
		name:  name,
		owner: parts[0],
		repo:  parts[1],
	}, nil
}

func (g *Github) Name() string {
	return g.name
}

func init() {
	config := core.GetConfig()
	for _, addon := range config.Addon {
		if addon.Provider == "github" {
			lang, err := NewGithub(addon.Name, addon.DataSourceName)
			if err == nil {
				core.RegisterLanguage(lang)
			}
		}
	}
}
