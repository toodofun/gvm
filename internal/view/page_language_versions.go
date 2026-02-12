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

	"github.com/toodofun/gvm/i18n"

	"github.com/toodofun/gvm/internal/core"
	"github.com/toodofun/gvm/internal/log"
	"github.com/toodofun/gvm/languages/golang"

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
		KeyI: NewKeyAction(
			i18n.GetTranslate("page.languageVersion.keyAction.i", nil),
			func(evt *tcell.EventKey) *tcell.EventKey {
				model := p.GetModel()
				model.Filter("installed", "")
				p.SetModel(model)
				p.Render()
				return evt
			},
			true,
		),
		tcell.KeyESC: NewKeyAction(
			i18n.GetTranslate("page.languageVersion.keyAction.esc", nil),
			func(evt *tcell.EventKey) *tcell.EventKey {
				p.app.SwitchPage(pageLanguages)
				return evt
			},
			true,
		),
		tcell.KeyEnter: NewKeyAction(
			i18n.GetTranslate("page.languageVersion.keyAction.enter", nil),
			func(evt *tcell.EventKey) *tcell.EventKey {
				vi, ok := p.GetSelection()
				if !ok {
					return evt
				}
				v := vi.(*version)
				if v.isInstalled {
					p.app.Confirm(
						i18n.GetTranslate("languages.setDefaultInfo", map[string]interface{}{
							"version": v.Version.String(),
						}),
						func() {
							if v.isDefault {
								p.app.Alert(
									i18n.GetTranslate("languages.alreadyDefault", map[string]interface{}{
										"version": v.Version.String(),
									}),
									p.table,
								)
							} else {
								p.doAsync(i18n.GetTranslate("languages.setDefault", map[string]interface{}{
									"version": v.Version.String(),
								}), func() (interface{}, error) {
									if err := p.app.lang.SetDefaultVersion(p.app.ctx, v.Version.String()); err != nil {
										return nil, err
									}
									return nil, nil
								}, func(i interface{}) {
									p.refresh()
								}, func(err error) {
									logger.Error(i18n.GetTranslate("languages.setDefaultError", map[string]interface{}{
										"version": v.Version.String(),
										"error":   err.Error(),
									}))
									p.app.Alert(i18n.GetTranslate("languages.setDefaultError", map[string]interface{}{
										"version": v.Version.String(),
										"error":   err.Error(),
									}), p.table)
								})
							}
						},
						func() {
							// nothing to do
						},
					)
				} else {
					installer := NewInstall(p.app, p.app.pages, p.app.lang, v.RemoteVersion, func(err error) {
						if err == nil {
							p.refresh()
						}
					})
					installer.Install()
				}
				return evt
			},
			true,
		),
		tcell.KeyCtrlD: NewKeyAction(
			i18n.GetTranslate("page.languageVersion.keyAction.ctrlD", nil),
			func(evt *tcell.EventKey) *tcell.EventKey {
				vi, ok := p.GetSelection()
				if !ok {
					return evt
				}
				v := vi.(*version)
				p.app.Confirm(i18n.GetTranslate("languages.uninstallInfo", map[string]any{
					"version": v.Version.String(),
				}), func() {
					p.doAsync(i18n.GetTranslate("languages.uninstall", map[string]any{
						"version": v.Version.String(),
					}), func() (interface{}, error) {
						return nil, p.app.lang.Uninstall(p.app.ctx, v.Version.String())
					}, func(i interface{}) {
						p.refresh()
					}, func(err error) {
						p.app.Alert(i18n.GetTranslate("languages.uninstallError", map[string]any{
							"version": v.Version.String(),
							"error":   err.Error(),
						}), p.table)
					})
				}, func() {
					// nothing to do
				})
				return evt
			},
			true,
		),
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
	p.app.ShowLoading(lang.Name(), loading)

	go func() {
		defer func() {
			p.app.QueueUpdateDraw(func() {
				p.app.HideLoading(loading)
				p.app.SetFocus(p.table)
			})
		}()

		versions, err := p.languageVersions.GetData(p.app.ctx, lang)

		p.app.QueueUpdateDraw(func() {
			if err != nil {
				onError(err)
			} else {
				onSuccess(versions)
			}
		})
	}()
}
