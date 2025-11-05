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

package view

import (
	"context"
	"sort"
	"strings"
	"unicode"

	"github.com/toodofun/gvm/i18n"

	"github.com/toodofun/gvm/internal/core"

	"github.com/duke-git/lancet/v2/slice"
	"github.com/gdamore/tcell/v2"
)

type version struct {
	*core.RemoteVersion
	location    string
	isInstalled bool
	isDefault   bool
}

type LanguageVersions struct {
	versions []*version
	data     []*version
	lang     core.Language

	installed bool
}

func NewLanguageVersions() *LanguageVersions {
	return &LanguageVersions{}
}

func (lv *LanguageVersions) GetData(ctx context.Context, lang core.Language) (*LanguageVersions, error) {
	rvs, err := lang.ListRemoteVersions(ctx)
	if err != nil {
		rvs = []*core.RemoteVersion{}
	}
	sort.Slice(rvs, func(i, j int) bool {
		return rvs[i].Version.GreaterThan(rvs[j].Version)
	})
	installedVersions, err := lang.ListInstalledVersions(ctx)
	if err != nil {
		return nil, err
	}
	installedVersionList := make(map[string]*core.InstalledVersion)
	for _, iv := range installedVersions {
		installedVersionList[iv.Version.String()] = iv
	}
	current := lang.GetDefaultVersion(ctx)

	versions := make([]*version, 0)

	if len(rvs) > 0 {
		for _, rv := range rvs {
			if iv, ok := installedVersionList[rv.Version.String()]; ok {
				versions = append(versions, &version{
					RemoteVersion: rv,
					isInstalled:   ok,
					isDefault:     current.Version.Equal(rv.Version),
					location:      iv.Location,
				})
			} else {
				versions = append(versions, &version{
					RemoteVersion: rv,
					isInstalled:   ok,
					isDefault:     current.Version.Equal(rv.Version),
					location:      "",
				})
			}
		}
	} else {
		for _, iv := range installedVersionList {
			versions = append(versions, &version{
				RemoteVersion: &core.RemoteVersion{
					Version: iv.Version,
					Origin:  iv.Origin,
				},
				isInstalled: true,
				isDefault:   current.Version.Equal(iv.Version),
				location:    iv.Location,
			})
		}
	}

	return &LanguageVersions{
		versions:  versions,
		data:      versions,
		installed: false,
		lang:      lang,
	}, nil
}
func Capitalize(s string) string {
	if s == "" {
		return ""
	}
	runes := []rune(s)
	runes[0] = unicode.ToUpper(runes[0])
	return string(runes)
}
func (lv *LanguageVersions) Title() string {
	lang := ""
	if lv.lang != nil {
		lang = Capitalize(lv.lang.Name())
	}
	return i18n.GetTranslate("page.languageVersion.fullName", map[string]any{
		"lang": lang,
	})
}

func (lv *LanguageVersions) RowCount() int {
	return len(lv.data)
}

func (lv *LanguageVersions) Headers() []*TableHeader {
	return []*TableHeader{
		{
			Title:      i18n.GetTranslate("page.languageVersion.table.header.version", nil),
			FixedWidth: 20,
		},
		{
			Title:     i18n.GetTranslate("page.languageVersion.table.header.comment", nil),
			Expansion: 1,
		},
		{
			Title:     i18n.GetTranslate("page.languageVersion.table.header.location", nil),
			Expansion: 1,
		},
		{
			Title: i18n.GetTranslate("page.languageVersion.table.header.installed", nil),
			Hide:  true,
		},
	}
}

func (lv *LanguageVersions) Rows() [][]string {
	res := make([][]string, 0)
	for _, v := range lv.data {
		vs := v.Version.String()
		if v.isDefault {
			vs = "*" + vs
		}
		isInstalled := "false"
		if v.isInstalled {
			isInstalled = "true"
		}
		res = append(res, []string{vs, v.Comment, v.location, isInstalled})
	}
	return res
}

func (lv *LanguageVersions) GetRow(i []string) interface{} {
	for _, v := range lv.versions {
		if v.Version.String() == strings.TrimPrefix(i[0], "*") {
			return v
		}
	}
	return &version{}
}

func (lv *LanguageVersions) Filter(k, v string) {
	switch k {
	case "installed":
		lv.installed = !lv.installed
		lv.data = slice.Filter(lv.versions, func(index int, item *version) bool {
			if lv.installed {
				return item.isInstalled
			} else {
				return true
			}
		})
	}
}

func (lv *LanguageVersions) GetRowColor(i []string) tcell.Color {
	if i[3] == "true" {
		return tcell.ColorGreen
	} else {
		return tcell.ColorSkyblue
	}
}
