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
	"github.com/gdamore/tcell/v2"
	"sort"
	"sync"
)

type (
	ActionHandler func(key *tcell.EventKey) *tcell.EventKey
	ActionOpts    struct {
		Visible bool
		Shared  bool
	}

	KeyAction struct {
		Description string
		Action      ActionHandler
		Opts        ActionOpts
	}

	KeyMap map[tcell.Key]KeyAction

	KeyActions struct {
		actions KeyMap
		mx      sync.RWMutex
	}

	RandFn func(tcell.Key, KeyAction)
)

func NewKeyAction(d string, a ActionHandler, visible bool) KeyAction {
	return NewKeyActionWithOpts(d, a, ActionOpts{Visible: visible})
}

func NewKeyActionWithOpts(d string, a ActionHandler, opts ActionOpts) KeyAction {
	return KeyAction{
		Description: d,
		Action:      a,
		Opts:        opts,
	}
}

func NewKeyActions() *KeyActions {
	return &KeyActions{
		actions: make(map[tcell.Key]KeyAction),
	}
}

func NewKeyActionsFromMap(m KeyMap) *KeyActions {
	return &KeyActions{
		actions: m,
	}
}

func (a *KeyActions) Get(key tcell.Key) (KeyAction, bool) {
	a.mx.RLock()
	defer a.mx.RUnlock()

	v, ok := a.actions[key]

	return v, ok
}

func (a *KeyActions) Len() int {
	a.mx.RLock()
	defer a.mx.RUnlock()

	return len(a.actions)
}

func (a *KeyActions) Clear() {
	a.mx.Lock()
	defer a.mx.Unlock()

	for k := range a.actions {
		delete(a.actions, k)
	}
}

func (a *KeyActions) Add(key tcell.Key, action KeyAction) {
	a.mx.Lock()
	defer a.mx.Unlock()

	a.actions[key] = action
}

func (a *KeyActions) Delete(keys ...tcell.Key) {
	a.mx.Lock()
	defer a.mx.Unlock()

	for _, k := range keys {
		delete(a.actions, k)
	}
}

func (a *KeyActions) Range(f RandFn) {
	a.mx.RLock()
	defer a.mx.RUnlock()

	// 提取键并排序
	var keys []tcell.Key
	for k := range a.actions {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool {
		return keys[i] < keys[j]
	})

	// 按顺序遍历
	for _, k := range keys {
		f(k, a.actions[k])
	}
}

//func (a *KeyActions) Range(f RandFn) {
//	var km KeyMap
//	a.mx.RLock()
//	km = a.actions
//	a.mx.RUnlock()
//
//	for k, v := range km {
//		f(k, v)
//	}
//}

func (a *KeyActions) Merge(as *KeyActions) {
	a.mx.Lock()
	defer a.mx.Unlock()

	for k, v := range as.actions {
		a.actions[k] = v
	}
}
