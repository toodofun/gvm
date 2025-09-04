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
	"os"
	"os/user"
	"runtime"
	"strings"

	"github.com/toodofun/gvm/internal/log"
	v "github.com/toodofun/gvm/internal/util/version"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

func (a *Application) createHeader(ctx context.Context) tview.Primitive {
	logger := log.GetLogger(ctx)
	u, err := user.Current()
	if err != nil {
		u = &user.User{Username: "unknow"}
	}

	hn, err := os.Hostname()
	if err != nil {
		hn = "unknow"
	}

	type KV struct {
		Key   string
		Value string
	}
	descMap := []KV{
		{Key: "[yellow] Hostname[-:-:-]", Value: hn},
		{Key: "[yellow] System[-:-:-]", Value: fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH)},
		{Key: "[yellow] GVM Rev.[-:-:-]", Value: v.Get().GitVersion},
		{Key: "[yellow] Username[-:-:-]", Value: u.Username},
		{Key: "[yellow] Loglevel[-:-:-]", Value: log.GetLevel()},
	}

	desc := tview.NewTable().
		SetBorders(false)
	for i, kv := range descMap {
		desc.SetCell(i, 0, tview.NewTableCell(kv.Key))
		desc.SetCell(i, 1, tview.NewTableCell(kv.Value))
	}

	// 异步检查更新
	go func() {
		has, latest := v.CheckUpdate(ctx)
		if has {
			a.QueueUpdateDraw(func() {
				newRow := desc.GetRowCount()
				desc.SetCell(newRow, 0, tview.NewTableCell("[yellow] New Ver.[-:-:-]"))
				desc.SetCell(newRow, 1, tview.NewTableCell("[blue]"+latest+"❗️[-:-:-]"))
			})
		}
	}()

	a.help = tview.NewFlex()

	logoWidth := len(strings.Split(logo, "\n")[1]) + 1
	logger.Debugf("logo width: %d", logoWidth)

	header := tview.NewFlex().
		AddItem(desc, 38, 0, false).
		AddItem(a.help, 0, 1, false).
		AddItem(tview.NewTextView().SetText(logo).SetTextColor(tcell.ColorYellow), logoWidth, 0, false)

	return header
}

func (a *Application) readerHelp(kas *KeyActions) {
	a.help.Clear()

	go func() {
		a.QueueUpdateDraw(func() {
			allActions := NewKeyActions()
			allActions.Merge(kas)
			allActions.Merge(a.actions)

			const maxPerColumn = 6
			col := 0
			row := 0
			table := tview.NewTable()

			allActions.Range(func(key tcell.Key, action KeyAction) {
				if !action.Opts.Visible {
					return
				}

				keyName := strings.ToLower(tcell.KeyNames[key])
				if len(action.Opts.DisplayName) > 0 {
					keyName = action.Opts.DisplayName
				}
				table.SetCell(
					row,
					col*2,
					tview.NewTableCell(fmt.Sprintf("[skyblue]<%s>[-:-:-]", keyName)).SetMaxWidth(12),
				)
				table.SetCell(
					row,
					col*2+1,
					tview.NewTableCell(fmt.Sprintf("[gray]%s[-:-:-]", action.Description)).SetMaxWidth(28),
				)

				row++
				if row >= maxPerColumn {
					row = 0
					col++
				}
			})

			a.help.AddItem(table, 0, 1, false)
		})
	}()
}
