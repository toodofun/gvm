# 测试框架指南

## 概述

GVM 项目使用 `testify` 测试框架，提供了一套完整的测试工具和辅助函数，帮助开发者编写高质量、可维护的测试用例。

## 测试框架组件

### 1. 测试环境设置

**测试辅助工具包**: `pkg/testutil`

```go
// 测试辅助函数
func SetupTest(t *testing.T) (*TestEnv, func())
func CreateTempDir(t *testing.T) string
func CreateTempFile(t *testing.T, content string) string
```

**功能**:
- 创建隔离的测试环境
- 自动清理临时文件
- 提供统一的测试配置

### 2. 测试夹具 (Fixtures)

**位置**: `testdata/`

```
testdata/
├── versions/
│   ├── go1.20.0.darwin-arm64.tar.gz
│   └── go1.21.0.darwin-arm64.tar.gz
├── scripts/
│   └── mock-install.sh
└── configs/
    └── test-config.yaml
```

**用途**:
- 存储测试用的版本文件
- 提供模拟脚本和配置
- 支持跨平台测试

### 3. 断言库

```go
import "github.com/stretchr/testify/assert"
import "github.com/stretchr/testify/require"

// assert - 失败后继续执行
assert.Equal(t, expected, actual)
assert.NoError(t, err)
assert.True(t, condition)

// require - 失败后立即停止
require.Equal(t, expected, actual)
require.NoError(t, err)
require.FileExists(t, filepath)
```

### 4. Mock 支持

```go
// HTTP Mock
import "github.com/stretchr/testify/http"

// Mock HTTP 请求
mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    w.WriteHeader(http.StatusOK)
    w.Write([]byte("response"))
}))
defer mockServer.Close()
```

### 5. 测试套件

```go
import "github.com/stretchr/testify/suite"

type ExampleTestSuite struct {
    suite.Suite
    env *TestEnv
}

func (s *ExampleTestSuite) SetupTest() {
    s.env = SetupTest(s.T())
}

func (s *ExampleTestSuite) TearDownTest() {
    cleanup(s.env)
}

func TestExampleTestSuite(t *testing.T) {
    suite.Run(t, new(ExampleTestSuite))
}
```

## 测试类型

### 1. 单元测试

**目标**: 测试单个函数或方法

**特征**:
- 快速执行 (< 1ms)
- 无外部依赖
- 隔离性好

**示例**:

```go
func TestParseVersion(t *testing.T) {
    tests := []struct {
        name    string
        input   string
        want    string
        wantErr bool
    }{
        {
            name:    "valid version",
            input:   "go1.20.0",
            want:    "1.20.0",
            wantErr: false,
        },
        {
            name:    "invalid version",
            input:   "invalid",
            want:    "",
            wantErr: true,
        },
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

### 2. 集成测试

**目标**: 测试多个组件协同工作

**特征**:
- 中等执行速度 (< 100ms)
- 可能有文件系统依赖
- 测试组件间接口

**示例**:

```go
func TestDownloadAndInstall(t *testing.T) {
    // 创建测试环境
    tmpDir := testutil.CreateTempDir(t)
    defer os.RemoveAll(tmpDir)

    // 创建 mock HTTP 服务器
    server := testutil.NewMockServer(t)
    defer server.Close()

    // 测试下载和安装流程
    downloader := NewDownloader()
    installer := NewInstaller(tmpDir)

    url := server.URL + "/go1.20.0.darwin-arm64.tar.gz"
    archivePath, err := downloader.Download(url, tmpDir)
    require.NoError(t, err)

    err = installer.Install(archivePath)
    require.NoError(t, err)

    // 验证安装结果
    installedPath := filepath.Join(tmpDir, "go", "bin", "go")
    assert.FileExists(t, installedPath)
}
```

### 3. 端到端测试 (E2E)

**目标**: 测试完整用户场景

**特征**:
- 执行时间较长
- 真实环境或高保真模拟
- 测试用户工作流

**示例**:

```go
func TestE2E_InstallWorkflow(t *testing.T) {
    if testing.Short() {
        t.Skip("skipping E2E test in short mode")
    }

    // 创建临时环境
    homeDir := testutil.CreateTempDir(t)
    defer os.RemoveAll(homeDir)

    // 设置环境变量
    t.Setenv("HOME", homeDir)
    t.Setenv("GVM_ROOT", filepath.Join(homeDir, ".gvm"))

    // 执行完整的安装流程
    cmd := exec.Command("gvm", "install", "go1.20.0")
    output, err := cmd.CombinedOutput()
    require.NoError(t, err, "output: %s", string(output))

    // 验证安装结果
    cmd = exec.Command("gvm", "use", "go1.20.0")
    output, err = cmd.CombinedOutput()
    require.NoError(t, err)

    cmd = exec.Command("go", "version")
    output, err = cmd.CombinedOutput()
    require.NoError(t, err)
    assert.Contains(t, string(output), "go1.20.0")
}
```

## 运行测试

### 基本命令

```bash
# 运行所有测试
go test ./...

# 运行特定包的测试
go test ./pkg/gvm

# 运行特定测试函数
go test -run TestDownload ./pkg/gvm

# 运行特定子测试
go test -run TestDownload/Success ./pkg/gvm

# 详细输出
go test -v ./...

# 显示覆盖率
go test -cover ./...
```

### 测试模式

```bash
# 短模式 (跳过慢速测试)
go test -short ./...

# 并行测试
go test -parallel 4 ./...

# 超时设置
go test -timeout 30s ./...

# 竞态检测
go test -race ./...
```

### 覆盖率报告

```bash
# 生成覆盖率报告
go test -coverprofile=coverage.out ./...

# 查看覆盖率
go tool cover -func=coverage.out

# 生成 HTML 报告
go tool cover -html=coverage.out -o coverage.html

# 按包查看覆盖率
go test -coverprofile=coverage.out ./pkg/...
go tool cover -func=coverage.out | grep total
```

### 基准测试

```bash
# 运行基准测试
go test -bench=. ./...

# 运行特定基准测试
go test -bench=BenchmarkDownload ./pkg/gvm

# 设置运行时间
go test -bench=. -benchtime=10s ./...

# 运行多次以获得稳定结果
go test -bench=. -count=5 ./...

# 内存分析
go test -bench=. -benchmem ./...
```

## 覆盖率目标

### 整体目标
- **整体覆盖率**: ≥ 70%
- **核心模块**: ≥ 85%
- **工具函数**: ≥ 90%

### 分类目标
- **pkg/gvm**: ≥ 85% (核心功能)
- **pkg/version**: ≥ 90% (版本管理)
- **pkg/http**: ≥ 80% (网络请求)
- **pkg/installer**: ≥ 85% (安装逻辑)
- **cmd/**: ≥ 60% (命令行工具)

### 覆盖率检查

```bash
# 检查整体覆盖率
go test -coverprofile=coverage.out ./...
go tool cover -func=coverage.out | tail -1

# 检查特定包
go test -coverprofile=coverage.out ./pkg/gvm
go tool cover -func=coverage.out | grep total
```

## 最佳实践

### 1. 测试命名

```go
// 好的命名
func TestDownload_WithValidURL_Success(t *testing.T)
func TestInstall_WithInvalidArchive_Error(t *testing.T)

// 避免的命名
func TestDownload1(t *testing.T)
func TestDownloadSuccess(t *testing.T) // 不够具体
```

### 2. 表驱动测试

```go
func TestValidateVersion(t *testing.T) {
    tests := []struct {
        name    string
        version string
        want    bool
    }{
        {"valid version", "go1.20.0", true},
        {"invalid version", "1.20.0", false},
        {"empty string", "", false},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got := ValidateVersion(tt.version)
            assert.Equal(t, tt.want, got)
        })
    }
}
```

### 3. 测试隔离

```go
func TestWriteConfig(t *testing.T) {
    // 每个测试使用独立的临时目录
    tmpDir := testutil.CreateTempDir(t)
    defer os.RemoveAll(tmpDir)

    configPath := filepath.Join(tmpDir, "config.yaml")
    err := WriteConfig(configPath, config)
    assert.NoError(t, err)
}
```

### 4. Mock 外部依赖

```go
func TestDownload(t *testing.T) {
    // 使用 mock HTTP 服务器
    server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        assert.Equal(t, "/go1.20.0.darwin-arm64.tar.gz", r.URL.Path)
        w.WriteHeader(http.StatusOK)
        w.Write([]byte("test data"))
    }))
    defer server.Close()

    downloader := NewDownloader()
    _, err := downloader.Download(server.URL+"/go1.20.0.darwin-arm64.tar.gz", tmpDir)
    assert.NoError(t, err)
}
```

### 5. 错误测试

```go
func TestInstall_WithError(t *testing.T) {
    tests := []struct {
        name      string
        archive   string
        wantErr   string
        errorCode string
    }{
        {
            name:      "file not found",
            archive:   "/nonexistent/file.tar.gz",
            wantErr:   "file not found",
            errorCode: "FILE_NOT_FOUND",
        },
        {
            name:      "invalid archive",
            archive:   "invalid.tar.gz",
            wantErr:   "invalid archive format",
            errorCode: "INVALID_ARCHIVE",
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := Install(tt.archive)
            assert.Error(t, err)
            assert.Contains(t, err.Error(), tt.wantErr)
        })
    }
}
```

### 6. 并行测试

```go
func TestParallelOperations(t *testing.T) {
    t.Parallel()

    // 测试代码
}

func TestMain(m *testing.M) {
    // 全局设置
    os.Exit(m.Run())
}
```

### 7. 测试辅助函数

```go
// 创建可重用的测试辅助函数
func assertInstalled(t *testing.T, installPath string) {
    t.Helper()

    binaryPath := filepath.Join(installPath, "go", "bin", "go")
    assert.FileExists(t, binaryPath)

    info, err := os.Stat(binaryPath)
    assert.NoError(t, err)
    assert.True(t, info.Mode().Perm()&0111 != 0, "binary should be executable")
}

func createMockArchive(t *testing.T, content string) string {
    t.Helper()

    tmpFile := testutil.CreateTempFile(t, content)
    return tmpFile
}
```

## 调试测试

### 使用 Print 调试

```go
func TestDownload(t *testing.T) {
    t.Log("Starting download test")
    // 测试代码
    t.Logf("Downloaded file: %s", path)
}
```

### 使用 Delve 调试

```bash
# 安装 delve
go install github.com/go-delve/delve/cmd/dlv@latest

# 调试测试
dlv test ./pkg/gvm -- -test.v -test.run TestDownload
```

### 使用 VS Code 调试

```json
{
    "version": "0.2.0",
    "configurations": [
        {
            "name": "Debug Test",
            "type": "go",
            "request": "launch",
            "mode": "test",
            "program": "${workspaceFolder}/pkg/gvm",
            "args": ["-test.run", "TestDownload"]
        }
    ]
}
```

## 持续集成

### CI 测试流程

```yaml
# .github/workflows/test.yml
name: Tests
on: [push, pull_request]

jobs:
  test:
    runs-on: ${{ matrix.os }}
    strategy:
      matrix:
        os: [ubuntu-latest, macos-latest, windows-latest]
        go-version: [1.20, 1.21, 1.22]

    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v4
        with:
          go-version: ${{ matrix.go-version }}

      - name: Run tests
        run: go test -v -race -coverprofile=coverage.out ./...

      - name: Upload coverage
        uses: codecov/codecov-action@v3
        with:
          files: ./coverage.out
```

### 测试检查清单

在提交代码前，确保:
- [ ] 所有测试通过 (`go test ./...`)
- [ ] 代码覆盖率达标 (≥ 70%)
- [ ] 无竞态条件 (`go test -race ./...`)
- [ ] 基准测试性能无退化
- [ ] 集成测试在多平台通过

## 参考资源

- [Go Testing 官方文档](https://golang.org/pkg/testing/)
- [Testify 文档](https://github.com/stretchr/testify)
- [TableDrivenTests](https://github.com/golang/go/wiki/TableDrivenTests)
- [Go Test Patterns](https://github.com/golang/go/wiki/TestComments)
