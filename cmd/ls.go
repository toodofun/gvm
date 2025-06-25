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
	"gvm/internal/core"
	"gvm/internal/util/color"
	"os"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/spf13/cobra"
)

func NewLsCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "ls <lang>",
		Short: "List installed versions of language",
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				return fmt.Errorf("you need to provide language information, such as golang or node")
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			lang := args[0]
			ctx := cmd.Context()

			language, exists := core.GetLanguage(lang)
			if !exists {
				return cmd.Help()
			}

			versions, err := language.ListInstalledVersions(ctx)
			if err != nil {
				return err
			}
			t := table.NewWriter()
			t.SetOutputMirror(os.Stdout)
			t.SetStyle(table.Style{
				Name: "custom",
				Box: table.BoxStyle{
					BottomLeft:       "-",
					BottomRight:      "-",
					BottomSeparator:  "-",
					Left:             "",
					LeftSeparator:    "",
					MiddleHorizontal: "-",
					MiddleSeparator:  "",
					PaddingLeft:      " ",
					PaddingRight:     " ",
					Right:            "",
					RightSeparator:   "",
					TopLeft:          "-",
					TopRight:         "-",
					TopSeparator:     "-",
					UnfinishedRow:    " ",
				},
				Options: table.Options{
					DrawBorder:      false,
					SeparateColumns: false,
					SeparateHeader:  true,
					SeparateRows:    false,
				},
			})

			// 获取已安装版本
			current := language.GetDefaultVersion(ctx)

			for _, version := range versions {
				flag := ""
				v := version.Version.String()
				l := version.Location
				if current.Version.Equal(version.Version) {
					flag = color.GreenFont("->")
					v = color.GreenFont(v)
					l = color.GreenFont(l)
				}
				t.AppendRow(table.Row{
					flag,
					v,
					l,
				})
			}
			t.Render()
			return nil
		},
	}
}
