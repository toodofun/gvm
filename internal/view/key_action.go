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
	"sort"
	"strings"
	"sync"

	"github.com/gdamore/tcell/v2"
)

type (
	ActionHandler func(key *tcell.EventKey) *tcell.EventKey
	ActionOpts    struct {
		DisplayName string
		Visible     bool
		Shared      bool
		Default     bool // 是否只显示快捷键而不注册方法[适用于tview默认快捷键展示]
	}

	ActionOptsFn func(opts *ActionOpts)

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

var ActionNil = func(evt *tcell.EventKey) *tcell.EventKey { return evt }

func NewKeyAction(d string, a ActionHandler, visible bool, opts ...ActionOptsFn) KeyAction {
	actionOpts := ActionOpts{
		Visible: visible,
	}
	for _, opt := range opts {
		opt(&actionOpts)
	}
	return NewKeyActionWithOpts(d, a, actionOpts)
}

func NewKeyActionWithOpts(d string, a ActionHandler, opts ActionOpts, optsFn ...ActionOptsFn) KeyAction {
	for _, fn := range optsFn {
		fn(&opts)
	}
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

func WithDisplayName(displayName string) ActionOptsFn {
	return func(opts *ActionOpts) {
		opts.DisplayName = displayName
	}
}

func WithDefault() ActionOptsFn {
	return func(opts *ActionOpts) {
		opts.Default = true
	}
}

func (a *KeyActions) Get(key tcell.Key) (KeyAction, bool) {
	a.mx.RLock()
	defer a.mx.RUnlock()

	v, ok := a.actions[key]

	if v.Opts.Default {
		return v, false
	}

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
		keyName := func(k int) string {
			kn := strings.ToLower(tcell.KeyNames[keys[k]])
			if v, ok := a.Get(keys[k]); ok && len(v.Opts.DisplayName) > 0 {
				kn = v.Opts.DisplayName
			}
			return kn
		}

		return keyName(i) < keyName(j)
	})

	// 按顺序遍历
	for _, k := range keys {
		f(k, a.actions[k])
	}
}

func (a *KeyActions) Merge(as *KeyActions) {
	a.mx.Lock()
	defer a.mx.Unlock()

	for k, v := range as.actions {
		a.actions[k] = v
	}
}
