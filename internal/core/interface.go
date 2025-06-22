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

package core

import (
	"context"
	"github.com/hashicorp/go-version"
)

type Language interface {
	Name() string
	ListRemoteVersions(ctx context.Context) ([]*RemoteVersion, error)
	ListInstalledVersions(ctx context.Context) ([]*InstalledVersion, error)
	SetDefaultVersion(ctx context.Context, version string) error
	GetDefaultVersion(ctx context.Context) *InstalledVersion
	Install(ctx context.Context, remoteVersion *RemoteVersion) error
	Uninstall(ctx context.Context, version string) error
}

type RemoteVersion struct {
	Version *version.Version
	Origin  string
	Comment string
}

type InstalledVersion struct {
	Version  *version.Version
	Origin   string
	Location string
}
