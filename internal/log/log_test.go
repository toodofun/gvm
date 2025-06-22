package log_test

import (
	"bytes"
	"context"
	"github.com/sirupsen/logrus"
	"gvm/internal/core"
	customlog "gvm/internal/log"
	"io"
	"os"
	"strings"
	"testing"
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
