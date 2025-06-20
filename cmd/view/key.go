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

func init() {
	initNumbKeys()
	initStdKeys()
	initShiftKeys()
	tcell.KeyNames[KeyHelp] = "?"
	tcell.KeyNames[KeySlash] = "/"
	tcell.KeyNames[KeySpace] = "space"
	tcell.KeyNames[KeyColon] = ":"
}

const (
	KeyHelp  tcell.Key = 63
	KeySpace tcell.Key = 32
	KeySlash tcell.Key = 47
	KeyColon tcell.Key = 58
)

const (
	Key0 tcell.Key = iota + 48
	Key1
	Key2
	Key3
	Key4
	Key5
	Key6
	Key7
	Key8
	Key9
)

const (
	KeyA tcell.Key = iota + 97
	KeyB
	KeyC
	KeyD
	KeyE
	KeyF
	KeyG
	KeyH
	KeyI
	KeyJ
	KeyK
	KeyL
	KeyM
	KeyN
	KeyO
	KeyP
	KeyQ
	KeyR
	KeyS
	KeyT
	KeyU
	KeyV
	KeyW
	KeyX
	KeyY
	KeyZ
)

const (
	KeyShiftA tcell.Key = iota + 65
	KeyShiftB
	KeyShiftC
	KeyShiftD
	KeyShiftE
	KeyShiftF
	KeyShiftG
	KeyShiftH
	KeyShiftI
	KeyShiftJ
	KeyShiftK
	KeyShiftL
	KeyShiftM
	KeyShiftN
	KeyShiftO
	KeyShiftP
	KeyShiftQ
	KeyShiftR
	KeyShiftS
	KeyShiftT
	KeyShiftU
	KeyShiftV
	KeyShiftW
	KeyShiftX
	KeyShiftY
	KeyShiftZ
)

func initNumbKeys() {
	tcell.KeyNames[Key0] = "0"
	tcell.KeyNames[Key1] = "1"
	tcell.KeyNames[Key2] = "2"
	tcell.KeyNames[Key3] = "3"
	tcell.KeyNames[Key4] = "4"
	tcell.KeyNames[Key5] = "5"
	tcell.KeyNames[Key6] = "6"
	tcell.KeyNames[Key7] = "7"
	tcell.KeyNames[Key8] = "8"
	tcell.KeyNames[Key9] = "9"
}

func initStdKeys() {
	tcell.KeyNames[KeyA] = "a"
	tcell.KeyNames[KeyB] = "b"
	tcell.KeyNames[KeyC] = "c"
	tcell.KeyNames[KeyD] = "d"
	tcell.KeyNames[KeyE] = "e"
	tcell.KeyNames[KeyF] = "f"
	tcell.KeyNames[KeyG] = "g"
	tcell.KeyNames[KeyH] = "h"
	tcell.KeyNames[KeyI] = "i"
	tcell.KeyNames[KeyJ] = "j"
	tcell.KeyNames[KeyK] = "k"
	tcell.KeyNames[KeyL] = "l"
	tcell.KeyNames[KeyM] = "m"
	tcell.KeyNames[KeyN] = "n"
	tcell.KeyNames[KeyO] = "o"
	tcell.KeyNames[KeyP] = "p"
	tcell.KeyNames[KeyQ] = "q"
	tcell.KeyNames[KeyR] = "r"
	tcell.KeyNames[KeyS] = "s"
	tcell.KeyNames[KeyT] = "t"
	tcell.KeyNames[KeyU] = "u"
	tcell.KeyNames[KeyV] = "v"
	tcell.KeyNames[KeyW] = "w"
	tcell.KeyNames[KeyX] = "x"
	tcell.KeyNames[KeyY] = "y"
	tcell.KeyNames[KeyZ] = "z"
}

func initShiftKeys() {
	tcell.KeyNames[KeyShiftA] = "Shift-A"
	tcell.KeyNames[KeyShiftB] = "Shift-B"
	tcell.KeyNames[KeyShiftC] = "Shift-C"
	tcell.KeyNames[KeyShiftD] = "Shift-D"
	tcell.KeyNames[KeyShiftE] = "Shift-E"
	tcell.KeyNames[KeyShiftF] = "Shift-F"
	tcell.KeyNames[KeyShiftG] = "Shift-G"
	tcell.KeyNames[KeyShiftH] = "Shift-H"
	tcell.KeyNames[KeyShiftI] = "Shift-I"
	tcell.KeyNames[KeyShiftJ] = "Shift-J"
	tcell.KeyNames[KeyShiftK] = "Shift-K"
	tcell.KeyNames[KeyShiftL] = "Shift-L"
	tcell.KeyNames[KeyShiftM] = "Shift-M"
	tcell.KeyNames[KeyShiftN] = "Shift-N"
	tcell.KeyNames[KeyShiftO] = "Shift-O"
	tcell.KeyNames[KeyShiftP] = "Shift-P"
	tcell.KeyNames[KeyShiftQ] = "Shift-Q"
	tcell.KeyNames[KeyShiftR] = "Shift-R"
	tcell.KeyNames[KeyShiftS] = "Shift-S"
	tcell.KeyNames[KeyShiftT] = "Shift-T"
	tcell.KeyNames[KeyShiftU] = "Shift-U"
	tcell.KeyNames[KeyShiftV] = "Shift-V"
	tcell.KeyNames[KeyShiftW] = "Shift-W"
	tcell.KeyNames[KeyShiftX] = "Shift-X"
	tcell.KeyNames[KeyShiftY] = "Shift-Y"
	tcell.KeyNames[KeyShiftZ] = "Shift-Z"
}
