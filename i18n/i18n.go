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

package i18n

import (
	"context"
	"embed"
	"io/fs"
	"strings"
	"sync"

	"github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
	"gopkg.in/yaml.v3"

	"github.com/duke-git/lancet/v2/slice"

	"github.com/toodofun/gvm/internal/core"
	"github.com/toodofun/gvm/internal/log"
	"github.com/toodofun/gvm/internal/util/file"
)

//go:embed *.yaml
var i18nFS embed.FS

var bundle *i18n.Bundle
var localizer *i18n.Localizer
var once sync.Once
var defaultLocalizer *i18n.Localizer

func InitI18n(ctx context.Context) {
	logger := log.GetLogger(ctx)
	once.Do(func() {
		bundle = i18n.NewBundle(language.English)
		logger.Debugf("bundle default language: %s", bundle.LanguageTags()[0])
		bundle.RegisterUnmarshalFunc("yaml", yaml.Unmarshal)

		if err := fs.WalkDir(i18nFS, ".", func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}

			if d.IsDir() {
				return nil
			}

			filename := strings.ToLower(d.Name())
			if strings.HasSuffix(filename, ".yaml") || strings.HasSuffix(filename, ".yml") {
				logger.Debugf("find i18n translate file: %s", path)
				data, err := i18nFS.ReadFile(path)
				if err != nil {
					logger.Errorf("failed to read i18n translate file: %s, %v", path, err)
					return nil
				}
				_, err = bundle.ParseMessageFileBytes(data, path)
				if err != nil {
					logger.Errorf("can not load i18n translate file: %v", err)
				}
			}
			return nil
		}); err != nil {
			logger.Errorf("can not find i18n config file: %v", err)
		}

		logger.Debugf("tags: %+v", bundle.LanguageTags())
		defaultLanguage := core.GetConfig().Language
		if len(defaultLanguage) == 0 {
			defaultLanguage = "en"
		}
		defaultLanguages := []string{defaultLanguage, "en"}
		slice.Unique(defaultLanguages)
		localizer = i18n.NewLocalizer(bundle, defaultLanguages...)
		defaultLocalizer = i18n.NewLocalizer(bundle, "en")
	})
}

func GetTranslate(id string, templateData map[string]any) string {
	InitI18n(context.Background())
	res, err := localizer.Localize(&i18n.LocalizeConfig{
		MessageID:    id,
		TemplateData: templateData,
	})
	if err != nil {
		if res, err = defaultLocalizer.Localize(&i18n.LocalizeConfig{
			MessageID:    id,
			TemplateData: templateData,
		}); err != nil {
			return id
		}
	}
	return res
}

func SetLanguage(language string) error {
	InitI18n(context.Background())
	localizer = i18n.NewLocalizer(bundle, language)
	config := core.GetConfig()
	config.Language = language
	return file.WriteJSONFile(core.GetConfigPath(), config)
}
