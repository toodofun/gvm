package file

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// 辅助函数：检查字符串是否包含子字符串
func contains(s, substr string) bool {
	return strings.Contains(s, substr)
}

// 测试用的结构体
type TestConfig struct {
	Name    string   `json:"name"`
	Age     int      `json:"age"`
	Enabled bool     `json:"enabled"`
	Items   []string `json:"items"`
}

// 测试WriteJSONFile函数
func TestWriteJSONFile(t *testing.T) {
	t.Run("successful write", func(t *testing.T) {
		// 创建临时文件
		tempFile := filepath.Join(t.TempDir(), "test.json")

		// 测试数据
		testData := TestConfig{
			Name:    "测试用户",
			Age:     25,
			Enabled: true,
			Items:   []string{"item1", "item2", "item3"},
		}

		// 写入JSON文件
		err := WriteJSONFile(tempFile, testData)
		if err != nil {
			t.Fatalf("WriteJSONFile failed: %v", err)
		}

		// 验证文件是否存在
		if _, err := os.Stat(tempFile); os.IsNotExist(err) {
			t.Fatal("文件写入后不存在")
		}

		// 读取文件内容验证
		content, err := os.ReadFile(tempFile)
		if err != nil {
			t.Fatalf("读取文件失败: %v", err)
		}

		// 验证JSON格式是否正确
		var result TestConfig
		if err := json.Unmarshal(content, &result); err != nil {
			t.Fatalf("文件内容不是有效的JSON: %v", err)
		}

		// 验证数据是否正确
		if result.Name != testData.Name || result.Age != testData.Age || result.Enabled != testData.Enabled {
			t.Errorf("写入的数据不正确: got %+v, want %+v", result, testData)
		}
	})

	t.Run("marshal error", func(t *testing.T) {
		tempFile := filepath.Join(t.TempDir(), "test.json")

		// 使用无法序列化的数据（如函数）
		invalidData := map[string]interface{}{
			"func": func() {},
		}

		err := WriteJSONFile(tempFile, invalidData)
		if err == nil {
			t.Error("期望写入无效数据时返回错误")
		}
	})

	t.Run("invalid file path", func(t *testing.T) {
		// 使用无效的文件路径
		invalidPath := "/root/nonexistent/directory/test.json"
		testData := TestConfig{Name: "test"}

		err := WriteJSONFile(invalidPath, testData)
		if err == nil {
			t.Error("期望写入无效路径时返回错误")
		}
	})
}

// 测试ReadJSONFile函数
func TestReadJSONFile(t *testing.T) {
	t.Run("successful read", func(t *testing.T) {
		// 创建临时文件
		tempFile := filepath.Join(t.TempDir(), "test.json")

		// 测试数据
		testData := TestConfig{
			Name:    "测试用户",
			Age:     30,
			Enabled: false,
			Items:   []string{"a", "b", "c"},
		}

		// 先写入测试数据
		err := WriteJSONFile(tempFile, testData)
		if err != nil {
			t.Fatalf("准备测试数据失败: %v", err)
		}

		// 读取JSON文件
		var result TestConfig
		err = ReadJSONFile(tempFile, &result)
		if err != nil {
			t.Fatalf("ReadJSONFile failed: %v", err)
		}

		// 验证数据是否正确
		if result.Name != testData.Name || result.Age != testData.Age || result.Enabled != testData.Enabled {
			t.Errorf("读取的数据不正确: got %+v, want %+v", result, testData)
		}

		// 验证数组数据
		if len(result.Items) != len(testData.Items) {
			t.Errorf("Items数组长度不正确: got %d, want %d", len(result.Items), len(testData.Items))
		}

		for i, item := range result.Items {
			if item != testData.Items[i] {
				t.Errorf("Items[%d] 不正确: got %s, want %s", i, item, testData.Items[i])
			}
		}
	})

	t.Run("file does not exist", func(t *testing.T) {
		nonexistentFile := filepath.Join(t.TempDir(), "nonexistent.json")

		var result TestConfig
		err := ReadJSONFile(nonexistentFile, &result)
		if err == nil {
			t.Error("期望读取不存在的文件时返回错误")
		}

		// 检查错误消息是否包含"does not exist"
		if !contains(err.Error(), "does not exist") {
			t.Errorf("期望返回文件不存在错误，但得到: %v", err)
		}
	})

	t.Run("invalid JSON content", func(t *testing.T) {
		// 创建包含无效JSON的文件
		tempFile := filepath.Join(t.TempDir(), "invalid.json")
		invalidJSON := `{"name": "test", "age": 25, "enabled": tru` // 不完整的JSON

		err := os.WriteFile(tempFile, []byte(invalidJSON), 0644)
		if err != nil {
			t.Fatalf("创建无效JSON文件失败: %v", err)
		}

		var result TestConfig
		err = ReadJSONFile(tempFile, &result)
		if err == nil {
			t.Error("期望解析无效JSON时返回错误")
		}
	})

	t.Run("empty file", func(t *testing.T) {
		// 创建空文件
		tempFile := filepath.Join(t.TempDir(), "empty.json")
		err := os.WriteFile(tempFile, []byte(""), 0644)
		if err != nil {
			t.Fatalf("创建空文件失败: %v", err)
		}

		var result TestConfig
		err = ReadJSONFile(tempFile, &result)
		if err == nil {
			t.Error("期望读取空文件时返回错误")
		}
	})
}

// 测试读写操作的集成测试
func TestReadWriteIntegration(t *testing.T) {
	tempFile := filepath.Join(t.TempDir(), "integration.json")

	// 原始数据
	originalData := TestConfig{
		Name:    "集成测试",
		Age:     35,
		Enabled: true,
		Items:   []string{"test1", "test2", "test3"},
	}

	// 写入数据
	err := WriteJSONFile(tempFile, originalData)
	if err != nil {
		t.Fatalf("写入失败: %v", err)
	}

	// 读取数据
	var readData TestConfig
	err = ReadJSONFile(tempFile, &readData)
	if err != nil {
		t.Fatalf("读取失败: %v", err)
	}

	// 验证数据一致性
	if readData.Name != originalData.Name ||
		readData.Age != originalData.Age ||
		readData.Enabled != originalData.Enabled {
		t.Errorf("数据不一致: got %+v, want %+v", readData, originalData)
	}

	// 验证数组数据
	if len(readData.Items) != len(originalData.Items) {
		t.Errorf("数组长度不一致: got %d, want %d", len(readData.Items), len(originalData.Items))
	}

	for i, item := range readData.Items {
		if item != originalData.Items[i] {
			t.Errorf("数组元素不一致: got %s, want %s", item, originalData.Items[i])
		}
	}
}

// 基准测试
func BenchmarkWriteJSONFile(b *testing.B) {
	tempDir := b.TempDir()
	testData := TestConfig{
		Name:    "基准测试",
		Age:     25,
		Enabled: true,
		Items:   []string{"item1", "item2", "item3"},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		filename := filepath.Join(tempDir, fmt.Sprintf("bench_%d.json", i))
		err := WriteJSONFile(filename, testData)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkReadJSONFile(b *testing.B) {
	// 准备测试文件
	tempFile := filepath.Join(b.TempDir(), "bench_read.json")
	testData := TestConfig{
		Name:    "基准测试",
		Age:     25,
		Enabled: true,
		Items:   []string{"item1", "item2", "item3"},
	}

	err := WriteJSONFile(tempFile, testData)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var result TestConfig
		err := ReadJSONFile(tempFile, &result)
		if err != nil {
			b.Fatal(err)
		}
	}
}
