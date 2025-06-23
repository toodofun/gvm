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

package node

import (
	"context"
	goversion "github.com/hashicorp/go-version"
	"github.com/stretchr/testify/require"
	"gvm/internal/core"
	"os/exec"
	"strings"
	"testing"
)

func TestListRemoteVersions(t *testing.T) {
	node := NewNode(defaultBaseURL, t.TempDir())
	list, err := node.ListRemoteVersions(context.TODO())
	require.NoError(t, err)
	require.NotNil(t, list)
}
func TestInstall(t *testing.T) {
	node := &Node{
		baseURL:    defaultBaseURL,
		baseDir:    t.TempDir(),
		versionMap: make(map[string]*Version),
	}
	ctx := context.TODO()
	list, err := node.ListRemoteVersions(ctx)
	require.NoError(t, err)
	require.NotNil(t, list)

	tests := []struct {
		name        string
		versionFunc func() *core.RemoteVersion
		err         error
	}{
		{
			name: "install return success",
			versionFunc: func() *core.RemoteVersion {
				version, err := goversion.NewVersion("23.0.0")
				require.NoError(t, err)
				return &core.RemoteVersion{
					Version: version,
					Origin:  "v23.0.0",
				}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actualErr := node.Install(ctx, tt.versionFunc())
			require.Equal(t, tt.err, actualErr)
			if actualErr != nil {
				return
			}
		})
	}
}

func equalVersion(execPath, expectedVersion string, expectedErr error, t *testing.T) {
	cmd := exec.Command(execPath, "-v")
	output, actualErr := cmd.Output()
	require.Equal(t, expectedErr, actualErr)

	if actualErr != nil {
		return
	}
	version := strings.TrimSpace(string(output))
	require.Equal(t, expectedVersion, version)
}
