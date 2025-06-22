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
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type Installer struct {
	*tview.Modal
	app      *Application
	pages    *tview.Pages
	buf      *log.NotifyBuffer
	lang     core.Language
	version  *core.RemoteVersion
	callback func(error)
	ticker   *time.Ticker
	stopChan chan struct{}
	lastText string
}

func NewInstall(
	app *Application,
	pages *tview.Pages,
	lang core.Language,
	version *core.RemoteVersion,
	callback func(err error),
) *Installer {
	installer := &Installer{
		Modal:    tview.NewModal(),
		app:      app,
		pages:    pages,
		lang:     lang,
		buf:      log.NewNotifyBuffer(),
		version:  version,
		callback: callback,
		stopChan: make(chan struct{}),
	}

	installer.SetBorder(true)
	installer.SetTitle(fmt.Sprintf("[blue] Install %s [-:-:-]", version.Version.String()))
	installer.Box.SetBackgroundColor(tcell.ColorBlack)
	installer.SetBackgroundColor(tcell.ColorBlack)
	installer.SetBorderColor(tcell.ColorBlue)
	installer.SetButtonBackgroundColor(tcell.ColorGray)
	installer.SetButtonTextColor(tcell.ColorBlack)
	installer.SetButtonActivatedStyle(tcell.StyleDefault.Foreground(tcell.ColorBlack).Background(tcell.ColorBlue))

	return installer
}

func (i *Installer) show() {
	// 使用定时器控制更新频率，避免过于频繁的界面刷新
	i.ticker = time.NewTicker(50 * time.Millisecond) // 50ms更新一次，保持流畅

	go func() {
		defer i.ticker.Stop()

		for {
			select {
			case <-i.buf.Updated:
				// 收到更新信号，等待下一次定时器触发
				continue
			case <-i.ticker.C:
				// 检查是否有新内容需要更新
				newText := i.buf.Read()
				if newText != i.lastText && newText != "" {
					i.lastText = newText
					i.app.QueueUpdateDraw(func() {
						i.SetText(newText)
					})
				}
			case <-i.stopChan:
				return
			}
		}
	}()
}

func (i *Installer) write(msg string) {
	i.buf.Write([]byte(msg))
}

func (i *Installer) Install() {
	i.show()

	// 先添加页面
	i.pages.AddPage(pageInstaller, i, false, true)
	ctx := context.Background()
	ctx = context.WithValue(ctx, core.ContextLogWriterKey, i.buf)

	go func() {
		defer func() {
			close(i.stopChan) // 停止更新协程
			i.buf.Close()     // 关闭缓冲区
		}()

		err := i.lang.Install(ctx, i.version)
		if err != nil {
			i.write("install failed: " + err.Error())
		} else {
			i.write("Installation completed successfully!")
		}

		// 等待一下确保最后的消息显示
		time.Sleep(100 * time.Millisecond)

		// 安装完成后设置按钮和事件处理
		i.app.QueueUpdateDraw(func() {
			// 先设置事件处理函数
			i.SetDoneFunc(func(buttonIndex int, buttonLabel string) {
				// 移除页面
				i.pages.RemovePage(pageInstaller)
				i.callback(err)
			})

			// 再添加按钮
			i.AddButtons([]string{"OK"})

			// 确保 Modal 获得焦点
			i.app.SetFocus(i)
		})
	}()
}
