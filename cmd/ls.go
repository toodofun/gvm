package cmd

import (
	"fmt"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/spf13/cobra"
	"gvm/core"
	"gvm/internal/common"
	"os"
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

			language, exists := core.GetLanguage(lang)
			if !exists {
				return cmd.Help()
			}

			versions, err := language.ListInstalledVersions()
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
			current := language.GetDefaultVersion()

			for _, version := range versions {
				flag := ""
				v := version.Version.String()
				l := version.Location
				if current.Version.Equal(version.Version) {
					flag = common.GreenFont("->")
					v = common.GreenFont(v)
					l = common.GreenFont(l)
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
