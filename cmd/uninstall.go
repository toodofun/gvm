package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"gvm/core"
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

			language, exists := core.GetLanguage(lang)
			if !exists {
				return cmd.Help()
			}

			return language.Uninstall(version)
		},
	}
}
