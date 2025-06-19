package cmd

import (
	"fmt"
	goversion "github.com/hashicorp/go-version"
	"github.com/spf13/cobra"
	"gvm/core"
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
