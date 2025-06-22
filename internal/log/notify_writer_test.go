package log

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNotifyBuffer_WriteAndRead(t *testing.T) {
	buf := NewNotifyBuffer()

	n, err := buf.Write([]byte("Hello World"))
	assert.NoError(t, err)
	assert.Equal(t, 11, n)
	assert.Equal(t, "Hello World", buf.Read())
}

func TestNotifyBuffer_WriteWithNewlines(t *testing.T) {
	buf := NewNotifyBuffer()

	buf.Write([]byte("line1\nline2\nline3"))
	assert.Equal(t, "line3", buf.Read())
}

func TestNotifyBuffer_WriteWithCarriageReturn(t *testing.T) {
	buf := NewNotifyBuffer()

	buf.Write([]byte("line1\rline2\rline3"))
	assert.Equal(t, "line3", buf.Read())
}

func TestNotifyBuffer_EmptyAndWhitespaceInput(t *testing.T) {
	buf := NewNotifyBuffer()

	n, err := buf.Write([]byte("   \n  "))
	assert.NoError(t, err)
	assert.Equal(t, 6, n)
	assert.Equal(t, "", buf.Read())
}

func TestNotifyBuffer_Notification(t *testing.T) {
	buf := NewNotifyBuffer()

	// drain the channel in case anything in there
	select {
	case <-buf.Updated:
	default:
	}

	buf.Write([]byte("trigger"))
	select {
	case <-buf.Updated:
		// OK
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Expected notification on Updated channel")
	}
}

func TestNotifyBuffer_NotificationNonBlocking(t *testing.T) {
	buf := NewNotifyBuffer()

	// fill the channel
	buf.Updated <- struct{}{}

	// this should not block or panic
	_, err := buf.Write([]byte("second write"))
	assert.NoError(t, err)
}

func TestNotifyBuffer_Close(t *testing.T) {
	buf := NewNotifyBuffer()

	go func() {
		// Should be able to detect closed channel
		<-buf.Updated
	}()

	buf.Close()

	// Multiple calls to Close should not panic
	buf.Close()

	// After close, write should be a no-op
	n, err := buf.Write([]byte("after close"))
	assert.NoError(t, err)
	assert.Equal(t, 0, n)
}

func TestNotifyBuffer_WriteEmptyInput(t *testing.T) {
	buf := NewNotifyBuffer()

	n, err := buf.Write([]byte{})
	assert.NoError(t, err)
	assert.Equal(t, 0, n)
}

func TestNotifyBuffer_WriteAfterClose(t *testing.T) {
	buf := NewNotifyBuffer()
	buf.Close()

	n, err := buf.Write([]byte("should be ignored"))
	assert.NoError(t, err)
	assert.Equal(t, 0, n)
}

