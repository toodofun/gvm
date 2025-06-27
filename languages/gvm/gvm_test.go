package gvm

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGVM_ListRemoteVersions(t *testing.T) {
	gvm := &GVM{}
	remoteVersions, err := gvm.ListRemoteVersions(context.Background())
	assert.NoError(t, err)
	assert.NotNil(t, remoteVersions)
}

func TestGVM_ListInstalledVersions(t *testing.T) {
	gvm := &GVM{}
	installedVersions, err := gvm.ListInstalledVersions(context.Background())
	assert.NoError(t, err)
	assert.NotNil(t, installedVersions)
}

func TestGVM_GetDefaultVersion(t *testing.T) {
	gvm := &GVM{}
	defaultVersion := gvm.GetDefaultVersion(context.Background())
	assert.NotNil(t, defaultVersion)
}
