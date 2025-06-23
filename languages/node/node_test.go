package node

import (
	goversion "github.com/hashicorp/go-version"
	"github.com/stretchr/testify/require"
	"gvm/core"
	"os/exec"
	"strings"
	"testing"
)

func TestListRemoteVersions(t *testing.T) {
	node := NewNode(defaultBaseURL, t.TempDir())
	list, err := node.ListRemoteVersions()
	require.NoError(t, err)
	require.NotNil(t, list)
}
func TestInstall(t *testing.T) {
	node := &Node{
		baseURL:    defaultBaseURL,
		baseDir:    t.TempDir(),
		versionMap: make(map[string]*Version),
	}
	list, err := node.ListRemoteVersions()
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
			actualErr := node.Install(tt.versionFunc())
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
