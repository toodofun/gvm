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
	"errors"

	"github.com/spf13/cobra"

	"github.com/toodofun/gvm/i18n"
)

func NewSetLanguageCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "set-language <language name>",
		Short: "Set default application language, supported languages: en, zh",
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				return errors.New("this command needs exactly one argument: <language name>, eg: en")
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return i18n.SetLanguage(args[0])
		},
	}
}
