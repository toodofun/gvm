package view

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type Page interface {
	tview.Primitive
	Init()
	SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey) *tview.Box
	GetKeyActions() *KeyActions
}
