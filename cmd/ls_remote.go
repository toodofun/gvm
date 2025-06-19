package cmd

import (
	"fmt"
	"github.com/duke-git/lancet/v2/slice"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/spf13/cobra"
	"gvm/core"
	"gvm/internal/common"
	"os"
)

func NewLsRemoteCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "ls-remote <lang>",
		Short: "List remote versions of language",
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				return fmt.Errorf("you need to provide language information, such as golang or node")
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			lang := args[0]

			language, exists := core.GetLanguage(lang)
			if !exists {
				return cmd.Help()
			}
			if versions, err := language.ListRemoteVersions(); err != nil {
				return err
			} else {
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

				// 获取已安装列表
				installedVersions, err := language.ListInstalledVersions()
				if err != nil {
					installedVersions = make([]*core.InstalledVersion, 0)
				}

				installedVersionList := make([]string, 0)
				for _, iv := range installedVersions {
					installedVersionList = append(installedVersionList, iv.Version.String())
				}

				// 获取已安装版本
				current := language.GetDefaultVersion()

				for _, version := range versions {
					v := version.Version.String()
					c := version.Comment
					flag := ""
					if current.Version.Equal(version.Version) {
						flag = common.GreenFont("->")
					}
					if slice.Contain(installedVersionList, v) {
						v = common.GreenFont(fmt.Sprintf("%s(installed)", v))
						c = common.GreenFont(c)
					}
					t.AppendRow(table.Row{
						flag,
						v,
						c,
					})
				}

				t.Render()
			}

			return nil
		},
	}
}
