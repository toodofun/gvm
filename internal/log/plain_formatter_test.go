package log

import (
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestPlainFormatter_Format(t *testing.T) {
	formatter := &PlainFormatter{}
	entry := &logrus.Entry{
		Message: "test message",
	}

	formatted, err := formatter.Format(entry)

	assert.NoError(t, err)
	assert.Equal(t, "test message\n", string(formatted))
}
