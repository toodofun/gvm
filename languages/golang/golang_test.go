package golang

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGolang_ListRemoteVersions(t *testing.T) {
	golang := &Golang{}
	remoteVersions, err := golang.ListRemoteVersions(context.Background())
	assert.NoError(t, err)
	assert.NotNil(t, remoteVersions)
}

func TestGolang_ListInstalledVersions(t *testing.T) {
	golang := &Golang{}
	installedVersions, err := golang.ListInstalledVersions(context.Background())
	assert.NoError(t, err)
	assert.NotNil(t, installedVersions)
}
