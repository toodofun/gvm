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
