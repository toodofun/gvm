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
	"testing"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/stretchr/testify/assert"
)

// mockTabular 实现了 Tabular 接口，用于测试
type mockTabular struct {
	headers []*TableHeader
	rows    [][]string
	title   string
}

func (m *mockTabular) Title() string {
	return m.title
}
func (m *mockTabular) RowCount() int {
	return len(m.rows)
}
func (m *mockTabular) Headers() []*TableHeader {
	return m.headers
}
func (m *mockTabular) Rows() [][]string {
	return m.rows
}
func (m *mockTabular) GetRow(row []string) interface{} {
	return row
}
func (m *mockTabular) GetRowColor(row []string) tcell.Color {
	return tcell.ColorWhite
}
func (m *mockTabular) Filter(field, value string) {
	// not needed for basic tests
}

func newMockTabular() *mockTabular {
	return &mockTabular{
		title: "MockTable",
		headers: []*TableHeader{
			{Title: "Name", Expansion: 1},
			{Title: "Age", Expansion: 1},
		},
		rows: [][]string{
			{"Alice", "30"},
			{"Bob", "25"},
			{"Charlie", "40"},
		},
	}
}

func newMockApp() *Application {
	return &Application{
		Application: tview.NewApplication(),
	}
}

func TestNewSearchTable(t *testing.T) {
	app := newMockApp()
	model := newMockTabular()
	st := NewSearchTable(model, app)

	assert.NotNil(t, st)
	assert.Equal(t, model, st.GetModel())
	assert.Equal(t, 3, len(st.rows))
}

func TestSearchTable_SetModel(t *testing.T) {
	app := newMockApp()
	model := newMockTabular()
	st := NewSearchTable(model, app)

	newModel := newMockTabular()
	newModel.rows = [][]string{{"Dave", "50"}}
	st.SetModel(newModel)

	assert.Equal(t, newModel, st.GetModel())
	assert.Equal(t, 1, len(st.rows))
}

func TestSearchTable_GetSelection(t *testing.T) {
	app := newMockApp()
	model := newMockTabular()
	st := NewSearchTable(model, app)

	st.Select(1, 0) // select first data row
	val, ok := st.GetSelection()
	assert.True(t, ok)
	assert.Equal(t, []string{"Alice", "30"}, val)

	st.Select(0, 0) // select header row
	_, ok = st.GetSelection()
	assert.False(t, ok)
}

func TestSearchTable_Search(t *testing.T) {
	app := newMockApp()
	model := newMockTabular()
	st := NewSearchTable(model, app)

	st.search("Bob")
	assert.Equal(t, 1, len(st.rows))
	assert.Equal(t, "Bob", st.rows[0][0])

	st.search("")
	assert.Equal(t, 3, len(st.rows))
}

func TestFixedWidth(t *testing.T) {
	assert.Equal(t, "abc  ", FixedWidth("abc", 5))
	assert.Equal(t, "abc", FixedWidth("abcdef", 3))
	assert.Equal(t, "abc", FixedWidth("abc", 3))
}
