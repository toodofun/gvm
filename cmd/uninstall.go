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

	"github.com/spf13/cobra"
)

func NewUninstallCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "uninstall <lang> <version>",
		Short: "Uninstall a specific version of a language",
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) != 2 {
				return fmt.Errorf("requires two arguments: <lang> <version>")
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			lang := args[0]
			version := args[1]
			ctx := cmd.Context()

			language, exists := core.GetLanguage(lang)
			if !exists {
				return cmd.Help()
			}

			return language.Uninstall(ctx, version)
		},
	}
}
