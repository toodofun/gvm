# TDD 工作流程指南

## 什么是 TDD

TDD (Test-Driven Development, 测试驱动开发) 是一种开发方法论，强调在编写功能代码之前先编写测试代码。

### TDD 的核心价值

1. **设计驱动**: 通过测试思考 API 设计和接口定义
2. **文档作用**: 测试代码即功能文档，展示如何使用代码
3. **重构保障**: 测试用例确保重构时不会破坏现有功能
4. **快速反馈**: 立即发现错误，减少调试时间
5. **质量保证**: 高覆盖率确保代码质量

## TDD 开发循环

TDD 遵循 **Red-Green-Refactor** 循环:

```
┌─────────────────────────────────────────┐
│                                         │
│  1. RED   (编写失败的测试)               │
│     ↓                                  │
│  2. GREEN (编写最少代码通过测试)         │
│     ↓                                  │
│  3. REFACTOR (重构代码，保持测试通过)     │
│     ↓                                  │
│  回到步骤 1                              │
│                                         │
└─────────────────────────────────────────┘
```

### 详细步骤

#### 1. RED - 编写失败的测试

**目标**: 编写一个测试用例，验证你想要实现的功能

**示例场景**: 实现版本解析功能

```go
// pkg/version/parser_test.go
package version

import "testing"

func TestParseVersion(t *testing.T) {
    // 编写测试，验证我们想要的功能
    input := "go1.20.0"
    want := "1.20.0"

    got, err := ParseVersion(input)

    if err != nil {
        t.Errorf("ParseVersion() error = %v", err)
        return
    }

    if got != want {
        t.Errorf("ParseVersion() = %v, want %v", got, want)
    }
}
```

**运行测试**: 必定失败 (因为函数还不存在)

```bash
$ go test ./pkg/version
# command-line-arguments
pkg/version/parser_test.go:10:9: undefined: ParseVersion
FAIL  pkg/version [build failed]
```

#### 2. GREEN - 编写最少代码通过测试

**目标**: 编写最简单的代码，使测试通过

**实现**:

```go
// pkg/version/parser.go
package version

import "fmt"

func ParseVersion(input string) (string, error) {
    // 最简单的实现：硬编码返回预期值
    if input == "go1.20.0" {
        return "1.20.0", nil
    }
    return "", fmt.Errorf("invalid version: %s", input)
}
```

**运行测试**: 通过

```bash
$ go test ./pkg/version
PASS
ok  pkg/version 0.002s
```

#### 3. REFACTOR - 重构代码

**目标**: 改进代码质量，同时保持测试通过

**重构实现**:

```go
// pkg/version/parser.go
package version

import (
    "errors"
    "regexp"
    "strings"
)

var versionRegex = regexp.MustCompile(`^go(\d+\.\d+\.\d+)$`)

func ParseVersion(input string) (string, error) {
    if input == "" {
        return "", errors.New("empty version string")
    }

    matches := versionRegex.FindStringSubmatch(input)
    if matches == nil {
        return "", fmt.Errorf("invalid version format: %s", input)
    }

    return matches[1], nil
}
```

**添加更多测试用例**:

```go
func TestParseVersion(t *testing.T) {
    tests := []struct {
        name    string
        input   string
        want    string
        wantErr bool
    }{
        {"valid version", "go1.20.0", "1.20.0", false},
        {"another valid version", "go1.21.5", "1.21.5", false},
        {"missing prefix", "1.20.0", "", true},
        {"empty string", "", "", true},
        {"invalid format", "go1.20", "", true},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := ParseVersion(tt.input)
            if (err != nil) != tt.wantErr {
                t.Errorf("ParseVersion() error = %v, wantErr %v", err, tt.wantErr)
                return
            }
            if got != tt.want {
                t.Errorf("ParseVersion() = %v, want %v", got, tt.want)
            }
        })
    }
}
```

**运行测试**: 确保所有测试通过

```bash
$ go test -v ./pkg/version
=== RUN   TestParseVersion
=== RUN   TestParseVersion/valid_version
=== RUN   TestParseVersion/another_valid_version
=== RUN   TestParseVersion/missing_prefix
=== RUN   TestParseVersion/empty_string
=== RUN   TestParseVersion/invalid_format
--- PASS: TestParseVersion (0.00s)
    --- PASS: TestParseVersion/valid_version (0.00s)
    --- PASS: TestParseVersion/another_valid_version (0.00s)
    --- PASS: TestParseVersion/missing_prefix (0.00s)
    --- PASS: TestParseVersion/empty_string (0.00s)
    --- PASS: TestParseVersion/invalid_format (0.00s)
PASS
ok  pkg/version 0.002s
```

## GVM 项目 TDD 规则

### 1. 测试文件位置

```
pkg/
├── version/
│   ├── parser.go          # 功能代码
│   ├── parser_test.go     # 测试代码
│   └── validator.go       # 功能代码
├── gvm/
│   ├── downloader.go
│   ├── downloader_test.go
│   └── installer.go
```

**规则**: 测试文件与功能代码在同一包内，命名为 `*_test.go`

### 2. 测试导入规则

```go
// 如果测试代码在同一个包内
package version

// 如果测试代码在独立的测试包内
package version_test

import (
    "testing"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"

    "your-project/pkg/version"  // 导入被测试的包
)
```

**推荐**: 默认使用 `package version`，只在需要测试未导出函数时使用

### 3. 测试命名约定

```go
// 函数测试
func TestFunctionName(t *testing.T)
func TestFunctionName_EdgeCase(t *testing.T)

// 方法测试
func TestStruct_Method(t *testing.T)
func TestStruct_Method_ErrorCase(t *testing.T)

// 基准测试
func BenchmarkFunction(b *testing.B)

// 表驱动测试子测试名称
func TestFunction(t *testing.T) {
    tests := []struct {
        name string  // 必须: 描述测试场景
        // ... 其他字段
    }{
        {"valid input", ...},
        {"invalid input", ...},
    }
}
```

### 4. 断言使用规则

```go
// 期望错误时使用 assert
func TestFunction_WithError(t *testing.T) {
    _, err := Function()
    assert.Error(t, err)
    assert.Contains(t, err.Error(), "expected error")
}

// 关键检查使用 require
func TestInstall(t *testing.T) {
    path, err := Download()
    require.NoError(t, err)  // 失败则立即停止

    err = Install(path)
    require.NoError(t, err)

    // 继续验证
    assert.FileExists(t, getBinaryPath())
}
```

**规则**:
- `assert`: 验证预期，失败后继续执行
- `require`: 关键检查，失败后立即停止

### 5. 外部依赖处理

**文件系统依赖**:

```go
func TestWriteConfig(t *testing.T) {
    // 使用临时目录
    tmpDir := testutil.CreateTempDir(t)
    defer os.RemoveAll(tmpDir)

    configPath := filepath.Join(tmpDir, "config.yaml")
    err := WriteConfig(configPath, config)
    assert.NoError(t, err)
}
```

**网络依赖**:

```go
func TestDownload(t *testing.T) {
    // 使用 mock HTTP 服务器
    server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusOK)
        w.Write([]byte("test data"))
    }))
    defer server.Close()

    downloader := NewDownloader()
    _, err := downloader.Download(server.URL, tmpDir)
    assert.NoError(t, err)
}
```

**环境变量依赖**:

```go
func TestWithEnvVar(t *testing.T) {
    // 使用 t.Setenv (Go 1.17+)
    t.Setenv("GVM_ROOT", "/tmp/gvm")

    // 测试代码
}

// 或者使用测试辅助函数
func setupTestEnv(t *testing.T) func() {
    oldHome := os.Getenv("HOME")
    os.Setenv("HOME", "/tmp/test-home")

    return func() {
        os.Setenv("HOME", oldHome)
    }
}
```

### 6. 并发测试规则

```go
func TestConcurrentDownloads(t *testing.T) {
    // 标记为可并行
    t.Parallel()

    // 测试代码
}
```

**规则**:
- 如果测试可以并行运行，使用 `t.Parallel()`
- 使用 `go test -parallel N` 控制并发数

## 常见错误

### 1. 测试不够具体

```go
// 不好: 测试太泛
func TestDownload(t *testing.T) {
    // 测试代码
}

// 好: 测试描述具体场景
func TestDownload_WithValidURL_ReturnsSuccess(t *testing.T) {
    // 测试代码
}
```

### 2. 测试耦合实现细节

```go
// 不好: 测试依赖于内部实现
func TestDownload_UsesThreeRetries(t *testing.T) {
    // 测试代码假设函数内部重试3次
}

// 好: 测试关注行为而非实现
func TestDownload_WithNetworkError_RetriesAndSucceeds(t *testing.T) {
    // 测试验证在网络错误后能重试成功
}
```

### 3. 测试包含过多逻辑

```go
// 不好: 测试逻辑复杂
func TestComplexScenario(t *testing.T) {
    // 50 行设置代码
    // 30 行测试逻辑
    // 20 行验证
}

// 好: 分解为多个小测试
func TestScenario_Step1(t *testing.T)
func TestScenario_Step2(t *testing.T)
func TestScenario_Step3(t *testing.T)
```

### 4. 忽略错误处理

```go
// 不好: 忽略错误
func TestWrite(t *testing.T) {
    WriteFile("test.txt", "data")
    // 假设成功
}

// 好: 检查错误
func TestWrite(t *testing.T) {
    err := WriteFile("test.txt", "data")
    if err != nil {
        t.Fatalf("WriteFile() failed: %v", err)
    }
}
```

### 5. 过度使用 Sleep

```go
// 不好: 使用 sleep 等待
func TestAsyncOperation(t *testing.T) {
    go AsyncOperation()
    time.Sleep(1 * time.Second)  // 不可靠且慢
    // 验证结果
}

// 好: 使用同步机制
func TestAsyncOperation(t *testing.T) {
    done := make(chan bool)
    go func() {
        AsyncOperation()
        done <- true
    }()
    <-done  // 等待完成
    // 验证结果
}
```

## 完整示例: 使用 TDD 开发版本查询功能

### 需求

实现一个功能，查询本地已安装的 Go 版本列表。

### 第 1 步: RED - 编写第一个测试

```go
// pkg/version/installed_test.go
package version

import (
    "os"
    "path/filepath"
    "testing"
)

func TestGetInstalledVersions(t *testing.T) {
    // 设置测试环境
    tmpDir := t.TempDir()
    versionsDir := filepath.Join(tmpDir, "versions")

    // 创建测试数据: 模拟已安装的版本
    os.MkdirAll(filepath.Join(versionsDir, "go1.20.0"), 0755)
    os.MkdirAll(filepath.Join(versionsDir, "go1.21.0"), 0755)

    // 执行测试
    versions, err := GetInstalledVersions(versionsDir)

    // 验证结果
    if err != nil {
        t.Fatalf("GetInstalledVersions() error = %v", err)
    }

    if len(versions) != 2 {
        t.Errorf("GetInstalledVersions() returned %d versions, want 2", len(versions))
    }

    expectedVersions := []string{"go1.20.0", "go1.21.0"}
    for i, v := range versions {
        if v != expectedVersions[i] {
            t.Errorf("versions[%d] = %v, want %v", i, v, expectedVersions[i])
        }
    }
}
```

**运行测试**: 失败 (函数不存在)

```bash
$ go test ./pkg/version
# command-line-arguments
pkg/version/installed_test.go:14:9: undefined: GetInstalledVersions
FAIL  pkg/version [build failed]
```

### 第 2 步: GREEN - 编写最少代码通过测试

```go
// pkg/version/installed.go
package version

import "fmt"

func GetInstalledVersions(versionsDir string) ([]string, error) {
    // 最简单的实现: 返回硬编码结果
    return []string{"go1.20.0", "go1.21.0"}, nil
}
```

**运行测试**: 通过

```bash
$ go test -v ./pkg/version
=== RUN   TestGetInstalledVersions
--- PASS: TestGetInstalledVersions (0.00s)
PASS
ok  pkg/version 0.002s
```

### 第 3 步: 添加更多测试用例

```go
func TestGetInstalledVersions_EmptyDirectory(t *testing.T) {
    tmpDir := t.TempDir()
    versionsDir := filepath.Join(tmpDir, "versions")

    versions, err := GetInstalledVersions(versionsDir)

    if err != nil {
        t.Fatalf("GetInstalledVersions() error = %v", err)
    }

    if len(versions) != 0 {
        t.Errorf("GetInstalledVersions() returned %d versions, want 0", len(versions))
    }
}

func TestGetInstalledVersions_NonExistentDirectory(t *testing.T) {
    tmpDir := t.TempDir()
    versionsDir := filepath.Join(tmpDir, "nonexistent")

    _, err := GetInstalledVersions(versionsDir)

    if err == nil {
        t.Error("GetInstalledVersions() expected error for non-existent directory")
    }
}
```

**运行测试**: 部分失败 (硬编码实现不适用)

```bash
$ go test -v ./pkg/version
=== RUN   TestGetInstalledVersions
--- PASS: TestGetInstalledVersions (0.00s)
=== RUN   TestGetInstalledVersions_EmptyDirectory
--- FAIL: TestGetInstalledVersions_EmptyDirectory (0.00s)
=== RUN   TestGetInstalledVersions_NonExistentDirectory
--- FAIL: TestGetInstalledVersions_NonExistentDirectory (0.00s)
```

### 第 4 步: REFACTOR - 实现真实逻辑

```go
// pkg/version/installed.go
package version

import (
    "errors"
    "os"
    "path/filepath"
    "sort"
    "strings"
)

func GetInstalledVersions(versionsDir string) ([]string, error) {
    // 检查目录是否存在
    info, err := os.Stat(versionsDir)
    if err != nil {
        if os.IsNotExist(err) {
            return nil, errors.New("versions directory does not exist")
        }
        return nil, err
    }

    if !info.IsDir() {
        return nil, errors.New("versions path is not a directory")
    }

    // 读取目录内容
    entries, err := os.ReadDir(versionsDir)
    if err != nil {
        return nil, err
    }

    // 过滤出 Go 版本目录
    var versions []string
    for _, entry := range entries {
        if entry.IsDir() && strings.HasPrefix(entry.Name(), "go") {
            versions = append(versions, entry.Name())
        }
    }

    // 排序版本
    sort.Strings(versions)

    return versions, nil
}
```

**运行所有测试**: 全部通过

```bash
$ go test -v ./pkg/version
=== RUN   TestGetInstalledVersions
--- PASS: TestGetInstalledVersions (0.00s)
=== RUN   TestGetInstalledVersions_EmptyDirectory
--- PASS: TestGetInstalledVersions_EmptyDirectory (0.00s)
=== RUN   TestGetInstalledVersions_NonExistentDirectory
--- PASS: TestGetInstalledVersions_NonExistentDirectory (0.00s)
PASS
ok  pkg/version 0.003s
```

### 第 5 步: 添加边界测试

```go
func TestGetInstalledVersions_WithNonGoDirectories(t *testing.T) {
    tmpDir := t.TempDir()
    versionsDir := filepath.Join(tmpDir, "versions")

    // 创建混合内容
    os.MkdirAll(filepath.Join(versionsDir, "go1.20.0"), 0755)
    os.MkdirAll(filepath.Join(versionsDir, "temp"), 0755)
    os.MkdirAll(filepath.Join(versionsDir, "golang"), 0755)

    versions, err := GetInstalledVersions(versionsDir)

    if err != nil {
        t.Fatalf("GetInstalledVersions() error = %v", err)
    }

    if len(versions) != 1 {
        t.Errorf("GetInstalledVersions() returned %d versions, want 1", len(versions))
    }

    if versions[0] != "go1.20.0" {
        t.Errorf("versions[0] = %v, want go1.20.0", versions[0])
    }
}

func TestGetInstalledVersions_Ordering(t *testing.T) {
    tmpDir := t.TempDir()
    versionsDir := filepath.Join(tmpDir, "versions")

    // 创建乱序的版本
    os.MkdirAll(filepath.Join(versionsDir, "go1.21.0"), 0755)
    os.MkdirAll(filepath.Join(versionsDir, "go1.19.0"), 0755)
    os.MkdirAll(filepath.Join(versionsDir, "go1.20.0"), 0755)

    versions, err := GetInstalledVersions(versionsDir)

    if err != nil {
        t.Fatalf("GetInstalledVersions() error = %v", err)
    }

    expected := []string{"go1.19.0", "go1.20.0", "go1.21.0"}
    for i, v := range versions {
        if v != expected[i] {
            t.Errorf("versions[%d] = %v, want %v", i, v, expected[i])
        }
    }
}
```

**运行测试**: 确保所有边界情况都覆盖

```bash
$ go test -v -cover ./pkg/version
=== RUN   TestGetInstalledVersions
=== RUN   TestGetInstalledVersions_EmptyDirectory
=== RUN   TestGetInstalledVersions_NonExistentDirectory
=== RUN   TestGetInstalledVersions_WithNonGoDirectories
=== RUN   TestGetInstalledVersions_Ordering
--- PASS: TestGetInstalledVersions (0.00s)
--- PASS: TestGetInstalledVersions_EmptyDirectory (0.00s)
--- PASS: TestGetInstalledVersions_NonExistentDirectory (0.00s)
--- PASS: TestGetInstalledVersions_WithNonGoDirectories (0.00s)
--- PASS: TestGetInstalledVersions_Ordering (0.00s)
PASS
coverage: 100.0% of statements
ok  pkg/version 0.003s
```

## TDD 工作流程检查清单

### 开始开发前
- [ ] 明确功能需求和预期行为
- [ ] 确定输入和输出
- [ ] 识别边界条件和错误情况

### RED 阶段
- [ ] 编写一个失败的测试用例
- [ ] 确保测试描述清晰具体
- [ ] 运行测试确认失败

### GREEN 阶段
- [ ] 编写最简单的代码使测试通过
- [ ] 不要担心代码质量
- [ ] 运行测试确认通过

### REFACTOR 阶段
- [ ] 改进代码结构和可读性
- [ ] 提取辅助函数
- [ ] 优化性能
- [ ] 确保所有测试仍然通过

### 循环继续
- [ ] 识别下一个功能点
- [ ] 重复 RED-GREEN-REFACTOR 循环

## 参考资源

- [Test-Driven Development by Example](https://www.amazon.com/Test-Driven-Development-Kent-Beck/dp/0321146530)
- [Growing Object-Oriented Software, Guided by Tests](https://www.amazon.com/Growing-Object-Oriented-Software-Guided-Tests/dp/0321503627)
- [Go Testing Blog](https://blog.golang.org/tour-testing)
- [Effective Go: Testing](https://golang.org/doc/effective_go.html#testing)
