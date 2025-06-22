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

package log

import (
	"context"
	"gvm/internal/core"
	"io"
	"os"
	"path"
	"sync"

	"github.com/sirupsen/logrus"
)

var (
	logFile     *os.File
	logFileOnce sync.Once
	logLevel    = logrus.InfoLevel
)

type ILogger interface {
	Tracef(format string, args ...interface{})
	Debugf(format string, args ...interface{})
	Infof(format string, args ...interface{})
	Warnf(format string, args ...interface{})
	Errorf(format string, args ...interface{})
	Fatalf(format string, args ...interface{})
	Panicf(format string, args ...interface{})
	Trace(args ...interface{})
	Debug(args ...interface{})
	Info(args ...interface{})
	Warn(args ...interface{})
	Error(args ...interface{})
	Fatal(args ...interface{})
	Panic(args ...interface{})
}

func GetWriter(ctx context.Context) (writer io.Writer) {
	writer = os.Stdout
	logWriterValue := ctx.Value(core.ContextLogWriterKey)
	if logWriterValue != nil {
		if logWriter, ok := logWriterValue.(io.Writer); ok {
			writer = logWriter
		}
	}
	return
}

func SetLevel(level logrus.Level) {
	logLevel = level
}

func GetLevel() string {
	return logLevel.String()
}

func GetLogger(ctx context.Context) ILogger {
	logger := logrus.New()
	logger.SetLevel(logLevel)
	logger.SetOutput(createWriter(GetWriter(ctx)))
	logger.SetFormatter(&PlainFormatter{})
	return logger
}

func createWriter(writer io.Writer) io.Writer {
	logFileOnce.Do(func() {
		var err error
		logFile, err = os.OpenFile(path.Join(core.GetRootDir(), "app.log"),
			os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			logFile = nil
		}
	})

	if logFile == nil {
		return writer
	}

	return io.MultiWriter(logFile, writer)
}
