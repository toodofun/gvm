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
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

func (a *Application) Confirm(msg string, onConfirm, onCancel func()) {
	modal := tview.NewModal()
	modal.Box.SetBackgroundColor(tcell.ColorBlack)
	modal.SetBackgroundColor(tcell.ColorBlack)
	modal.SetText(msg)
	modal.SetTextColor(tcell.ColorBlue)
	modal.SetBorder(true).
		SetTitle(fmt.Sprintf(" [blue]%s[-:-:-] ", "< confirm >")).
		SetBorderColor(tcell.ColorBlue)
	modal.AddButtons([]string{"Confirm", "Cancel"})
	modal.SetButtonBackgroundColor(tcell.ColorGray)
	modal.SetButtonTextColor(tcell.ColorBlack)
	modal.SetButtonActivatedStyle(tcell.StyleDefault.Foreground(tcell.ColorBlack).Background(tcell.ColorBlue))
	modal.SetFocus(0)

	modal.SetDoneFunc(func(buttonIndex int, buttonLabel string) {
		if buttonLabel == "Confirm" {
			onConfirm()
		} else {
			onCancel()
		}
		a.pages.RemovePage(confirmKey)
	})

	a.pages.AddPage(confirmKey, modal, false, true)
}

func (a *Application) Info(msg string, after tview.Primitive) {
	modal := tview.NewModal()
	modal.Box.SetBackgroundColor(tcell.ColorBlack)
	modal.SetBackgroundColor(tcell.ColorBlack)
	modal.SetText(msg)
	modal.SetBorder(true).
		SetTitle(fmt.Sprintf(" [blue]%s[-:-:-] ", "< info >")).
		SetBorderColor(tcell.ColorBlue)
	modal.AddButtons([]string{"Dismiss"})
	modal.SetButtonBackgroundColor(tcell.ColorGray)
	modal.SetButtonTextColor(tcell.ColorBlack)
	modal.SetButtonActivatedStyle(tcell.StyleDefault.Foreground(tcell.ColorBlack).Background(tcell.ColorBlue))

	modal.SetDoneFunc(func(buttonIndex int, buttonLabel string) {
		a.pages.RemovePage(infoKey)
		a.Application.SetFocus(after)
	})

	a.pages.AddPage(infoKey, modal, false, true)
}

func (a *Application) Alert(msg string, after tview.Primitive) {
	modal := tview.NewModal()
	modal.Box.SetBackgroundColor(tcell.ColorBlack)
	modal.SetBackgroundColor(tcell.ColorBlack)
	modal.SetText(fmt.Sprintf("< %s >\n%s", msg, errorMsg))
	modal.SetTextColor(tcell.ColorRed)
	modal.SetBorder(true).
		SetTitle(fmt.Sprintf(" [red]%s[-:-:-] ", "< error >")).
		SetBorderColor(tcell.ColorBlue)
	modal.AddButtons([]string{"Dismiss"})
	modal.SetButtonBackgroundColor(tcell.ColorGray)
	modal.SetButtonTextColor(tcell.ColorBlack)
	modal.SetButtonActivatedStyle(tcell.StyleDefault.Foreground(tcell.ColorBlack).Background(tcell.ColorBlue))

	modal.SetDoneFunc(func(buttonIndex int, buttonLabel string) {
		a.pages.RemovePage(alertKey)
		a.Application.SetFocus(after)
	})

	a.pages.AddPage(alertKey, modal, false, true)
}

// ShowLoading Loading（带动画）
func (a *Application) ShowLoading(message string, name string) {
	modal := tview.NewModal()
	modal.Box.SetBackgroundColor(tcell.ColorBlack)
	modal.SetBackgroundColor(tcell.ColorBlack)
	modal.SetBorder(true)
	modal.SetTitle(" [blue]Loading[-] ")
	modal.SetBorderColor(tcell.ColorBlue)

	// 动画字符
	spinners := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
	spinnerIndex := 0

	// 设置初始文本
	modal.SetText(fmt.Sprintf("[yellow]%s %s...\n\n[blue]Just a moment...[-:-:-]", spinners[0], message))

	// 创建一个 goroutine 来更新动画
	go func() {
		ticker := time.NewTicker(100 * time.Millisecond)
		defer ticker.Stop()
		for range ticker.C {
			if !a.pages.HasPage(name) {
				// 页面已关闭，退出
				break
			}

			a.QueueUpdateDraw(func() {
				if a.pages.HasPage(name) {
					spinner := spinners[spinnerIndex%len(spinners)]
					modal.SetText(fmt.Sprintf("[yellow]%s %s...\n\n[blue]Just a moment...[-:-:-]", spinner, message))
					spinnerIndex++
				}
			})
		}
	}()

	a.pages.AddPage(name, modal, false, true)
}

// HideLoading 隐藏Loading
func (a *Application) HideLoading(name string) {
	a.pages.RemovePage(name)
}
