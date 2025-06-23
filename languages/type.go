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
