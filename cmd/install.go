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

package cmd

import (
	"fmt"

	"github.com/toodofun/gvm/internal/core"
	"github.com/toodofun/gvm/internal/log"
	"github.com/toodofun/gvm/internal/util/match"

	vers "github.com/hashicorp/go-version"
	"github.com/spf13/cobra"
)

func NewInstallCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "install <lang> <version>",
		Short: "Install a specific version of a language",
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) != 2 {
				return fmt.Errorf("requires two arguments: <lang> <version>")
			}
			return nil
		},
	}

	var setDefault bool
	cmd.Flags().BoolVar(&setDefault, "set-default", false, "Set the installed version as default")

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		lang := args[0]
		version := args[1]
		ctx := cmd.Context()
		logger := log.GetLogger(ctx)

		language, exists := core.GetLanguage(lang)
		if !exists {
			return cmd.Help()
		}

		// 检查远程是否存在
		versions, err := language.ListRemoteVersions(ctx)
		if err != nil {
			return err
		}

		vs := make([]*vers.Version, len(versions))
		versionMap := make(map[string]*core.RemoteVersion)
		for i, v := range versions {
			vs[i] = v.Version
			versionMap[v.Version.String()] = v
		}

		matchedVersion, err := match.MatchVersion(version, vs)
		if err != nil {
			return err
		}
		logger.Infof("Matched version %s", versionMap[matchedVersion.String()].Version.String())

		if err := language.Install(ctx, versionMap[matchedVersion.String()]); err != nil {
			return err
		}

		if setDefault {
			if err := language.SetDefaultVersion(ctx, matchedVersion.String()); err != nil {
				return err
			}
		}

		return nil
	}

	return cmd
}
