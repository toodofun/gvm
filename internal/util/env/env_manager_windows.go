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

//go:build windows

package env

import (
	"fmt"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/registry"
)

const pathSeparator = ";"

func (m *Manager) GetEnv(key string) (string, error) {
	envKey, err := registry.OpenKey(registry.CURRENT_USER, `Environment`, registry.QUERY_VALUE)
	if err != nil {
		return "", err
	}
	defer envKey.Close()

	val, _, err := envKey.GetStringValue(key)
	if err != nil {
		return "", nil
	} else {
		return val, nil
	}
}

func (m *Manager) SetEnv(key, value string) error {
	value = m.quoteValue(value)
	envKey, err := registry.OpenKey(registry.CURRENT_USER, `Environment`, registry.SET_VALUE)
	if err != nil {
		return err
	}
	defer envKey.Close()

	if err = envKey.SetStringValue(key, value); err != nil {
		return err
	}

	// 通知系统环境变量已更改
	_ = m.notifyEnvChange()
	return nil
}

func (m *Manager) DeleteEnv(key string) error {
	envKey, err := registry.OpenKey(registry.CURRENT_USER, `Environment`, registry.SET_VALUE)
	if err != nil {
		return err
	}
	defer envKey.Close()

	if err = envKey.DeleteValue(key); err != nil {
		return err
	}

	// 通知系统环境变量已更改
	_ = m.notifyEnvChange()
	return nil
}

// notifyEnvChange 通知系统环境变量已更改
func (m *Manager) notifyEnvChange() error {
	ptr, err := windows.UTF16PtrFromString("Environment")
	if err != nil {
		return err
	}
	ret, _, _ := syscall.NewLazyDLL("user32.dll").
		NewProc("SendMessageTimeoutW").
		Call(
			0xFFFF, // HWND_BROADCAST
			0x001A, // WM_SETTINGCHANGE
			0,
			uintptr(unsafe.Pointer(ptr)),
			0,
			1000,
			0,
		)
	if ret == 0 {
		// 发送失败，可能是因为没有足够的权限或其他原因
		return fmt.Errorf("failed to notify environment change")
	}
	return nil
}
