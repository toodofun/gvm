// Copyright 2025 The Toodofun Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package env

import (
	"path/filepath"
	"runtime"
	"strings"
)

// GetJavaEnvVars returns environment variables for Java
// P0: JAVA_HOME (critical - was completely missing)
// P1: CLASSPATH
func GetJavaEnvVars(installPath string) map[string]string {
	return map[string]string{
		"JAVA_HOME": installPath,
		"CLASSPATH": ".",
	}
}

// GetPythonEnvVars returns environment variables for Python
// P0: PYTHONHOME (critical - was commented out)
// P1: PYTHONPATH, PYTHONDONTWRITEBYTECODE, PYTHONUNBUFFERED, PYTHONIOENCODING
func GetPythonEnvVars(installPath string) map[string]string {
	// Extract version from path for PYTHONPATH, e.g. /path/to/3.11.0 -> 3.11
	// Default to python3 if we can't determine version
	version := filepath.Base(installPath)
	parts := strings.Split(version, ".")
	majorMinor := "python3" // A sensible fallback
	if len(parts) >= 2 {
		majorMinor = "python" + parts[0] + "." + parts[1]
	}
	libPath := filepath.Join(installPath, "lib", majorMinor, "site-packages")

	return map[string]string{
		"PYTHONHOME":              installPath,
		"PYTHONPATH":              libPath,
		"PYTHONDONTWRITEBYTECODE": "1",
		"PYTHONUNBUFFERED":        "1",
		"PYTHONIOENCODING":        "utf-8",
	}
}

// GetGolangEnvVars returns environment variables for Go
// P1: GO111MODULE (critical - was missing)
// P1: GOROOT, GOPATH, GOBIN
func GetGolangEnvVars(installPath, gopath string) map[string]string {
	return map[string]string{
		"GO111MODULE": "on",
		"GOROOT":      filepath.Join(installPath, "go"),
		"GOPATH":      gopath,
		"GOBIN":       filepath.Join(gopath, "bin"),
	}
}

// GetNodeEnvVars returns environment variables for Node.js
// P1: NODE_PATH, NPM_CONFIG_PREFIX
func GetNodeEnvVars(installPath string) map[string]string {
	nodePath := filepath.Join(installPath, "node")
	if runtime.GOOS == "windows" {
		nodePath = installPath
	}

	return map[string]string{
		"NODE_PATH":         filepath.Join(nodePath, "lib", "node_modules"),
		"NPM_CONFIG_PREFIX": nodePath,
	}
}
