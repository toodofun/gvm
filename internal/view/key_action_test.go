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

package view_test

import (
	"testing"

	"github.com/toodofun/gvm/internal/view"

	"github.com/gdamore/tcell/v2"
	"github.com/stretchr/testify/assert"
)

func dummyHandler(evt *tcell.EventKey) *tcell.EventKey {
	return evt
}

func TestNewKeyAction(t *testing.T) {
	action := view.NewKeyAction("test", dummyHandler, true, view.WithDisplayName("Test Key"), view.WithDefault())

	assert.Equal(t, "test", action.Description)
	assert.Equal(t, true, action.Opts.Visible)
	assert.Equal(t, "Test Key", action.Opts.DisplayName)
	assert.Equal(t, true, action.Opts.Default)
	assert.NotNil(t, action.Action)
}

func TestKeyActions_AddAndGet(t *testing.T) {
	ka := view.NewKeyActions()
	action := view.NewKeyAction("action", dummyHandler, true)
	ka.Add(tcell.KeyEnter, action)

	got, ok := ka.Get(tcell.KeyEnter)
	assert.True(t, ok)
	assert.Equal(t, "action", got.Description)
}

func TestKeyActions_GetDefaultFalse(t *testing.T) {
	ka := view.NewKeyActions()
	action := view.NewKeyAction("default", dummyHandler, true, view.WithDefault())
	ka.Add(tcell.KeyCtrlC, action)

	_, ok := ka.Get(tcell.KeyCtrlC)
	assert.False(t, ok)
}

func TestKeyActions_Delete(t *testing.T) {
	ka := view.NewKeyActions()
	action := view.NewKeyAction("delete", dummyHandler, true)
	ka.Add(tcell.KeyTab, action)

	ka.Delete(tcell.KeyTab)
	_, ok := ka.Get(tcell.KeyTab)
	assert.False(t, ok)
}

func TestKeyActions_Len(t *testing.T) {
	ka := view.NewKeyActions()
	ka.Add(tcell.KeyEnter, view.NewKeyAction("enter", dummyHandler, true))
	ka.Add(tcell.KeyTab, view.NewKeyAction("tab", dummyHandler, true))
	assert.Equal(t, 2, ka.Len())
}

func TestKeyActions_Clear(t *testing.T) {
	ka := view.NewKeyActions()
	ka.Add(tcell.KeyEnter, view.NewKeyAction("clear", dummyHandler, true))
	ka.Clear()
	assert.Equal(t, 0, ka.Len())
}

func TestKeyActions_Merge(t *testing.T) {
	a1 := view.NewKeyActions()
	a2 := view.NewKeyActions()

	a1.Add(tcell.KeyEnter, view.NewKeyAction("a1", dummyHandler, true))
	a2.Add(tcell.KeyTab, view.NewKeyAction("a2", dummyHandler, true))

	a1.Merge(a2)

	_, ok1 := a1.Get(tcell.KeyEnter)
	_, ok2 := a1.Get(tcell.KeyTab)

	assert.True(t, ok1)
	assert.True(t, ok2)
}

func TestKeyActions_Range(t *testing.T) {
	ka := view.NewKeyActions()
	ka.Add(tcell.KeyCtrlA, view.NewKeyAction("a", dummyHandler, true, view.WithDisplayName("Alpha")))
	ka.Add(tcell.KeyCtrlB, view.NewKeyAction("b", dummyHandler, true))

	var visited []string
	ka.Range(func(key tcell.Key, action view.KeyAction) {
		visited = append(visited, action.Description)
	})

	assert.Contains(t, visited, "a")
	assert.Contains(t, visited, "b")
}
