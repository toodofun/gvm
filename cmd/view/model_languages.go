package view

import (
	"github.com/gdamore/tcell/v2"
	"gvm/core"
)

type Languages struct {
	languages []string
}

func NewLanguages() *Languages {
	languages := core.GetAllLanguage()
	return &Languages{
		languages: languages,
	}
}

func (l *Languages) Title() string {
	return "Global Version Manager"
}

func (l *Languages) RowCount() int {
	return len(l.languages)
}

func (l *Languages) Headers() []*TableHeader {
	return []*TableHeader{
		{
			Title:     "Language",
			Expansion: 1,
		},
	}
}

func (l *Languages) GetRow(i []string) interface{} {
	for _, v := range l.languages {
		if v == i[0] {
			return v
		}
	}
	return ""
}

func (l *Languages) GetRowColor(i []string) tcell.Color {
	return tcell.ColorSkyblue
}

func (l *Languages) Filter(k, v string) {

}

func (l *Languages) Rows() [][]string {
	res := make([][]string, 0)
	for _, lang := range l.languages {
		res = append(res, []string{lang})
	}
	return res
}
