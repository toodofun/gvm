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
	"gvm/core"

	goversion "github.com/hashicorp/go-version"
	"github.com/spf13/cobra"
)

func NewCurrentCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "current <lang>",
		Short: "Show Current version of a language",
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				return fmt.Errorf("requires two arguments: <lang>")
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			lang := args[0]

			language, exists := core.GetLanguage(lang)
			if !exists {
				return cmd.Help()
			}

			v := language.GetDefaultVersion()
			if v.Version.Equal(goversion.Must(goversion.NewVersion("0.0.0"))) {
				fmt.Println("not set")
			} else {
				fmt.Println("version: " + v.Version.String())
			}
			return nil
		},
	}
}
