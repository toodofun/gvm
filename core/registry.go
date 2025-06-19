package core

import "sort"

var languages = make(map[string]Language)

func RegisterLanguage(lang Language) {
	languages[lang.Name()] = lang
}

func GetLanguage(name string) (Language, bool) {
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
