package cmd

import (
	"errors"
	"fmt"

	"github.com/toodofun/gvm/languages/github"

	"github.com/spf13/cobra"

	"github.com/toodofun/gvm/internal/core"
	"github.com/toodofun/gvm/internal/util/file"
)

func NewAddAddonCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "add <addon>",
		Short: "Add a new addon to the GVM",
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) != 3 {
				return errors.New("this command needs exactly two arguments: <provider> <name> <data source name>")
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if _, ok := core.GetLanguage(args[1]); ok {
				return fmt.Errorf("language %s already exists, please use a different name", args[1])
			}
			config := core.GetConfig()
			var err error
			switch args[0] {
			case "github":
				_, err = github.NewGithub(args[1], args[2])
			default:
				err = fmt.Errorf("unsupported addon type: %s", args[0])
			}
			if err != nil {
				return err
			}
			config.Addon = append(config.Addon, core.LanguageItem{
				Provider:       args[0],
				Name:           args[1],
				DataSourceName: args[2],
			})
			return file.WriteJSONFile(core.GetConfigPath(), config)
		},
	}
}
