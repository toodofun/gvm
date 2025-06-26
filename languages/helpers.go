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

package languages

import (
	"context"

	"github.com/toodofun/gvm/internal/core"
	"github.com/toodofun/gvm/internal/log"

	goversion "github.com/hashicorp/go-version"
)

// HasInstall the Check if it is installed
func HasInstall(ctx context.Context, language core.Language, version goversion.Version) (error, bool) {
	logger := log.GetLogger(ctx)
	installed, err := language.ListInstalledVersions(ctx)
	if err != nil {
		logger.Errorf("Failed to list installed versions: %+v", err)
		return err, false
	}
	for _, ver := range installed {
		if ver.Version.Equal(&version) {
			logger.Infof("Version %s already installed", version.String())
			return nil, true
		}
	}
	return nil, false
}
