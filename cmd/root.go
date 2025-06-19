package cmd

import (
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"gvm/internal/log"
	"io"
)

var (
	debug bool
)

func NewRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "gvm",
		Short: "Language Version Manager",
		Long:  "A tool to manage multiple versions of programming languages.",
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			if cmd.Name() == "ui" {
				log.SetWriter(io.Discard)
			} else {
				log.SetWriter(cmd.OutOrStdout())
			}
			if debug {
				log.Logger.SetLevel(logrus.DebugLevel)
				log.Logger.Debugf("Debug logging enabled")
			}
		},
	}

	cmd.SilenceUsage = true
	cmd.SilenceErrors = true
	cmd.AddCommand(
		NewLsRemoteCmd(),
		NewLsCmd(),
		NewUseCmd(),
		NewInstallCmd(),
		NewUninstallCmd(),
		NewCurrentCmd(),
		NewUICmd(),
	)
	cmd.PersistentFlags().BoolVarP(&debug, "debug", "d", false, "debug mode")

	return cmd
}
