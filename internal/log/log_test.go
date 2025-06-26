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

package log_test

import (
	"bytes"
	"context"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/toodofun/gvm/internal/core"
	customlog "github.com/toodofun/gvm/internal/log"

	"github.com/sirupsen/logrus"
)

func TestSetAndGetLevel(t *testing.T) {
	customlog.SetLevel(logrus.DebugLevel)
	if customlog.GetLevel() != "debug" {
		t.Errorf("Expected debug level, got %s", customlog.GetLevel())
	}

	customlog.SetLevel(logrus.WarnLevel)
	if customlog.GetLevel() != "warning" {
		t.Errorf("Expected warn level, got %s", customlog.GetLevel())
	}
}

func TestGetWriter_Default(t *testing.T) {
	ctx := context.Background()
	writer := customlog.GetWriter(ctx)
	if writer != os.Stdout {
		t.Errorf("Expected os.Stdout, got %v", writer)
	}
}

func TestGetWriter_Custom(t *testing.T) {
	buf := &bytes.Buffer{}
	ctx := context.WithValue(context.Background(), core.ContextLogWriterKey, io.Writer(buf))
	writer := customlog.GetWriter(ctx)
	writer.Write([]byte("test"))

	if !strings.Contains(buf.String(), "test") {
		t.Errorf("Expected buffer to contain 'test', got %s", buf.String())
	}
}

func TestGetLogger_WriteToBuffer(t *testing.T) {
	buf := &bytes.Buffer{}
	ctx := context.WithValue(context.Background(), core.ContextLogWriterKey, io.Writer(buf))

	customlog.SetLevel(logrus.InfoLevel)
	logger := customlog.GetLogger(ctx)
	logger.Infof("hello log")

	if !strings.Contains(buf.String(), "hello log") {
		t.Errorf("Expected 'hello log' in buffer, got: %s", buf.String())
	}
}
