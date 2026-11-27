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
	"testing"
)

func TestGetJavaEnvVars(t *testing.T) {
	tests := []struct {
		name          string
		installPath   string
		wantJAVA_HOME string
		wantCLASSPATH string
	}{
		{
			name:          "standard path",
			installPath:   "/opt/java/jdk-17",
			wantJAVA_HOME: "/opt/java/jdk-17",
			wantCLASSPATH: ".",
		},
		{
			name:          "path with spaces",
			installPath:   "/Users/test/Library/Java/JavaVirtualMachines/jdk-17",
			wantJAVA_HOME: "/Users/test/Library/Java/JavaVirtualMachines/jdk-17",
			wantCLASSPATH: ".",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetJavaEnvVars(tt.installPath)

			// Check JAVA_HOME (P0 critical)
			if javaHome, ok := got["JAVA_HOME"]; !ok {
				t.Errorf("GetJavaEnvVars() missing JAVA_HOME (P0 critical)")
			} else if javaHome != tt.wantJAVA_HOME {
				t.Errorf("GetJavaEnvVars() JAVA_HOME = %v, want %v", javaHome, tt.wantJAVA_HOME)
			}

			// Check CLASSPATH
			if classpath, ok := got["CLASSPATH"]; !ok {
				t.Errorf("GetJavaEnvVars() missing CLASSPATH")
			} else if classpath != tt.wantCLASSPATH {
				t.Errorf("GetJavaEnvVars() CLASSPATH = %v, want %v", classpath, tt.wantCLASSPATH)
			}
		})
	}
}

func TestGetPythonEnvVars(t *testing.T) {
	tests := []struct {
		name                        string
		installPath                 string
		wantPYTHONHOME              string
		wantPYTHONPATH              string
		wantPYTHONDONTWRITEBYTECODE string
		wantPYTHONUNBUFFERED        string
		wantPYTHONIOENCODING        string
	}{
		{
			name:                        "standard path",
			installPath:                 "/opt/python/3.11.0",
			wantPYTHONHOME:              "/opt/python/3.11.0",
			wantPYTHONPATH:              "/opt/python/3.11.0/lib/python3.11/site-packages",
			wantPYTHONDONTWRITEBYTECODE: "1",
			wantPYTHONUNBUFFERED:        "1",
			wantPYTHONIOENCODING:        "utf-8",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetPythonEnvVars(tt.installPath)

			// Check PYTHONHOME (P0 critical - was commented out)
			if pythonhome, ok := got["PYTHONHOME"]; !ok {
				t.Errorf("GetPythonEnvVars() missing PYTHONHOME (P0 critical)")
			} else if pythonhome != tt.wantPYTHONHOME {
				t.Errorf("GetPythonEnvVars() PYTHONHOME = %v, want %v", pythonhome, tt.wantPYTHONHOME)
			}

			// Check PYTHONPATH
			if pythonpath, ok := got["PYTHONPATH"]; !ok {
				t.Errorf("GetPythonEnvVars() missing PYTHONPATH")
			} else if pythonpath != tt.wantPYTHONPATH {
				t.Errorf("GetPythonEnvVars() PYTHONPATH = %v, want %v", pythonpath, tt.wantPYTHONPATH)
			}

			// Check PYTHONDONTWRITEBYTECODE
			if dontwrite, ok := got["PYTHONDONTWRITEBYTECODE"]; !ok {
				t.Errorf("GetPythonEnvVars() missing PYTHONDONTWRITEBYTECODE")
			} else if dontwrite != tt.wantPYTHONDONTWRITEBYTECODE {
				t.Errorf("GetPythonEnvVars() PYTHONDONTWRITEBYTECODE = %v, want %v", dontwrite, tt.wantPYTHONDONTWRITEBYTECODE)
			}

			// Check PYTHONUNBUFFERED
			if unbuffered, ok := got["PYTHONUNBUFFERED"]; !ok {
				t.Errorf("GetPythonEnvVars() missing PYTHONUNBUFFERED")
			} else if unbuffered != tt.wantPYTHONUNBUFFERED {
				t.Errorf("GetPythonEnvVars() PYTHONUNBUFFERED = %v, want %v", unbuffered, tt.wantPYTHONUNBUFFERED)
			}

			// Check PYTHONIOENCODING
			if ioencoding, ok := got["PYTHONIOENCODING"]; !ok {
				t.Errorf("GetPythonEnvVars() missing PYTHONIOENCODING")
			} else if ioencoding != tt.wantPYTHONIOENCODING {
				t.Errorf("GetPythonEnvVars() PYTHONIOENCODING = %v, want %v", ioencoding, tt.wantPYTHONIOENCODING)
			}
		})
	}
}

func TestGetGolangEnvVars(t *testing.T) {
	tests := []struct {
		name            string
		installPath     string
		gopath          string
		wantGO111MODULE string
		wantGOROOT      string
		wantGOPATH      string
		wantGOBIN       string
	}{
		{
			name:            "standard path",
			installPath:     "/opt/go/1.21.0",
			gopath:          "/Users/test/.gvm/go/gopath",
			wantGO111MODULE: "on",
			wantGOROOT:      "/opt/go/1.21.0/go",
			wantGOPATH:      "/Users/test/.gvm/go/gopath",
			wantGOBIN:       "/Users/test/.gvm/go/gopath/bin",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetGolangEnvVars(tt.installPath, tt.gopath)

			// Check GO111MODULE (P1 critical)
			if go111module, ok := got["GO111MODULE"]; !ok {
				t.Errorf("GetGolangEnvVars() missing GO111MODULE (P1 critical)")
			} else if go111module != tt.wantGO111MODULE {
				t.Errorf("GetGolangEnvVars() GO111MODULE = %v, want %v", go111module, tt.wantGO111MODULE)
			}

			// Check GOROOT
			if goroot, ok := got["GOROOT"]; !ok {
				t.Errorf("GetGolangEnvVars() missing GOROOT")
			} else if goroot != tt.wantGOROOT {
				t.Errorf("GetGolangEnvVars() GOROOT = %v, want %v", goroot, tt.wantGOROOT)
			}

			// Check GOPATH
			if gopath, ok := got["GOPATH"]; !ok {
				t.Errorf("GetGolangEnvVars() missing GOPATH")
			} else if gopath != tt.wantGOPATH {
				t.Errorf("GetGolangEnvVars() GOPATH = %v, want %v", gopath, tt.wantGOPATH)
			}

			// Check GOBIN
			if gobin, ok := got["GOBIN"]; !ok {
				t.Errorf("GetGolangEnvVars() missing GOBIN")
			} else if gobin != tt.wantGOBIN {
				t.Errorf("GetGolangEnvVars() GOBIN = %v, want %v", gobin, tt.wantGOBIN)
			}
		})
	}
}

func TestGetNodeEnvVars(t *testing.T) {
	tests := []struct {
		name                  string
		installPath           string
		wantNODE_PATH         string
		wantNPM_CONFIG_PREFIX string
	}{
		{
			name:                  "standard path",
			installPath:           "/opt/node/20.0.0",
			wantNODE_PATH:         "/opt/node/20.0.0/node/lib/node_modules",
			wantNPM_CONFIG_PREFIX: "/opt/node/20.0.0/node",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetNodeEnvVars(tt.installPath)

			// Check NODE_PATH
			if nodePath, ok := got["NODE_PATH"]; !ok {
				t.Errorf("GetNodeEnvVars() missing NODE_PATH")
			} else if nodePath != tt.wantNODE_PATH {
				t.Errorf("GetNodeEnvVars() NODE_PATH = %v, want %v", nodePath, tt.wantNODE_PATH)
			}

			// Check NPM_CONFIG_PREFIX
			if npmPrefix, ok := got["NPM_CONFIG_PREFIX"]; !ok {
				t.Errorf("GetNodeEnvVars() missing NPM_CONFIG_PREFIX")
			} else if npmPrefix != tt.wantNPM_CONFIG_PREFIX {
				t.Errorf("GetNodeEnvVars() NPM_CONFIG_PREFIX = %v, want %v", npmPrefix, tt.wantNPM_CONFIG_PREFIX)
			}
		})
	}
}
