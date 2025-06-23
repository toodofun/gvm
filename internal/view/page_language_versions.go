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
	"fmt"
	"gvm/internal/core"
	"gvm/internal/log"
	"gvm/languages/golang"

	"github.com/gdamore/tcell/v2"
	"github.com/google/uuid"
)

type PageLanguageVersions struct {
	*SearchTable
	languageVersions *LanguageVersions

	app *Application
}

func NewPageLanguageVersions(app *Application) *PageLanguageVersions {
	logger := log.GetLogger(app.ctx)
	lv := NewLanguageVersions()
	p := &PageLanguageVersions{
		SearchTable:      NewSearchTable(lv, app),
		languageVersions: lv,
		app:              app,
	}
	p.BindKeys(KeyMap{
		KeyI: NewKeyAction("Filter by installed", func(evt *tcell.EventKey) *tcell.EventKey {
			model := p.GetModel()
			model.Filter("installed", "")
			p.SetModel(model)
			p.Render()
			return evt
		}, true),
		tcell.KeyESC: NewKeyAction("Go back", func(evt *tcell.EventKey) *tcell.EventKey {
			p.app.SwitchPage(pageLanguages)
			return evt
		}, true),
		tcell.KeyEnter: NewKeyAction("Install or set as default", func(evt *tcell.EventKey) *tcell.EventKey {
			vi, ok := p.GetSelection()
			if !ok {
				return evt
			}
			v := vi.(*version)
			if v.isInstalled {
				p.app.Confirm(fmt.Sprintf("Are you sure you want to set %s as default", v.Version.String()), func() {
					if v.isDefault {
						p.app.Alert(fmt.Sprintf("%s is already the default version", v.Version.String()), p.table)
					} else {
						p.doAsync(fmt.Sprintf("Set %s as default", v.Version.String()), func() (interface{}, error) {
							if err := p.app.lang.SetDefaultVersion(context.Background(), v.Version.String()); err != nil {
								return nil, err
							}
							return nil, nil
						}, func(i interface{}) {
							p.refresh()
						}, func(err error) {
							logger.Errorf("Set %s as default error: %+v", v.Version.String(), err)
							p.app.Alert(fmt.Sprintf("Set %s as default failed: %+v", v.Version.String(), err), p.table)
						})
					}
				}, func() {
					// nothing to do
				})
			} else {
				installer := NewInstall(p.app, p.app.pages, p.app.lang, v.RemoteVersion, func(err error) {
					if err == nil {
						p.refresh()
					}
				})
				installer.Install()
			}
			return evt
		}, true),
		tcell.KeyCtrlD: NewKeyAction("Uninstall selected", func(evt *tcell.EventKey) *tcell.EventKey {
			vi, ok := p.GetSelection()
			if !ok {
				return evt
			}
			v := vi.(*version)
			p.app.Confirm(fmt.Sprintf("Are you sure you want to uninstall %s", v.Version.String()), func() {
				p.doAsync(fmt.Sprintf("Uninstalling %s", v.Version.String()), func() (interface{}, error) {
					return nil, p.app.lang.Uninstall(context.Background(), v.Version.String())
				}, func(i interface{}) {
					p.refresh()
				}, func(err error) {
					p.app.Alert(fmt.Sprintf("Uninstall failed: %+v", err), p.table)
				})
			}, func() {
				// nothing to do
			})
			return evt
		}, true),
	})

	return p
}

func (p *PageLanguageVersions) Init(ctx context.Context) {
	if p.app.lang == nil {
		p.app.lang = &golang.Golang{}
	}
	p.SetModel(&LanguageVersions{lang: p.app.lang})
	p.Render()
	p.refresh()
}

func (p *PageLanguageVersions) refresh() {
	p.loadLanguageVersionsAsync(p.app.lang,
		func(tabular Tabular) {
			data := tabular
			p.SetModel(data)
		}, func(err error) {
			p.app.Alert(
				fmt.Sprintf("Error getting remote versions for %s\n< %s >", p.app.lang.Name(), err.Error()),
				p.table,
			)
		})
}

func (p *PageLanguageVersions) GetKeyActions() *KeyActions {
	return p.actions
}

func (p *PageLanguageVersions) doAsync(
	loadingMsg string,
	do func() (interface{}, error),
	onSuccess func(interface{}),
	onError func(error),
) {
	loading := uuid.NewString()
	p.app.ShowLoading(loadingMsg, loading)
	go func() {
		defer func() {
			p.app.QueueUpdateDraw(func() {
				p.app.HideLoading(loading)
			})
		}()

		data, err := do()

		p.app.QueueUpdateDraw(func() {
			if err != nil {
				onError(err)
			} else {
				onSuccess(data)
			}
		})
	}()
}

func (p *PageLanguageVersions) loadLanguageVersionsAsync(
	lang core.Language,
	onSuccess func(versions Tabular),
	onError func(error),
) {
	loading := uuid.NewString()
	p.app.ShowLoading(fmt.Sprintf("Loading %s version info", lang.Name()), loading)

	go func() {
		defer func() {
			p.app.QueueUpdateDraw(func() {
				p.app.HideLoading(loading)
				p.app.SetFocus(p.table)
			})
		}()

		versions, err := p.languageVersions.GetData(lang)

		p.app.QueueUpdateDraw(func() {
			if err != nil {
				onError(err)
			} else {
				onSuccess(versions)
			}
		})
	}()
}
