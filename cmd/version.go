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
	"fmt"
	"os"

	"github.com/toodofun/gvm/internal/util/version"

	"github.com/spf13/cobra"
)

// NewCmdVersion returns a cobra command for fetching versions.
func NewCmdVersion() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "version",
		Short: "Print version information",
		Long:  "Print version information for the current context",
		Example: "Print versions for the current context " +
			"\n\t\t gvm version",
		Run: func(cmd *cobra.Command, args []string) {
			versionInfo := version.Get()
			_, _ = fmt.Fprintf(os.Stdout, "%#v\n", versionInfo)
		},
	}
	return cmd
}
