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
	"context"
	"io"
	"os"

	"github.com/toodofun/gvm/internal/core"
	"github.com/toodofun/gvm/internal/log"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
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
			ctx := cmd.Context()
			if cmd.Name() == "ui" {
				ctx = context.WithValue(ctx, core.ContextLogWriterKey, io.Discard)
			} else {
				ctx = context.WithValue(ctx, core.ContextLogWriterKey, os.Stdout)
			}
			cmd.SetContext(ctx)
			if debug {
				log.SetLevel(logrus.DebugLevel)
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
		NewCmdVersion(),
		NewAddAddonCmd(),
	)
	cmd.PersistentFlags().BoolVarP(&debug, "debug", "d", false, "debug mode")

	return cmd
}
