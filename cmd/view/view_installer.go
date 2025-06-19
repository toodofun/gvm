package view

import (
	"fmt"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/sirupsen/logrus"
	"gvm/core"
	"gvm/internal/log"
)

type Installer struct {
	*tview.Modal
	app      *Application
	pages    *tview.Pages
	buf      *log.NotifyBuffer
	lang     core.Language
	version  *core.RemoteVersion
	callback func(error)
}

func (i *Installer) show() {
	go func() {
		for range i.buf.Updated {
			i.app.QueueUpdateDraw(func() {
				i.SetText(i.buf.ReadAndReset())
			})
		}
	}()
}

func NewInstall(app *Application, pages *tview.Pages, lang core.Language, version *core.RemoteVersion, callback func(err error)) *Installer {
	installer := &Installer{
		Modal:    tview.NewModal(),
		app:      app,
		pages:    pages,
		lang:     lang,
		buf:      log.NewNotifyBuffer(),
		version:  version,
		callback: callback,
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

func (i *Installer) write(msg string) {
	i.buf.Write([]byte(msg))
}

type PlainFormatter struct{}

func (f *PlainFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	return []byte(entry.Message + "\n"), nil
}

func (i *Installer) Install() {
	i.show()
	log.SwitchTo(i.buf)
	log.Logger.SetFormatter(&PlainFormatter{})

	// 先添加页面
	i.pages.AddPage(pageInstaller, i, false, true)

	go func() {
		defer log.SwitchTo(log.Writer)
		err := i.lang.Install(i.version)
		if err != nil {
			i.write("install failed: " + err.Error())
		}

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
