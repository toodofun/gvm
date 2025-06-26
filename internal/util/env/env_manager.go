package env

import (
	"fmt"
	"github.com/duke-git/lancet/v2/slice"
	"runtime"
	"strings"
)

type IManager interface {
	// AppendEnv 向环境变量中追加一个值
	AppendEnv(key, value string) error
	// RemoveEnv 从环境变量中移除一个值
	RemoveEnv(key, value string) error
	// GetEnv 获取环境变量的值
	GetEnv(key string) (string, error)
	// SetEnv 设置环境变量的值
	SetEnv(key, value string) error
	// DeleteEnv 删除环境变量
	DeleteEnv(key string) error
}
type Manager struct {
}

func NewEnvManager() *Manager {
	return &Manager{}
}

func (m *Manager) AppendEnv(key, value string) error {
	if len(value) == 0 {
		return fmt.Errorf("value is empty")
	}

	value = m.quoteValue(value)

	values, err := m.GetEnv(key)
	if err != nil {
		return err
	}
	var valueList []string

	if len(values) > 0 {
		valueList = strings.Split(value, pathSeparator)
		valueList = append(valueList, strings.Split(values, pathSeparator)...)
	} else {
		valueList = []string{value}
	}

	if runtime.GOOS == "windows" {
		valueList = append(valueList, "%"+key+"%")
	} else {
		valueList = append(valueList, "$"+key)
	}
	// 去重
	valueList = slice.Unique(valueList)
	slice.Sort(valueList)
	return m.SetEnv(key, strings.Join(valueList, pathSeparator))
}

func (m *Manager) RemoveEnv(key, value string) error {
	if len(value) == 0 {
		return fmt.Errorf("value is empty")
	}

	value = m.quoteValue(value)

	values, err := m.GetEnv(key)
	if err != nil {
		return err
	}
	if len(values) == 0 {
		return nil
	}

	valueList := strings.Split(values, pathSeparator)
	newValueList := slice.Filter(valueList, func(index int, item string) bool {
		return item != value
	})

	if len(newValueList) == 0 {
		return m.DeleteEnv(key)
	}
	if runtime.GOOS == "windows" {
		newValueList = append(newValueList, "%"+key+"%")
	} else {
		newValueList = append(newValueList, "$"+key)
	}
	// 去重
	newValueList = slice.Unique(newValueList)
	slice.Sort(newValueList)
	return m.SetEnv(key, strings.Join(newValueList, pathSeparator))
}

func (m *Manager) quoteValue(value string) string {
	// 如果值包含空格或特殊字符，需要加引号
	if strings.ContainsAny(value, " \t\n\r\"'\\$`") {
		return fmt.Sprintf("\"%s\"", strings.ReplaceAll(value, "\"", "\\\""))
	}
	return value
}
