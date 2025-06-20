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
	"gvm/core"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type Application struct {
	*tview.Application
	help    *tview.Flex
	pages   *tview.Pages
	lang    core.Language
	pageMap map[string]Page
	actions *KeyActions
}

func CreateApplication() *Application {
	tview.Borders = Borders
	app := tview.NewApplication()

	return &Application{
		Application: app,
		actions:     NewKeyActions(),
	}
}

func (a *Application) AsKey(evt *tcell.EventKey) tcell.Key {
	if evt.Key() != tcell.KeyRune {
		return evt.Key()
	}

	key := tcell.Key(evt.Rune())
	if evt.Modifiers() == tcell.ModAlt {
		key = tcell.Key(int16(evt.Rune()) * int16(evt.Modifiers()))
	}
	return key
}

func (a *Application) HasAction(key tcell.Key) (KeyAction, bool) {
	return a.actions.Get(key)
}

func (a *Application) SwitchPage(page string) {
	p := a.pageMap[page]
	a.pages.SwitchToPage(page)
	p.Init()
	a.readerHelp(p.GetKeyActions())
}

func (a *Application) createMain() tview.Primitive {
	main := tview.NewFlex().SetDirection(tview.FlexRow)
	pages := tview.NewPages()

	a.pageMap = map[string]Page{
		pageLanguages:        NewPageLanguages(a),
		pageLanguageVersions: NewPageLanguageVersions(a),
	}

	for name, page := range a.pageMap {
		pages.AddPage(name, page, true, false)
	}

	main.AddItem(pages, 0, 1, true)
	a.pages = pages
	a.SwitchPage(pageLanguages)

	return main
}

func (a *Application) quitCmd(evt *tcell.EventKey) *tcell.EventKey {
	a.Stop()
	return nil
}

func (a *Application) IsTopDialog() bool {
	return a.pages.HasPage(alertKey) || a.pages.HasPage(confirmKey)
}

func (a *Application) bindKeys() {
	a.actions.Merge(NewKeyActionsFromMap(KeyMap{
		tcell.KeyCtrlC: NewKeyAction("Quit", a.quitCmd, true),
	}))
}

func (a *Application) bindKey(evt *tcell.EventKey) *tcell.EventKey {
	if k, ok := a.HasAction(a.AsKey(evt)); ok && !a.IsTopDialog() {
		return k.Action(evt)
	}
	return evt
}

func (a *Application) Run() error {
	a.SetInputCapture(a.bindKey)
	a.bindKeys()
	// 创建主布局
	root := tview.NewFlex().SetDirection(tview.FlexRow)
	root.AddItem(a.createHeader(), 6, 0, false)
	root.AddItem(a.createMain(), 0, 1, true)

	return a.SetRoot(root, true).EnableMouse(false).Run()
}
