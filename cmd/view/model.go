package view

import "github.com/gdamore/tcell/v2"

type Tabular interface {
	Title() string
	RowCount() int
	Headers() []*TableHeader
	Rows() [][]string
	GetRow([]string) interface{}
	GetRowColor([]string) tcell.Color
	Filter(string, string)
}

type TableHeader struct {
	Title      string // 名称
	FixedWidth int
	Expansion  int // 占的列宽比例
	Hide       bool
}
