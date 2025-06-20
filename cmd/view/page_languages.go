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
	"fmt"
	"github.com/gdamore/tcell/v2"
	"gvm/core"
)

type PageLanguages struct {
	*SearchTable

	app *Application
}

func NewPageLanguages(app *Application) *PageLanguages {
	languages := NewLanguages()
	root := NewSearchTable(languages, app)

	page := &PageLanguages{
		SearchTable: root,
		app:         app,
	}

	return page
}

func (p *PageLanguages) Init() {
	p.BindKeys(KeyMap{
		tcell.KeyEnter: NewKeyAction("Enter", func(evt *tcell.EventKey) *tcell.EventKey {
			languageName := p.GetSelection().(string)
			lang, ok := core.GetLanguage(languageName)
			if !ok {
				p.app.Alert(fmt.Sprintf("language not found: %s", languageName), p.table)
			}
			p.app.lang = lang
			p.app.SwitchPage(pageLanguageVersions)
			return evt
		}, true),
	})
}

func (p *PageLanguages) GetKeyActions() *KeyActions {
	return p.actions
}
