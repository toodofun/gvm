package core

import (
	"github.com/hashicorp/go-version"
)

type Language interface {
	Name() string
	ListRemoteVersions() ([]*RemoteVersion, error)
	ListInstalledVersions() ([]*InstalledVersion, error)
	SetDefaultVersion(version string) error
	GetDefaultVersion() *InstalledVersion
	Install(remoteVersion *RemoteVersion) error
	Uninstall(version string) error
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
