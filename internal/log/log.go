package log

import (
	"github.com/sirupsen/logrus"
	"gvm/core"
	"io"
	"os"
	"path"
	"strings"
	"sync"
)

var (
	Logger *logrus.Logger
	Writer io.Writer = nil
)

func SwitchTo(writer io.Writer) {
	Logger.SetOutput(writer)
}

func SetWriter(writer io.Writer) {
	Logger.SetOutput(createWriter(writer))
}

func createWriter(writer io.Writer) io.Writer {
	file, err := os.OpenFile(path.Join(core.GetRootDir(), "app.log"), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		Logger.Fatal("无法打开日志文件:", err)
	}

	Writer = io.MultiWriter(file, writer)
	return Writer
}

type NotifyBuffer struct {
	buf     string
	mu      sync.Mutex
	Updated chan struct{}
}

func NewNotifyBuffer() *NotifyBuffer {
	return &NotifyBuffer{
		Updated: make(chan struct{}, 1),
	}
}

func (b *NotifyBuffer) Write(p []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	// 处理进度条的 \r 或 \n：保留最后一行
	lines := strings.Split(string(p), "\r")
	b.buf = lines[len(lines)-1]

	select {
	case b.Updated <- struct{}{}:
	default:
	}

	return len(p), nil
}

func (b *NotifyBuffer) ReadAndReset() string {
	b.mu.Lock()
	defer b.mu.Unlock()
	s := b.buf
	return s
}

func init() {
	Logger = logrus.New()
}
