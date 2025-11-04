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

package main

import (
	"os"

	"github.com/toodofun/gvm/cmd"
	_ "github.com/toodofun/gvm/languages/github"
	_ "github.com/toodofun/gvm/languages/golang"
	_ "github.com/toodofun/gvm/languages/gvm"
	_ "github.com/toodofun/gvm/languages/java"
	_ "github.com/toodofun/gvm/languages/node"
	_ "github.com/toodofun/gvm/languages/python"
	_ "github.com/toodofun/gvm/languages/ruby"
	_ "github.com/toodofun/gvm/languages/rust"

	"github.com/sirupsen/logrus"
)

//func initI18n() {
//	bundle := i18n.NewBundle(language.English)
//	bundle.RegisterUnmarshalFunc("yaml", yaml.Unmarshal)
//
//	_, err := bundle.LoadMessageFile("i18n/active.zh.yaml")
//	if err != nil {
//		logrus.Fatalf("can not load i18n translate file: %v", err)
//	}
//}

func main() {
	//initI18n()

	root := cmd.NewRootCmd()
	if len(os.Args) == 1 {
		os.Args = append(os.Args, "ui")
	}
	if err := root.Execute(); err != nil {
		logrus.Fatalf("%v", err)
	}
}
