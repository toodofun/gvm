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
