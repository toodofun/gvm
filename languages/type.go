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

package languages

import (
	"fmt"
	"strings"
)

type PackageSuffix string

func (p PackageSuffix) String() string {
	return string(p)
}

func (p PackageSuffix) Composition(name string) string {
	if strings.HasSuffix(name, p.String()) {
		return name
	}
	return fmt.Sprintf("%s%s", name, p.String())
}

const (
	Tar PackageSuffix = ".tar.gz"
	Zip PackageSuffix = ".zip"
	Pkg PackageSuffix = ".pkg"
)

type PackageSuffixes []PackageSuffix

var AllSuffix PackageSuffixes = []PackageSuffix{Tar, Zip, Pkg}

func (p PackageSuffixes) Has(packageName string) bool {
	for _, t := range p {
		if strings.HasSuffix(packageName, t.String()) {
			return true
		}
	}
	return false
}

func (p PackageSuffixes) GetSuffix(packageName string) PackageSuffix {
	for _, t := range p {
		if strings.HasSuffix(packageName, t.String()) {
			return t
		}
	}
	return ""
}

func (p PackageSuffixes) Trim(packageName string) string {
	for _, t := range p {
		if strings.HasSuffix(packageName, t.String()) {
			return strings.TrimSuffix(packageName, t.String())
		}
	}
	return packageName
}
