package cmd

import (
	"github.com/spf13/cobra"
	"gvm/cmd/view"
)

func NewUICmd() *cobra.Command {
	return &cobra.Command{
		Use:   "ui",
		Short: "Run in the terminal UI",
		RunE: func(cmd *cobra.Command, args []string) error {
			return view.CreateApplication().Run()
		},
	}
}
