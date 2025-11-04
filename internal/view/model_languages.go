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
	"github.com/toodofun/gvm/i18n"
	"github.com/toodofun/gvm/internal/core"

	"github.com/gdamore/tcell/v2"
)

type Languages struct {
	languages []string
}

func NewLanguages() *Languages {
	languages := core.GetAllLanguage()
	return &Languages{
		languages: languages,
	}
}

func (l *Languages) Title() string {
	return i18n.GetTranslate("page.language.fullName", nil)
}

func (l *Languages) RowCount() int {
	return len(l.languages)
}

func (l *Languages) Headers() []*TableHeader {
	return []*TableHeader{
		{
			Title:     "Language",
			Expansion: 1,
		},
	}
}

func (l *Languages) GetRow(i []string) interface{} {
	for _, v := range l.languages {
		if v == i[0] {
			return v
		}
	}
	return ""
}

func (l *Languages) GetRowColor(i []string) tcell.Color {
	return tcell.ColorSkyblue
}

func (l *Languages) Filter(k, v string) {

}

func (l *Languages) Rows() [][]string {
	res := make([][]string, 0)
	for _, lang := range l.languages {
		res = append(res, []string{lang})
	}
	return res
}
