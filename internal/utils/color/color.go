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

package color

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
