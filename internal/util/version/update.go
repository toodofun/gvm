package version

import (
	"context"
	"encoding/json"
	"strings"

	goversion "github.com/hashicorp/go-version"

	"github.com/toodofun/gvm/internal/http"
	"github.com/toodofun/gvm/internal/log"
)

type Release struct {
	Name       string `json:"name"`
	TagName    string `json:"tag_name"`
	Draft      bool   `json:"draft"`
	Prerelease bool   `json:"prerelease"`
}

func CheckUpdate(ctx context.Context) (has bool, latest string) {
	logger := log.GetLogger(ctx)
	const url = "https://api.github.com/repos/toodofun/gvm/releases/latest"
	body, err := http.Default().Get(ctx, url)
	if err != nil {
		logger.Debugf("Failed to get latest version from %s: %s", url, err)
		return false, ""
	}
	release := &Release{}
	if err := json.Unmarshal(body, release); err != nil {
		logger.Debugf("Failed to unmarshal latest version: %s", err)
		return false, ""
	}

	if release.Draft || release.Prerelease {
		return false, ""
	}

	newVersion := strings.TrimPrefix(release.TagName, "v")
	currentVersion := strings.TrimPrefix(Get().GitVersion, "v")

	nv, err := goversion.NewVersion(newVersion)
	if err != nil {
		logger.Debugf("Failed to parse new version %s: %s", newVersion, err)
		return false, ""
	}
	cv, err := goversion.NewVersion(currentVersion)
	if err != nil {
		logger.Debugf("Failed to parse current version %s: %s", currentVersion, err)
		return false, ""
	}

	if nv.GreaterThan(cv) {
		return true, release.TagName
	}
	return false, ""
}
