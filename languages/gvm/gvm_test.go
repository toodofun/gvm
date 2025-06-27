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
