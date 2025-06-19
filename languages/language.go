package languages

import (
	"fmt"
	goversion "github.com/hashicorp/go-version"
	"gvm/core"
	"gvm/internal/common"
	"gvm/internal/log"
	"os"
	"path"
	"path/filepath"
)

// Language 默认方法
type Language struct {
	lang core.Language
}

func NewLanguage(lang core.Language) *Language {
	return &Language{lang: lang}
}

func (l *Language) SetDefaultVersion(version string) error {
	versions, err := l.lang.ListInstalledVersions()
	if err != nil {
		return err
	}

	versionSet := make(map[string]bool)
	for _, v := range versions {
		versionSet[v.Version.String()] = true
	}

	if _, ok := versionSet[version]; !ok {
		return fmt.Errorf("version %s not installed", version)
	}

	source := path.Join(common.GetLangRoot(l.lang.Name()), version)
	target := path.Join(common.GetLangRoot(l.lang.Name()), common.Current)

	// 检查是否加入了环境变量
	pathManager, err := common.NewPathManager()
	if err != nil {
		return fmt.Errorf("get path manager error: %w", err)
	}

	if err = pathManager.AddIfNotExists(path.Join(target, "go", "bin"), common.PositionPrepend); err != nil {
		return fmt.Errorf("add to path error: %v", err)
	}

	if !common.IsPathExist(source) {
		return fmt.Errorf("%s is not installed", version)
	}

	return common.SetSymlink(source, target)
}

func (l *Language) GetDefaultVersion() *core.InstalledVersion {
	defaultVersion := &core.InstalledVersion{
		Version: goversion.Must(goversion.NewVersion("0.0.0")),
	}
	target := path.Join(common.GetLangRoot(l.lang.Name()), common.Current)
	absTarget, err := os.Readlink(target)
	if err != nil {
		return defaultVersion
	}
	_, err = os.Lstat(absTarget)
	if err != nil {
		return defaultVersion
	}

	version := filepath.Base(absTarget)
	return &core.InstalledVersion{
		Version:  goversion.Must(goversion.NewVersion(version)),
		Location: absTarget,
	}
}

func (l *Language) ListInstalledVersions(binPath string) ([]*core.InstalledVersion, error) {
	installedVersions, err := common.GetInstalledVersion(l.lang.Name(), binPath)
	if err != nil {
		return nil, err
	}

	res := make([]*core.InstalledVersion, 0)
	for _, installedVersion := range installedVersions {
		version, err := goversion.NewVersion(installedVersion)
		if err != nil {
			log.Logger.Warnf("Failed to parse installed version %s: %+v", installedVersion, err)
			continue
		}
		res = append(res, &core.InstalledVersion{
			Version:  version,
			Origin:   installedVersion,
			Location: path.Join(common.GetLangRoot(l.lang.Name()), installedVersion),
		})
	}
	return res, nil
}

func (l *Language) Uninstall(version string) error {
	source := path.Join(common.GetLangRoot(l.lang.Name()), version)
	return os.RemoveAll(source)
}
