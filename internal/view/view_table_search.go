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
	"strings"

	"github.com/duke-git/lancet/v2/slice"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type SearchTable struct {
	*tview.Flex

	app     *Application
	table   *tview.Table
	actions *KeyActions

	condition string

	model Tabular
	rows  [][]string
}

func NewSearchTable(model Tabular, app *Application) *SearchTable {
	root := tview.NewFlex().SetDirection(tview.FlexRow)
	table := tview.NewTable()
	root.AddItem(table, 0, 1, true)
	t := &SearchTable{
		Flex:    root,
		table:   table,
		app:     app,
		actions: NewKeyActions(),
		model:   model,
		rows:    model.Rows(),
	}

	t.table.SetInputCapture(t.bindKey)

	t.init()

	return t
}

func (t *SearchTable) bindKey(evt *tcell.EventKey) *tcell.EventKey {
	key := evt.Key()
	if key == tcell.KeyUp || key == tcell.KeyDown {
		return evt
	}

	if a, ok := t.actions.Get(t.app.AsKey(evt)); ok && !t.app.IsTopDialog() {
		return a.Action(evt)
	}
	return evt
}

func (t *SearchTable) SetModel(model Tabular) {
	t.model = model
	t.rows = model.Rows()
}

func (t *SearchTable) GetSelection() interface{} {
	r, _ := t.table.GetSelection()
	return t.model.GetRow(t.rows[r-1])
}

func (t *SearchTable) GetModel() Tabular {
	return t.model
}

func (t *SearchTable) Select(row, column int) {
	t.table.Select(row, column)
}

func (t *SearchTable) GetRowCount() int {
	return t.table.GetRowCount()
}

func (t *SearchTable) BindKeys(km KeyMap) {
	km[KeyColon] = NewKeyAction("Command", func(evt *tcell.EventKey) *tcell.EventKey {
		input := tview.NewInputField()
		input.SetBorder(true).
			SetBorderColor(tcell.ColorGreen).
			SetBorderAttributes(tcell.AttrDim).
			SetBorderPadding(0, 0, 1, 1)
		input.SetFieldBackgroundColor(tcell.ColorBlack).
			SetFieldTextColor(tcell.ColorWhite).
			SetLabel("💻 ").SetLabelColor(tcell.ColorBlueViolet)
		input.SetDoneFunc(func(key tcell.Key) {
			if key == tcell.KeyEnter {
				t.app.SetFocus(t.table)
				t.command(input.GetText())
				t.RemoveItem(input)
			}
		})
		input.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
			switch event.Key() {
			case tcell.KeyESC:
				t.app.SetFocus(t.table)
				t.command("")
				t.RemoveItem(input)
			default:
				// nothing to do
			}
			return event
		})
		t.AddItem(input, 3, 0, true)
		t.RemoveItem(t.table)
		t.AddItem(t.table, 0, 1, false)
		t.app.SetFocus(input)
		return evt
	}, true)
	km[KeySlash] = NewKeyAction("Search", func(evt *tcell.EventKey) *tcell.EventKey {
		input := tview.NewInputField()
		input.SetBorder(true).
			SetBorderColor(tcell.ColorGreen).
			SetBorderAttributes(tcell.AttrDim).
			SetBorderPadding(0, 0, 1, 1)
		input.SetFieldBackgroundColor(tcell.ColorBlack).
			SetFieldTextColor(tcell.ColorWhite).
			SetLabel("🔭 ").SetLabelColor(tcell.ColorBlueViolet)

		input.SetChangedFunc(func(text string) {
			t.search(text)
		})
		input.SetDoneFunc(func(key tcell.Key) {
			if key == tcell.KeyEnter {
				t.RemoveItem(input)
				t.app.SetFocus(t.table)
			}
		})
		input.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
			switch event.Key() {
			case tcell.KeyESC:
				t.search("")
				t.RemoveItem(input)
				t.app.SetFocus(t.table)
			default:
				// nothing to do
			}
			return event
		})
		t.search("")
		t.AddItem(input, 3, 0, true)
		t.RemoveItem(t.table)
		t.AddItem(t.table, 0, 1, false)
		t.app.SetFocus(input)
		return evt
	}, true)
	t.actions.Merge(NewKeyActionsFromMap(km))
}

func (t *SearchTable) init() {
	t.table.SetFixed(1, 0)
	t.table.SetBorder(true).
		SetBorderAttributes(tcell.AttrNone).
		SetBorderColor(tcell.ColorSkyblue)
	t.table.SetBorderPadding(0, 0, 1, 1)
	t.table.SetSelectable(true, false)
	t.table.SetSelectedStyle(tcell.Style{}.Background(tcell.ColorSkyblue).Foreground(tcell.ColorBlack))
	t.setTitle()
	t.Render()
}

func (t *SearchTable) setTitle() {
	// 设置标题
	if len(t.condition) > 0 {
		t.table.SetTitle(
			fmt.Sprintf(
				" [aqua::b]%s[-:-:-] [skyblue][%d][-] </%s> ",
				t.model.Title(),
				t.model.RowCount(),
				t.condition,
			),
		)
	} else {
		t.table.SetTitle(fmt.Sprintf(" [aqua::b]%s[-:-:-] [skyblue][%d][-] ", t.model.Title(), t.model.RowCount()))
	}
}

func (t *SearchTable) command(cmd string) {
	if len(cmd) == 0 {
		return
	}
	switch cmd {
	case "q":
		t.app.Stop()
	default:
		t.app.Alert(fmt.Sprintf("command `%s` not found", cmd), t.table)
	}
}

func (t *SearchTable) search(condition string) {
	t.condition = condition
	if len(condition) == 0 {
		t.rows = t.model.Rows()
	} else {
		tmp := t.model.Rows()
		t.rows = slice.Filter(tmp, func(index int, item []string) bool {
			for _, row := range item {
				if strings.Contains(strings.ToLower(row), strings.ToLower(condition)) {
					return true
				}
			}
			return false
		})
	}
	t.setTitle()
	t.Render()
}

func (t *SearchTable) Render() {
	t.setTitle()
	t.table.Clear()

	// 设置表格头
	for i, h := range t.model.Headers() {
		if h.Hide {
			continue
		}
		if h.FixedWidth > 0 {
			h.Title = FixedWidth(h.Title, h.FixedWidth)
		}
		cell := tview.NewTableCell(h.Title).
			SetExpansion(h.Expansion).
			SetSelectable(false)

		t.table.SetCell(0, i, cell)
	}

	for i, row := range t.rows {
		for j, cell := range row {
			if t.model.Headers()[j].Hide {
				continue
			}
			if t.model.Headers()[j].FixedWidth > 0 {
				cell = FixedWidth(cell, t.model.Headers()[j].FixedWidth)
			}
			c := tview.NewTableCell(cell).SetTextColor(t.model.GetRowColor(row)).
				SetExpansion(t.model.Headers()[j].Expansion)

			t.table.SetCell(i+1, j, c)
		}
	}

	t.Select(1, 0)
	t.table.SetOffset(0, 0)
}

func FixedWidth(text string, width int) string {
	runes := []rune(text)
	if len(runes) > width {
		return string(runes[:width])
	}
	return text + strings.Repeat(" ", width-len(runes))
}
