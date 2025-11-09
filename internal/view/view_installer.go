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
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/toodofun/gvm/i18n"

	"github.com/toodofun/gvm/internal/core"
	"github.com/toodofun/gvm/internal/log"
	"github.com/toodofun/gvm/languages"

	"github.com/gdamore/tcell/v2"
	goversion "github.com/hashicorp/go-version"
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
	installer.SetTitle(
		fmt.Sprintf("[blue] "+i18n.GetTranslate("page.installer.install", nil)+" %s [-:-:-]", version.Version.String()),
	)
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
					// 格式化文本，确保显示正确
					formattedText := i.formatDisplayText(newText)
					i.app.QueueUpdateDraw(func() {
						i.SetText(formattedText)
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

// formatDisplayText 格式化显示文本，避免重复和混乱
func (i *Installer) formatDisplayText(text string) string {
	if text == "" {
		return ""
	}

	// 分割为行
	lines := strings.Split(text, "\n")
	uniqueLines := make([]string, 0, len(lines))
	seen := make(map[string]bool)

	// 去重并保持顺序
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed != "" && !seen[trimmed] {
			uniqueLines = append(uniqueLines, trimmed)
			seen[trimmed] = true
		}
	}

	// 限制最多显示5行，避免界面过载
	maxLines := 5
	if len(uniqueLines) > maxLines {
		uniqueLines = uniqueLines[len(uniqueLines)-maxLines:]
	}

	return strings.Join(uniqueLines, "\n")
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
		var preReleaseErr *languages.PreReleaseError
		if err != nil {
			i.write("install failed: " + err.Error())
		} else {
			i.write("Installation completed successfully!")
		}

		// 等待一下确保最后的消息显示
		time.Sleep(100 * time.Millisecond)

		// 安装完成后设置按钮和事件处理
		i.app.QueueUpdateDraw(func() {
			// 检查是否是预发布版本错误
			if errors.As(err, &preReleaseErr) && preReleaseErr.GetRecommendedVersion() != "" {
				// 提供两个选项：安装推荐版本或取消
				i.SetDoneFunc(func(buttonIndex int, buttonLabel string) {
					if buttonIndex == 0 { // 安装推荐版本
						// 创建新的版本对象
						recommendedVer, verErr := goversion.NewVersion(preReleaseErr.GetRecommendedVersion())
						if verErr == nil {
							newVersion := &core.RemoteVersion{
								Version: recommendedVer,
								Origin:  preReleaseErr.GetRecommendedVersion(),
								Comment: "",
							}
							// 移除当前页面
							i.pages.RemovePage(pageInstaller)
							// 创建新的安装器安装推荐版本
							newInstaller := NewInstall(i.app, i.pages, i.lang, newVersion, i.callback)
							newInstaller.Install()
						} else {
							i.pages.RemovePage(pageInstaller)
							i.callback(err)
						}
					} else { // 取消
						i.pages.RemovePage(pageInstaller)
						i.callback(err)
					}
				})
				i.AddButtons([]string{"安装 " + preReleaseErr.GetRecommendedVersion(), "取消"})
			} else {
				// 普通错误，只显示 OK 按钮
				i.SetDoneFunc(func(buttonIndex int, buttonLabel string) {
					i.pages.RemovePage(pageInstaller)
					i.callback(err)
				})
				i.AddButtons([]string{"OK"})
			}

			// 确保 Modal 获得焦点
			i.app.SetFocus(i)
		})
	}()
}
