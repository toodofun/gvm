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

package core

import "sort"

var languages = make(map[string]Language)

func RegisterLanguage(lang Language) {
	languages[lang.Name()] = lang
}

var GetLanguage = func(name string) (Language, bool) {
	lang, exists := languages[name]
	return lang, exists
}

func GetAllLanguage() []string {
	res := make([]string, 0)

	for k := range languages {
		res = append(res, k)
	}
	sort.Strings(res)
	return res
}
