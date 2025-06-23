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

import "github.com/rivo/tview"

const (
	// 注意: Logo最多6行，最少2行
	logo = `____________    _______  ___
__  ____/_ |  / /__   |/  /
_  / __ __ | / /__  /|_/ / 
/ /_/ / __ |/ / _  /  / /  
\____/  _____/  /_/  /_/`
)

const (
	pageLanguages        = "languages"
	pageLanguageVersions = "languageVersions"
	pageInstaller        = "installer"
)

const (
	alertKey   = "alert"
	infoKey = "info"
	confirmKey = "confirm"
	errorMsg   = `
  (\_/)    
  ( •_•)   
  / >......`
)

var Borders = struct {
	Horizontal  rune
	Vertical    rune
	TopLeft     rune
	TopRight    rune
	BottomLeft  rune
	BottomRight rune

	LeftT   rune
	RightT  rune
	TopT    rune
	BottomT rune
	Cross   rune

	HorizontalFocus  rune
	VerticalFocus    rune
	TopLeftFocus     rune
	TopRightFocus    rune
	BottomLeftFocus  rune
	BottomRightFocus rune
}{
	Horizontal:  tview.BoxDrawingsLightHorizontal,
	Vertical:    tview.BoxDrawingsLightVertical,
	TopLeft:     tview.BoxDrawingsLightDownAndRight,
	TopRight:    tview.BoxDrawingsLightDownAndLeft,
	BottomLeft:  tview.BoxDrawingsLightUpAndRight,
	BottomRight: tview.BoxDrawingsLightUpAndLeft,

	LeftT:   tview.BoxDrawingsLightVerticalAndRight,
	RightT:  tview.BoxDrawingsLightVerticalAndLeft,
	TopT:    tview.BoxDrawingsLightDownAndHorizontal,
	BottomT: tview.BoxDrawingsLightUpAndHorizontal,
	Cross:   tview.BoxDrawingsLightVerticalAndHorizontal,

	HorizontalFocus:  tview.BoxDrawingsLightHorizontal,
	VerticalFocus:    tview.BoxDrawingsLightVertical,
	TopLeftFocus:     tview.BoxDrawingsLightDownAndRight,
	TopRightFocus:    tview.BoxDrawingsLightDownAndLeft,
	BottomLeftFocus:  tview.BoxDrawingsLightUpAndRight,
	BottomRightFocus: tview.BoxDrawingsLightUpAndLeft,
}
