package cmd

import (
	"errors"

	"github.com/spf13/cobra"
	"github.com/toodofun/gvm/i18n"
)

func NewSetLanguageCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "set-language <language name>",
		Short: "Set defaule application language",
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
