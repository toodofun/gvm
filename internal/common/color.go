package common

import "github.com/fatih/color"

var (
	fgRed   = color.New(color.FgRed).SprintFunc()
	fgGreen = color.New(color.FgGreen).SprintFunc()
	fgBlue  = color.New(color.FgBlue).SprintFunc()
)

func RedFont(s string) string {
	return fgRed(s)
}

func GreenFont(s string) string {
	return fgGreen(s)
}

func BlueFont(s string) string {
	return fgBlue(s)
}
