package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"gvm/core"
)

func NewUseCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "use <lang>",
		Short: "Set default versions of language",
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) != 2 {
				return fmt.Errorf("you need to provide language information, such as golang or node")
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

			return language.SetDefaultVersion(version)
		},
	}
}
