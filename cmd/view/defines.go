package view

import "github.com/rivo/tview"

const (
	// 注意: Logo最多6行，最少2行
	logo = `
___________    _______  ___
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
