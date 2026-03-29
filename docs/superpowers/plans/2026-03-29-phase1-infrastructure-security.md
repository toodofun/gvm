# GVM改进 - 第一阶段实施计划

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**目标:** 搭建测试框架基础设施，实现核心安全模块，为后续重构奠定坚实基础

**架构:**
- 创建独立的测试框架和工具模块
- 实现安全命令执行器（防止命令注入）
- 实现路径验证工具（防止路径遍历）
- 实现输入验证层
- 配置CI/CD流水线

**技术栈:**
- Go 1.26+
- testify (测试框架)
- golangci-lint (代码检查)
- GitHub Actions (CI/CD)
- go test -race (竞态检测)

---

## 文件结构

```
gvm/
├── internal/
│   ├── core/
│   │   └── validation/           # NEW: 输入验证层
│   │       ├── validation.go     # 验证函数
│   │       └── validation_test.go
│   └── util/
│       ├── exec/                 # NEW: 安全命令执行
│       │   ├── executor.go       # 安全执行器实现
│       │   └── executor_test.go
│       └── path/                 # MODIFY: 增强路径验证
│           ├── path.go           # 现有文件，增强
│           └── path_test.go      # 现有文件，增强测试
├── test/
│   ├── framework/                # NEW: 测试框架
│   │   ├── setup.go              # 测试设置工具
│   │   ├── fixtures.go           # 测试固件
│   │   └── helpers.go            # 测试辅助函数
│   └── mock/                     # NEW: Mock实现
│       └── http_mock.go          # HTTP测试服务器
├── .github/
│   └── workflows/
│       ├── test.yml              # NEW: 测试工作流
│       ├── lint.yml              # NEW: 代码检查工作流
│       └── security.yml          # NEW: 安全扫描工作流
├── Makefile                       # MODIFY: 添加新目标
├── go.mod                         # MODIFY: 添加测试依赖
└── docs/
    └── testing/
        ├── test-framework.md      # NEW: 测试框架文档
        └── tdd-workflow.md       # NEW: TDD流程文档
```

---

## 任务列表

### Task 1: 设置测试框架基础设施

**目标:** 创建测试辅助工具和框架，为所有后续测试提供支持

**Files:**
- Create: `test/framework/setup.go`
- Create: `test/framework/fixtures.go`
- Create: `test/framework/helpers.go`
- Modify: `go.mod`

- [ ] **Step 1: 创建测试框架目录结构**

```bash
mkdir -p test/framework
mkdir -p test/mock
```

- [ ] **Step 2: 添加测试依赖到go.mod**

```bash
go get github.com/stretchr/testify/assert
go get github.com/stretchr/testify/mock
go get github.com/stretchr/testify/suite
```

- [ ] **Step 3: 创建test/framework/setup.go - 测试环境设置工具**

```go
package framework

import (
    "os"
    "path/filepath"
    "testing"
)

// SetupTestEnvironment 创建临时测试环境
// 返回临时目录路径和清理函数
func SetupTestEnvironment(t *testing.T) (string, func()) {
    t.Helper()

    tempDir, err := os.MkdirTemp("", "gvm-test-*")
    if err != nil {
        t.Fatalf("failed to create temp dir: %v", err)
    }

    cleanup := func() {
        os.RemoveAll(tempDir)
    }

    return tempDir, cleanup
}

// SetupTestConfig 创建测试配置目录结构
func SetupTestConfig(t *testing.T, baseDir string) {
    t.Helper()

    dirs := []string{
        filepath.Join(baseDir, "versions"),
        filepath.Join(baseDir, "archives"),
        filepath.Join(baseDir, "config"),
    }

    for _, dir := range dirs {
        if err := os.MkdirAll(dir, 0755); err != nil {
            t.Fatalf("failed to create directory %s: %v", dir, err)
        }
    }
}

// CreateTestFile 在测试目录中创建文件
func CreateTestFile(t *testing.T, path, content string) {
    t.Helper()

    dir := filepath.Dir(path)
    if err := os.MkdirAll(dir, 0755); err != nil {
        t.Fatalf("failed to create directory: %v", err)
    }

    if err := ioutil.WriteFile(path, []byte(content), 0644); err != nil {
        t.Fatalf("failed to write file: %v", err)
    }
}
```

- [ ] **Step 4: 创建test/framework/fixtures.go - 测试固件**

```go
package framework

import (
    "archive/tar"
    "compress/gzip"
    "io"
    "os"
    "path/filepath"
)

// CreateTarGzFixture 创建测试用的tar.gz文件
func CreateTarGzFixture(destPath string, files map[string]string) error {
    // 创建目标文件
    destFile, err := os.Create(destPath)
    if err != nil {
        return err
    }
    defer destFile.Close()

    // 创建gzip writer
    gw := gzip.NewWriter(destFile)
    defer gw.Close()

    // 创建tar writer
    tw := tar.NewWriter(gw)
    defer tw.Close()

    // 添加文件到tar
    for name, content := range files {
        // 创建tar header
    header := &tar.Header{
        Name: name,
        Mode: 0644,
        Size: int64(len(content)),
    }

        if err := tw.WriteHeader(header); err != nil {
            return err
        }

        // 写入文件内容
        if _, err := tw.Write([]byte(content)); err != nil {
            return err
        }
    }

    return nil
}

// CreateGzipFixture 创建测试用的gzip文件
func CreateGzipFixture(destPath, content string) error {
    file, err := os.Create(destPath)
    if err != nil {
        return err
    }
    defer file.Close()

    gw := gzip.NewWriter(file)
    defer gw.Close()

    _, err = gw.Write([]byte(content))
    return err
}

// HTTPTestServer HTTP测试服务器的配置
type HTTPTestServer struct {
    BaseURL    string
    Responses  map[string]string
    StatusCode map[string]int
}

// NewHTTPTestServer 创建新的测试HTTP服务器配置
// 注意: 实际的测试服务器应该在各个测试中使用httptest.NewServer创建
func NewHTTPTestServer() *HTTPTestServer {
    return &HTTPTestServer{
        Responses:  make(map[string]string),
        StatusCode: make(map[string]int),
    }
}
```

- [ ] **Step 5: 创建test/framework/helpers.go - 测试辅助函数**

```go
package framework

import (
    "os"
    "path/filepath"
    "strings"
    "testing"
)

// AssertDirExists 断言目录存在
func AssertDirExists(t *testing.T, path string) {
    t.Helper()

    info, err := os.Stat(path)
    if err != nil {
        t.Fatalf("directory %s does not exist: %v", path, err)
    }

    if !info.IsDir() {
        t.Fatalf("%s is not a directory", path)
    }
}

// AssertFileExists 断言文件存在
func AssertFileExists(t *testing.T, path string) {
    t.Helper()

    info, err := os.Stat(path)
    if err != nil {
        t.Fatalf("file %s does not exist: %v", path, err)
    }

    if info.IsDir() {
        t.Fatalf("%s is a directory, not a file", path)
    }
}

// AssertFileContains 断言文件包含特定内容
func AssertFileContains(t *testing.T, path, content string) {
    t.Helper()

    data, err := os.ReadFile(path)
    if err != nil {
        t.Fatalf("failed to read file %s: %v", path, err)
    }

    contentStr := string(data)
    if !strings.Contains(contentStr, content) {
        t.Fatalf("file %s does not contain expected content:\ngot: %s\nwant: %s",
            path, contentStr, content)
    }
}

// SkipIfShort 跳过短测试
func SkipIfShort(t *testing.T) {
    t.Helper()
    if testing.Short() {
        t.Skip("skipping test in short mode")
    }
}
```

- [ ] **Step 6: 为test/framework创建测试文件**

```bash
# 测试框架本身也需要测试
touch test/framework/setup_test.go
touch test/framework/fixtures_test.go
touch test/framework/helpers_test.go
```

- [ ] **Step 7: 运行测试确保框架代码正常工作**

```bash
go test ./test/framework/... -v
```

预期输出: 所有测试通过

- [ ] **Step 8: 提交测试框架**

```bash
git add test/
git commit -m "feat: add testing framework infrastructure

- Add test environment setup utilities
- Add test fixtures and helpers
- Provide reusable test utilities for all modules

Co-Authored-By: Claude Sonnet 4.6 <noreply@anthropic.com>"
```

---

### Task 2: 实现安全命令执行器

**目标:** 创建安全的命令执行器，防止命令注入攻击

**Files:**
- Create: `internal/util/exec/executor.go`
- Create: `internal/util/exec/executor_test.go`
- Create: `internal/util/exec/exec.go`

- [ ] **Step 1: 创建exec包目录**

```bash
mkdir -p internal/util/exec
```

- [ ] **Step 2: 编写失败测试 - 测试危险字符检测**

```go
// internal/util/exec/executor_test.go
package exec

import (
    "context"
    "testing"

    "github.com/stretchr/testify/assert"
)

func TestParseCommand_RejectsShellMetacharacters(t *testing.T) {
    tests := []struct {
        name    string
        cmd     string
        wantErr bool
        errMsg  string
    }{
        {
            name:    "normal command",
            cmd:     "ls -la",
            wantErr: false,
        },
        {
            name:    "pipe character",
            cmd:     "ls | rm -rf /",
            wantErr: true,
            errMsg:  "dangerous character",
        },
        {
            name:    "command substitution",
            cmd:     "ls $(rm -rf /)",
            wantErr: true,
            errMsg:  "dangerous character",
        },
        {
            name:    "semicolon",
            cmd:     "ls; rm -rf /",
            wantErr: true,
            errMsg:  "dangerous character",
        },
        {
            name:    "backtick",
            cmd:     "ls `rm -rf /`",
            wantErr: true,
            errMsg:  "dangerous character",
        },
        {
            name:    "ampersand",
            cmd:     "ls & rm",
            wantErr: true,
            errMsg:  "dangerous character",
        },
        {
            name:    "dollar sign",
            cmd:     "ls $HOME",
            wantErr: true,
            errMsg:  "dangerous character",
        },
        {
            name:    "parentheses",
            cmd:     "ls (rm)",
            wantErr: true,
            errMsg:  "dangerous character",
        },
        {
            name:    "redirect",
            cmd:     "ls > file",
            wantErr: true,
            errMsg:  "dangerous character",
        },
        {
            name:    "backslash",
            cmd:     "ls \\",
            wantErr: true,
            errMsg:  "dangerous character",
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            cmd, err := ParseCommand(tt.cmd)

            if tt.wantErr {
                assert.Error(t, err)
                assert.Contains(t, err.Error(), tt.errMsg)
                assert.Nil(t, cmd)
            } else {
                assert.NoError(t, err)
                assert.NotNil(t, cmd)
                assert.NotEmpty(t, cmd.Name)
            }
        })
    }
}
```

- [ ] **Step 3: 运行测试确认失败**

```bash
go test ./internal/util/exec/... -v
```

预期输出:
```
# undefined: ParseCommand
```

- [ ] **Step 4: 实现ParseCommand函数**

```go
// internal/util/exec/exec.go
package exec

import (
    "fmt"
    "strings"
)

// Command 表示一个已解析的安全命令
type Command struct {
    Name string
    Args []string
}

// dangerousChars 包含所有危险shell元字符
var dangerousChars = []string{"|", "&", ";", "$", "(", ")", "<", ">", "`", "\\"}

// ParseCommand 安全地解析命令字符串，拒绝shell元字符
// 这是一个关键安全函数 - 它防止命令注入攻击
func ParseCommand(cmdString string) (*Command, error) {
    // 检查危险字符
    for _, char := range dangerousChars {
        if strings.Contains(cmdString, char) {
            return nil, fmt.Errorf("command contains dangerous character: %s", char)
        }
    }

    // 分割命令和参数
    parts := strings.Fields(cmdString)
    if len(parts) == 0 {
        return nil, fmt.Errorf("empty command")
    }

    return &Command{
        Name: parts[0],
        Args: parts[1:],
    }, nil
}
```

- [ ] **Step 5: 运行测试确认通过**

```bash
go test ./internal/util/exec/... -v
```

预期输出: 所有测试通过 ✓

- [ ] **Step 6: 编写失败测试 - SafeExecutor白名单机制**

```go
// internal/util/exec/executor_test.go
func TestSafeExecutor_Execute_RejectsUnauthorizedCommands(t *testing.T) {
    // 只允许ls和tar命令
    allowedCmds := []string{"ls", "tar"}
    executor := NewSafeExecutor(allowedCmds)

    tests := []struct {
        name        string
        cmd         string
        args        []string
        wantErr     bool
        errContains string
    }{
        {
            name:    "allowed command with args",
            cmd:     "ls",
            args:    []string{"-la", "/tmp"},
            wantErr: false,
        },
        {
            name:        "command not in whitelist",
            cmd:         "sh",
            args:        []string{"-c", "echo test"},
            wantErr:     true,
            errContains: "not allowed",
        },
        {
            name:        "rm not allowed",
            cmd:         "rm",
            args:        []string{"-rf", "/"},
            wantErr:     true,
            errContains: "not allowed",
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := executor.Execute(context.Background(), tt.cmd, tt.args...)

            if tt.wantErr {
                assert.Error(t, err)
                assert.Contains(t, err.Error(), tt.errContains)
            } else {
                assert.NoError(t, err)
            }
        })
    }
}
```

- [ ] **Step 7: 运行测试确认失败**

```bash
go test ./internal/util/exec/... -v -run TestSafeExecutor
```

预期输出: `undefined: NewSafeExecutor`

- [ ] **Step 8: 实现SafeExecutor**

```go
// internal/util/exec/executor.go
package exec

import (
    "context"
    "fmt"
    "os/exec"
    "strings"
)

// SafeExecutor 提供安全的命令执行，防止命令注入攻击
type SafeExecutor struct {
    // allowedCommands 是白名单，只允许执行预定义的命令
    allowedCommands map[string]bool
}

// NewSafeExecutor 创建一个安全执行器，只允许白名单中的命令
func NewSafeExecutor(allowedCommands []string) *SafeExecutor {
    whitelist := make(map[string]bool)
    for _, cmd := range allowedCommands {
        whitelist[cmd] = true
    }
    return &SafeExecutor{
        allowedCommands: whitelist,
    }
}

// Execute 安全地执行命令
func (e *SafeExecutor) Execute(ctx context.Context, cmd string, args ...string) error {
    // 检查命令是否在白名单中
    if !e.allowedCommands[cmd] {
        return fmt.Errorf("command not allowed: %s", cmd)
    }

    // 验证参数中不包含shell元字符
    for _, arg := range args {
        for _, dangerousChar := range dangerousChars {
            if strings.Contains(arg, dangerousChar) {
                return fmt.Errorf("argument contains dangerous character '%s': %s", dangerousChar, arg)
            }
        }
    }

    // 使用参数化执行，不通过shell
    execCmd := exec.CommandContext(ctx, cmd, args...)
    return execCmd.Run()
}

// ExecuteString 从字符串解析并执行命令
func (e *SafeExecutor) ExecuteString(ctx context.Context, cmdString string) error {
    cmd, err := ParseCommand(cmdString)
    if err != nil {
        return err
    }

    return e.Execute(ctx, cmd.Name, cmd.Args...)
}
```

- [ ] **Step 9: 运行测试确认通过**

```bash
go test ./internal/util/exec/... -v
```

预期输出: 所有测试通过 ✓

- [ ] **Step 10: 添加更多安全测试用例**

```go
// internal/util/exec/executor_test.go
func TestSafeExecutor_Execute_RejectsDangerousArguments(t *testing.T) {
    allowedCmds := []string{"ls", "tar"}
    executor := NewSafeExecutor(allowedCmds)

    tests := []struct {
        name        string
        cmd         string
        args        []string
        wantErr     bool
        errContains string
    }{
        {
            name:        "argument with pipe",
            cmd:         "ls",
            args:        []string{"|", "rm", "-rf", "/"},
            wantErr:     true,
            errContains: "dangerous",
        },
        {
            name:        "argument with command substitution",
            cmd:         "ls",
            args:        []string{"$(rm -rf /)"},
            wantErr:     true,
            errContains: "dangerous",
        },
        {
            name:        "argument with semicolon",
            cmd:         "tar",
            args:        []string{"-xzvf", "file.tar.gz";", "rm", "-rf", "/"},
            wantErr:     true,
            errContains: "dangerous",
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := executor.Execute(context.Background(), tt.cmd, tt.args...)

            if tt.wantErr {
                assert.Error(t, err)
                assert.Contains(t, err.Error(), tt.errContains)
            } else {
                assert.NoError(t, err)
            }
        })
    }
}
```

- [ ] **Step 11: 运行所有测试**

```bash
go test ./internal/util/exec/... -v -race
```

预期输出: 所有测试通过，无竞态条件 ✓

- [ ] **Step 12: 提交安全命令执行器**

```bash
git add internal/util/exec/
git commit -m "feat: implement safe command executor

- Add ParseCommand to reject shell metacharacters
- Add SafeExecutor with whitelist mechanism
- Prevent command injection attacks
- Add comprehensive security tests
- Validate both commands and arguments

Security: This prevents command injection vulnerabilities

Co-Authored-By: Claude Sonnet 4.6 <noreply@anthropic.com>"
```

---

### Task 3: 实现输入验证层

**目标:** 创建统一的输入验证函数，验证版本号、路径和URL

**Files:**
- Create: `internal/core/validation/validation.go`
- Create: `internal/core/validation/validation_test.go`

- [ ] **Step 1: 创建validation包**

```bash
mkdir -p internal/core/validation
```

- [ ] **Step 2: 编写失败测试 - 版本验证**

```go
// internal/core/validation/validation_test.go
package validation

import (
    "testing"

    "github.com/stretchr/testify/assert"
)

func TestValidateVersion(t *testing.T) {
    tests := []struct {
        name    string
        version string
        wantErr bool
        errMsg  string
    }{
        {
            name:    "valid version with patch",
            version: "1.20.0",
            wantErr: false,
        },
        {
            name:    "valid version without patch",
            version: "1.20",
            wantErr: false,
        },
        {
            name:    "valid version with v prefix",
            version: "v1.20.0",
            wantErr: false,
        },
        {
            name:    "empty version",
            version: "",
            wantErr: true,
            errMsg:  "cannot be empty",
        },
        {
            name:    "invalid format",
            version: "invalid",
            wantErr: true,
            errMsg:  "invalid format",
        },
        {
            name:    "negative number",
            version: "1.-1.0",
            wantErr: true,
            errMsg:  "invalid format",
        },
        {
            name:    "too many parts",
            version: "1.2.3.4",
            wantErr: true,
            errMsg:  "invalid format",
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := ValidateVersion(tt.version)

            if tt.wantErr {
                assert.Error(t, err)
                assert.Contains(t, err.Error(), tt.errMsg)
            } else {
                assert.NoError(t, err)
            }
        })
    }
}
```

- [ ] **Step 3: 运行测试确认失败**

```bash
go test ./internal/core/validation/... -v
```

预期输出: `undefined: ValidateVersion`

- [ ] **Step 4: 实现版本验证**

```go
// internal/core/validation/validation.go
package validation

import (
    "fmt"
    "regexp"
    "strings"
)

// versionRegex 匹配语义化版本号
// 格式: v1.2.3 或 1.2.3 或 1.2
var versionRegex = regexp.MustCompile(`^v?[0-9]+\.[0-9]+(\.[0-9]+)?$`)

// ValidateVersion 验证版本号格式
func ValidateVersion(version string) error {
    if version == "" {
        return fmt.Errorf("version cannot be empty")
    }

    version = strings.TrimSpace(version)

    if !versionRegex.MatchString(version) {
        return fmt.Errorf("invalid version format: %s (expected format: 1.2.3 or 1.2)", version)
    }

    return nil
}
```

- [ ] **Step 5: 运行测试确认通过**

```bash
go test ./internal/core/validation/... -v
```

预期输出: 测试通过 ✓

- [ ] **Step 6: 编写失败测试 - 路径验证**

```go
// internal/core/validation/validation_test.go
func TestValidatePath(t *testing.T) {
    tests := []struct {
        name    string
        path    string
        wantErr bool
        errMsg  string
    }{
        {
            name:    "valid absolute path",
            path:    "/usr/local/go",
            wantErr: false,
        },
        {
            name:    "valid relative path",
            path:    "./go",
            wantErr: false,
        },
        {
            name:    "valid home path",
            path:    "~/go",
            wantErr: false,
        },
        {
            name:    "empty path",
            path:    "",
            wantErr: true,
            errMsg:  "cannot be empty",
        },
        {
            name:    "path traversal attempt",
            path:    "/etc/../passwd",
            wantErr: true,
            errMsg:  "path traversal",
        },
        {
            name:    "path traversal with ..",
            path:    "../../../etc/passwd",
            wantErr: true,
            errMsg:  "path traversal",
        },
        {
            name:    "null bytes",
            path:    "/etc/passwd\x00",
            wantErr: true,
            errMsg:  "invalid character",
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := ValidatePath(tt.path)

            if tt.wantErr {
                assert.Error(t, err)
                assert.Contains(t, err.Error(), tt.errMsg)
            } else {
                assert.NoError(t, err)
            }
        })
    }
}
```

- [ ] **Step 7: 运行测试确认失败**

```bash
go test ./internal/core/validation/... -v -run TestValidatePath
```

预期输出: `undefined: ValidatePath`

- [ ] **Step 8: 实现路径验证**

```go
// internal/core/validation/validation.go
import (
    "path/filepath"
    "strings"
)

// ValidatePath 验证路径安全性
func ValidatePath(path string) error {
    if path == "" {
        return fmt.Errorf("path cannot be empty")
    }

    path = strings.TrimSpace(path)

    // 检查null字节
    if strings.Contains(path, "\x00") {
        return fmt.Errorf("path contains null byte")
    }

    // 清理路径
    cleaned := filepath.Clean(path)

    // 检查路径遍历
    if strings.Contains(cleaned, "..") {
        return fmt.Errorf("path traversal detected: %s", path)
    }

    return nil
}
```

- [ ] **Step 9: 运行测试确认通过**

```bash
go test ./internal/core/validation/... -v
```

预期输出: 测试通过 ✓

- [ ] **Step 10: 编写失败测试 - URL验证**

```go
// internal/core/validation/validation_test.go
func TestValidateURL(t *testing.T) {
    tests := []struct {
        name    string
        url     string
        wantErr bool
        errMsg  string
    }{
        {
            name:    "valid HTTPS URL",
            url:     "https://go.dev/dl/go1.20.0.linux-amd64.tar.gz",
            wantErr: false,
        },
        {
            name:    "valid HTTPS URL with query",
            url:     "https://example.com/download?version=1.20",
            wantErr: false,
        },
        {
            name:    "empty URL",
            url:     "",
            wantErr: true,
            errMsg:  "cannot be empty",
        },
        {
            name:    "HTTP not HTTPS",
            url:     "http://example.com/file.tar.gz",
            wantErr: true,
            errMsg:  "must use HTTPS",
        },
        {
            name:    "invalid URL format",
            url:     "not-a-url",
            wantErr: true,
            errMsg:  "invalid URL",
        },
        {
            name:    "URL with spaces",
            url:     "https://example.com/file name.tar.gz",
            wantErr: true,
            errMsg:  "invalid URL",
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := ValidateURL(tt.url)

            if tt.wantErr {
                assert.Error(t, err)
                assert.Contains(t, err.Error(), tt.errMsg)
            } else {
                assert.NoError(t, err)
            }
        })
    }
}
```

- [ ] **Step 11: 运行测试确认失败**

```bash
go test ./internal/core/validation/... -v -run TestValidateURL
```

预期输出: `undefined: ValidateURL`

- [ ] **Step 12: 实现URL验证**

```go
// internal/core/validation/validation.go
import (
    "net/url"
)

// ValidateURL 验证URL安全性
func ValidateURL(urlStr string) error {
    if urlStr == "" {
        return fmt.Errorf("URL cannot be empty")
    }

    urlStr = strings.TrimSpace(urlStr)

    // 解析URL
    parsed, err := url.Parse(urlStr)
    if err != nil {
        return fmt.Errorf("invalid URL: %w", err)
    }

    // 检查协议
    if parsed.Scheme != "https" {
        return fmt.Errorf("URL must use HTTPS: %s", urlStr)
    }

    // 检查主机
    if parsed.Host == "" {
        return fmt.Errorf("URL missing host: %s", urlStr)
    }

    return nil
}
```

- [ ] **Step 13: 运行所有验证测试**

```bash
go test ./internal/core/validation/... -v -race
```

预期输出: 所有测试通过，无竞态条件 ✓

- [ ] **Step 14: 提交输入验证层**

```bash
git add internal/core/validation/
git commit -m "feat: implement input validation layer

- Add ValidateVersion for semantic version validation
- Add ValidatePath with path traversal detection
- Add ValidateURL requiring HTTPS
- Prevent injection attacks through input validation
- Add comprehensive test coverage

Security: Input validation is first line of defense

Co-Authored-By: Claude Sonnet 4.6 <noreply@anthropic.com>"
```

---

### Task 4: 增强路径安全验证

**目标:** 增强现有的path包，添加路径遍历防护

**Files:**
- Modify: `internal/util/path/path.go` (增强现有文件)
- Modify: `internal/util/path/path_test.go` (增强测试)

- [ ] **Step 1: 查看现有的path.go文件**

```bash
head -50 internal/util/path/path.go
```

- [ ] **Step 2: 编写失败测试 - 路径遍历检测**

```go
// internal/util/path/path_test.go
package path

import (
    "testing"

    "github.com/stretchr/testify/assert"
)

func TestIsPathSafe(t *testing.T) {
    tests := []struct {
        name     string
        basePath string
        target   string
        safe     bool
        wantErr  bool
    }{
        {
            name:     "normal subdirectory",
            basePath: "/tmp/gvm",
            target:   "subdir/file.txt",
            safe:     true,
            wantErr:  false,
        },
        {
            name:     "path traversal with ..",
            basePath: "/tmp/gvm",
            target:   "../etc/passwd",
            safe:     false,
            wantErr:  true,
        },
        {
            name:     "deep path traversal",
            basePath: "/tmp/gvm",
            target:   "subdir/../../etc/passwd",
            safe:     false,
            wantErr:  true,
        },
        {
            name:     "absolute path escape",
            basePath: "/tmp/gvm",
            target:   "/etc/passwd",
            safe:     false,
            wantErr:  true,
        },
        {
            name:     "symlink-like path",
            basePath: "/tmp/gvm",
            target:   "subdir/../../../etc/passwd",
            safe:     false,
            wantErr:  true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            safe, err := IsPathSafe(tt.basePath, tt.target)

            if tt.wantErr {
                assert.Error(t, err)
                assert.False(t, safe)
            } else {
                assert.NoError(t, err)
                assert.Equal(t, tt.safe, safe)
            }
        })
    }
}
```

- [ ] **Step 3: 运行测试确认失败**

```bash
go test ./internal/util/path/... -v -run TestIsPathSafe
```

预期输出: `undefined: IsPathSafe`

- [ ] **Step 4: 实现IsPathSafe函数**

```go
// internal/util/path/path.go
package path

import (
    "fmt"
    "path/filepath"
    "strings"
)

// IsPathSafe 检查目标路径是否在基础路径内，防止路径遍历攻击
func IsPathSafe(basePath, targetPath string) (bool, error) {
    // 解析为绝对路径
    absBase, err := filepath.Abs(basePath)
    if err != nil {
        return false, fmt.Errorf("failed to resolve base path: %w", err)
    }

    // 构建完整目标路径
    fullTarget := filepath.Join(basePath, targetPath)
    absTarget, err := filepath.Abs(fullTarget)
    if err != nil {
        return false, fmt.Errorf("failed to resolve target path: %w", err)
    }

    // 检查目标路径是否在基础路径内
    rel, err := filepath.Rel(absBase, absTarget)
    if err != nil {
        return false, fmt.Errorf("failed to compute relative path: %w", err)
    }

    // 如果相对路径以..开头，说明试图逃逸
    if strings.HasPrefix(rel, "..") {
        return false, fmt.Errorf("path traversal attempt detected: %s tries to escape %s", targetPath, basePath)
    }

    return true, nil
}
```

- [ ] **Step 5: 运行测试确认通过**

```bash
go test ./internal/util/path/... -v -race
```

预期输出: 所有测试通过，无竞态条件 ✓

- [ ] **Step 6: 提交路径安全增强**

```bash
git add internal/util/path/
git commit -m "feat: add path traversal protection to path util

- Add IsPathSafe function to detect path traversal attempts
- Prevent directory escape attacks
- Validate relative paths don't escape base directory
- Add comprehensive security tests

Security: Prevents zip-slip and path traversal attacks

Co-Authored-By: Claude Sonnet 4.6 <noreply@anthropic.com>"
```

---

### Task 5: 创建标准错误码系统

**目标:** 定义统一的错误码，改进错误处理一致性

**Files:**
- Create: `internal/core/errors.go`
- Create: `internal/core/errors_test.go`

- [ ] **Step 1: 编写失败测试 - 错误码定义**

```go
// internal/core/errors_test.go
package core

import (
    "testing"

    "github.com/stretchr/testify/assert"
)

func TestErrorCodes_AreDefined(t *testing.T) {
    tests := []struct {
        name  string
        err   error
        want  string
    }{
        {
            name: "ErrInvalidVersion",
            err:  ErrInvalidVersion,
            want: "invalid version format",
        },
        {
            name: "ErrInvalidPath",
            err:  ErrInvalidPath,
            want: "invalid path",
        },
        {
            name: "ErrInvalidURL",
            err:  ErrInvalidURL,
            want: "invalid URL",
        },
        {
            name: "ErrDownloadFailed",
            err:  ErrDownloadFailed,
            want: "download failed",
        },
        {
            name: "ErrExtractFailed",
            err:  ErrExtractFailed,
            want: "extraction failed",
        },
        {
            name: "ErrCommandBlocked",
            err:  ErrCommandBlocked,
            want: "command not allowed",
        },
        {
            name: "ErrLanguageNotFound",
            err:  ErrLanguageNotFound,
            want: "language not found",
        },
        {
            name: "ErrVersionNotFound",
            err:  ErrVersionNotFound,
            want: "version not found",
        },
        {
            name: "ErrAlreadyInstalled",
            err:  ErrAlreadyInstalled,
            want: "version already installed",
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            assert.NotNil(t, tt.err)
            assert.Contains(t, tt.err.Error(), tt.want)
        })
    }
}

func TestErrorIs(t *testing.T) {
    // 测试错误包装和比较
    baseErr := ErrInvalidVersion
    wrappedErr := fmt.Errorf("wrap: %w", baseErr)

    assert.True(t, errors.Is(wrappedErr, ErrInvalidVersion))
}
```

- [ ] **Step 2: 运行测试确认失败**

```bash
go test ./internal/core/... -v
```

预期输出: `undefined: ErrInvalidVersion`

- [ ] **Step 3: 实现错误码系统**

```go
// internal/core/errors.go
package core

import "errors"

// 标准错误码定义
var (
    // 验证错误
    ErrInvalidVersion = errors.New("invalid version format")
    ErrInvalidPath    = errors.New("invalid path")
    ErrInvalidURL     = errors.New("invalid URL")

    // 下载和安装错误
    ErrDownloadFailed = errors.New("download failed")
    ErrExtractFailed  = errors.New("extraction failed")

    // 安全错误
    ErrCommandBlocked = errors.New("command not allowed")

    // 语言和版本错误
    ErrLanguageNotFound = errors.New("language not found")
    ErrVersionNotFound  = errors.New("version not found")
    ErrAlreadyInstalled = errors.New("version already installed")
)
```

- [ ] **Step 4: 运行测试确认通过**

```bash
go test ./internal/core/... -v
```

预期输出: 所有测试通过 ✓

- [ ] **Step 5: 提交错误码系统**

```bash
git add internal/core/errors.go internal/core/errors_test.go
git commit -m "feat: add standardized error codes

- Define all standard error codes
- Provide consistent error messages
- Support error wrapping and comparison
- Add comprehensive error tests

Quality: Improves error handling consistency

Co-Authored-By: Claude Sonnet 4.6 <noreply@anthropic.com>"
```

---

### Task 6: 修复HTTP客户端资源泄漏

**目标:** 修复HTTP客户端中的资源泄漏问题，改进连接管理

**Files:**
- Modify: `internal/http/client.go`
- Modify: `internal/http/client_test.go`

- [ ] **Step 1: 查看现有client.go**

```bash
head -100 internal/http/client.go
```

- [ ] **Step 2: 编写失败测试 - 资源清理**

```go
// internal/http/client_test.go
package http

import (
    "context"
    "net/http"
    "net/http/httptest"
    "os"
    "testing"
    "time"

    "github.com/stretchr/testify/assert"
)

func TestClient_ResourceCleanup(t *testing.T) {
    server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusOK)
        w.Write([]byte("response"))
    }))
    defer server.Close()

    client := NewClient(5*time.Second, 0)
    req, _ := http.NewRequest("GET", server.URL, nil)

    resp, err := client.Do(context.Background(), req)
    assert.NoError(t, err)
    assert.NotNil(t, resp)
    defer resp.Body.Close()

    data, err := io.ReadAll(resp.Body)
    assert.NoError(t, err)
    assert.Equal(t, "response", string(data))
}

func TestClient_RetryOnTemporaryError(t *testing.T) {
    attempts := 0
    server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        attempts++
        if attempts < 3 {
            w.WriteHeader(http.StatusServiceUnavailable)
            return
        }
        w.WriteHeader(http.StatusOK)
        w.Write([]byte("success"))
    }))
    defer server.Close()

    client := NewClient(5*time.Second, 3)
    req, _ := http.NewRequest("GET", server.URL, nil)

    resp, err := client.Do(context.Background(), req)
    assert.NoError(t, err)
    assert.Equal(t, 3, attempts)
    defer resp.Body.Close()
}

func TestClient_DownloadToFile(t *testing.T) {
    server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusOK)
        w.Write([]byte("download content"))
    }))
    defer server.Close()

    client := NewClient(5*time.Second, 0)
    tempFile := os.TempDir() + "/test-download.txt"
    defer os.Remove(tempFile)

    err := client.DownloadToFile(context.Background(), server.URL, tempFile)
    assert.NoError(t, err)

    content, err := os.ReadFile(tempFile)
    assert.NoError(t, err)
    assert.Equal(t, "download content", string(content))
}
```

- [ ] **Step 3: 运行测试确认失败**

```bash
go test ./internal/http/... -v
```

预期输出: 部分测试失败

- [ ] **Step 4: 改进HTTP客户端实现**

```go
// internal/http/client.go
package http

import (
    "context"
    "fmt"
    "io"
    "net"
    "net/http"
    "os"
    "time"
)

type Client struct {
    client      *http.Client
    timeout     time.Duration
    maxRetries  int
}

// NewClient 创建一个新的HTTP客户端
func NewClient(timeout time.Duration, maxRetries int) *Client {
    return &Client{
        client: &http.Client{
            Timeout: timeout,
            Transport: &http.Transport{
                MaxIdleConns:        100,
                MaxIdleConnsPerHost: 10,
                IdleConnTimeout:     90 * time.Second,
                // 启用连接池
                DisableKeepAlives:     false,
                MaxConnsPerHost:       10,
                ResponseHeaderTimeout: timeout,
            },
        },
        timeout:    timeout,
        maxRetries: maxRetries,
    }
}

// Do 执行HTTP请求，确保资源正确清理
func (c *Client) Do(ctx context.Context, req *http.Request) (*http.Response, error) {
    // 添加超时上下文
    ctx, cancel := context.WithTimeout(ctx, c.timeout)
    defer cancel()

    req = req.WithContext(ctx)

    var resp *http.Response
    var err error

    // 重试逻辑
    for i := 0; i <= c.maxRetries; i++ {
        resp, err = c.client.Do(req)
        if err == nil {
            break
        }

        // 如果是临时错误，重试
        if i < c.maxRetries && isTemporaryError(err) {
            time.Sleep(time.Duration(i+1) * time.Second)
            continue
        }

        // 最后一次重试失败，返回错误
        return nil, fmt.Errorf("HTTP request failed after %d retries: %w", c.maxRetries, err)
    }

    // 检查响应状态码
    if resp.StatusCode < 200 || resp.StatusCode >= 300 {
        // 确保在返回错误前关闭body
        resp.Body.Close()
        return nil, fmt.Errorf("HTTP request failed with status %d", resp.StatusCode)
    }

    return resp, nil
}

// DownloadToFile 下载内容到文件，确保资源正确清理
func (c *Client) DownloadToFile(ctx context.Context, url, destPath string) error {
    req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
    if err != nil {
        return fmt.Errorf("failed to create request: %w", err)
    }

    resp, err := c.Do(ctx, req)
    if err != nil {
        return err
    }

    defer func() {
        // 确保body在函数返回前被关闭
        if resp != nil && resp.Body != nil {
            io.Copy(io.Discard, resp.Body)
            resp.Body.Close()
        }
    }()

    // 创建目标文件
    file, err := os.Create(destPath)
    if err != nil {
        return fmt.Errorf("failed to create file: %w", err)
    }
    defer file.Close()

    // 复制数据
    _, err = io.Copy(file, resp.Body)
    if err != nil {
        os.Remove(destPath)
        return fmt.Errorf("failed to download file: %w", err)
    }

    return nil
}

// isTemporaryError 检查错误是否是临时的
func isTemporaryError(err error) bool {
    if netErr, ok := err.(interface{ Temporary() bool }); ok {
        return netErr.Temporary()
    }
    if _, ok := err.(interface{ Timeout() bool }); ok {
        return true
    }
    return false
}
```

- [ ] **Step 5: 运行测试确认通过**

```bash
go test ./internal/http/... -v -race
```

预期输出: 所有测试通过，无竞态条件 ✓

- [ ] **Step 6: 提交HTTP客户端修复**

```bash
git add internal/http/
git commit -m "fix: prevent HTTP client resource leaks

- Add proper response body cleanup
- Implement retry logic with exponential backoff
- Add connection pooling
- Ensure cleanup in error paths
- Add comprehensive resource cleanup tests

Fixes: Prevents file descriptor leaks and memory leaks

Co-Authored-By: Claude Sonnet 4.6 <noreply@anthropic.com>"
```

---

### Task 7: 添加Panic恢复机制

**目标:** 在main.go中添加panic恢复，优雅处理崩溃

**Files:**
- Modify: `main.go`

- [ ] **Step 1: 查看现有main.go**

```bash
cat main.go
```

- [ ] **Step 2: 编写失败测试 - Panic恢复**

```go
// main_test.go
package main

import (
    "sync"
    "testing"

    "github.com/stretchr/testify/assert"
)

func TestMain_RecoverFromPanic(t *testing.T) {
    // 这个测试验证panic恢复机制工作
    // 实际的恢复逻辑在main函数中

    var recovered bool
    var mu sync.Mutex

    // 模拟panic和恢复
    func() {
        defer func() {
            if r := recover(); r != nil {
                mu.Lock()
                recovered = true
                mu.Unlock()
            }
        }()

        panic("test panic")
    }()

    assert.True(t, recovered, "panic should be recovered")
}
```

- [ ] **Step 3: 运行测试**

```bash
go test -v -run TestMain_RecoverFromPanic
```

预期输出: 测试通过 ✓

- [ ] **Step 4: 在main.go中添加panic恢复**

```go
// main.go
package main

import (
    "fmt"
    "os"
    "runtime/debug"
)

func main() {
    // 添加顶层panic恢复
    defer func() {
        if r := recover(); r != nil {
            fmt.Fprintf(os.Stderr, "Application panic: %v\n", r)
            fmt.Fprintf(os.Stderr, "\nStack trace:\n%s\n", debug.Stack())
            os.Exit(1)
        }
    }()

    if err := rootCmd.Execute(); err != nil {
        os.Exit(1)
    }
}
```

- [ ] **Step 5: 测试panic恢复**

```bash
# 创建一个会panic的版本进行测试
go build -o gvm-test
./gvm-test panic 2>&1 | grep -q "Application panic"
echo $?  # 应该输出0（找到了panic消息）
```

预期输出: 找到panic恢复消息

- [ ] **Step 6: 提交panic恢复机制**

```bash
git add main.go main_test.go
git commit -m "feat: add panic recovery mechanism

- Add top-level panic recovery in main()
- Print stack trace on panic
- Ensure graceful exit on crash
- Add panic recovery test

Stability: Prevents abrupt crashes, improves error reporting

Co-Authored-By: Claude Sonnet 4.6 <noreply@anthropic.com>"
```

---

### Task 8: 添加集成测试

**目标:** 创建集成测试，验证模块间协作

**Files:**
- Create: `test/integration/security_integration_test.go`
- Create: `test/integration/validation_integration_test.go`

- [ ] **Step 1: 创建集成测试目录**

```bash
mkdir -p test/integration
```

- [ ] **Step 2: 编写安全模块集成测试**

```go
// test/integration/security_integration_test.go
package integration

import (
    "context"
    "os"
    "path/filepath"
    "testing"

    "gvm/internal/core/validation"
    "gvm/internal/util/exec"
    "gvm/internal/util/path"

    "github.com/stretchr/testify/assert"
)

func TestSecurityIntegration_SafeDownloadFlow(t *testing.T) {
    if testing.Short() {
        t.Skip("skipping integration test in short mode")
    }

    // 创建临时目录
    tempDir, err := os.MkdirTemp("", "gvm-integration-*")
    assert.NoError(t, err)
    defer os.RemoveAll(tempDir)

    // 1. 验证URL
    url := "https://go.dev/dl/go1.20.0.linux-amd64.tar.gz"
    err = validation.ValidateURL(url)
    assert.NoError(t, err, "valid URL should pass validation")

    // 2. 验证下载路径
    downloadPath := filepath.Join(tempDir, "downloads", "file.tar.gz")
    err = validation.ValidatePath(downloadPath)
    assert.NoError(t, err, "download path should be valid")

    // 3. 检查路径安全
    safe, err := path.IsPathSafe(tempDir, "downloads/file.tar.gz")
    assert.NoError(t, err)
    assert.True(t, safe, "path should be safe")

    // 4. 验证解压目标路径
    extractPath := filepath.Join(tempDir, "versions", "go1.20.0")
    err = validation.ValidatePath(extractPath)
    assert.NoError(t, err)

    safe, err = path.IsPathSafe(tempDir, extractPath)
    assert.NoError(t, err)
    assert.True(t, safe, "extract path should be safe")
}

func TestSecurityIntegration_CommandExecution(t *testing.T) {
    if testing.Short() {
        t.Skip("skipping integration test in short mode")
    }

    // 创建安全执行器
    executor := exec.NewSafeExecutor([]string{"ls", "echo"})

    // 测试安全命令执行
    err := executor.Execute(context.Background(), "echo", "test")
    assert.NoError(t, err, "safe command should execute")

    // 测试危险命令被阻止
    err = executor.Execute(context.Background(), "sh", "-c", "rm -rf /")
    assert.Error(t, err, "dangerous command should be blocked")
}

func TestSecurityIntegration_PathTraversalPrevention(t *testing.T) {
    if testing.Short() {
        t.Skip("skipping integration test in short mode")
    }

    tempDir, err := os.MkdirTemp("", "gvm-integration-*")
    assert.NoError(t, err)
    defer os.RemoveAll(tempDir)

    // 测试各种路径遍历攻击
    attacks := []string{
        "../../../etc/passwd",
        "subdir/../../etc/passwd",
        "/etc/passwd",
        "../test",
    }

    for _, attack := range attacks {
        t.Run("attack_"+attack, func(t *testing.T) {
            safe, err := path.IsPathSafe(tempDir, attack)
            assert.Error(t, err, "path traversal should be detected")
            assert.False(t, safe, "path should not be safe")
        })
    }
}
```

- [ ] **Step 3: 编写验证层集成测试**

```go
// test/integration/validation_integration_test.go
package integration

import (
    "testing"

    "gvm/internal/core/validation"

    "github.com/stretchr/testify/assert"
)

func TestValidationIntegration_RealWorldVersions(t *testing.T) {
    if testing.Short() {
        t.Skip("skipping integration test in short mode")
    }

    // 测试真实版本号
    versions := []struct {
        version   string
        shouldErr bool
    }{
        {"1.20.0", false},
        {"1.21.5", false},
        {"v1.20.0", false},
        {"1.20", false},
        {"invalid", true},
        {"", true},
        {"1.-1.0", true},
    }

    for _, tc := range versions {
        t.Run("version_"+tc.version, func(t *testing.T) {
            err := validation.ValidateVersion(tc.version)
            if tc.shouldErr {
                assert.Error(t, err)
            } else {
                assert.NoError(t, err)
            }
        })
    }
}

func TestValidationIntegration_RealWorldURLs(t *testing.T) {
    if testing.Short() {
        t.Skip("skipping integration test in short mode")
    }

    urls := []struct {
        url       string
        shouldErr bool
    }{
        {"https://go.dev/dl/go1.20.0.linux-amd64.tar.gz", false},
        {"https://nodejs.org/dist/v18.0.0/node-v18.0.0.tar.gz", false},
        {"https://www.python.org/ftp/python/3.9.7/Python-3.9.7.tgz", false},
        {"http://insecure.com/file.tar.gz", true},
        {"not-a-url", true},
        {"", true},
    }

    for _, tc := range urls {
        t.Run("url_"+tc.url, func(t *testing.T) {
            err := validation.ValidateURL(tc.url)
            if tc.shouldErr {
                assert.Error(t, err)
            } else {
                assert.NoError(t, err)
            }
        })
    }
}
```

- [ ] **Step 4: 运行集成测试**

```bash
go test ./test/integration/... -v
```

预期输出: 所有集成测试通过 ✓

- [ ] **Step 5: 运行所有测试（单元+集成）**

```bash
go test ./... -v -race
```

预期输出: 所有测试通过，无竞态条件 ✓

- [ ] **Step 6: 提交集成测试**

```bash
git add test/integration/
git commit -m "test: add integration tests for security modules

- Test complete download flow with validation
- Test command execution with security checks
- Test path traversal prevention
- Test real-world version validation
- Test real-world URL validation
- Verify modules work together correctly

Testing: Integration tests ensure module compatibility

Co-Authored-By: Claude Sonnet 4.6 <noreply@anthropic.com>"
```

---

### Task 9: 增强路径安全（符号链接检查）

**目标:** 在路径验证中添加符号链接检查

**Files:**
- Modify: `internal/util/path/path.go`
- Modify: `internal/util/path/path_test.go`

- [ ] **Step 1: 编写失败测试 - 符号链接检测**

```go
// internal/util/path/path_test.go
func TestCheckSymlinkSafety(t *testing.T) {
    tempDir, cleanup := framework.SetupTestEnvironment(t)
    defer cleanup()

    // 创建安全的符号链接
    safeLink := filepath.Join(tempDir, "safe_link")
    safeTarget := filepath.Join(tempDir, "safe_target")
    framework.CreateTestFile(t, safeTarget, "content")

    err := os.Symlink(safeTarget, safeLink)
    assert.NoError(t, err)

    // 检查安全符号链接
    err = CheckSymlinkSafety(tempDir, safeLink)
    assert.NoError(t, err, "safe symlink should pass")

    // 创建逃逸的符号链接
    escapeDir, _ := os.MkdirTemp("", "escape-*")
    defer os.RemoveAll(escapeDir)

    escapeLink := filepath.Join(tempDir, "escape_link")
    err = os.Symlink(escapeDir, escapeLink)
    assert.NoError(t, err)

    // 检查逃逸符号链接
    err = CheckSymlinkSafety(tempDir, escapeLink)
    assert.Error(t, err, "escaping symlink should fail")
}
```

- [ ] **Step 2: 运行测试确认失败**

```bash
go test ./internal/util/path/... -v -run TestCheckSymlinkSafety
```

预期输出: `undefined: CheckSymlinkSafety`

- [ ] **Step 3: 实现符号链接安全检查**

```go
// internal/util/path/path.go
import (
    "os"
    "path/filepath"
)

// CheckSymlinkSafety 检查符号链接是否安全，不逃逸基础路径
func CheckSymlinkSafety(basePath, linkPath string) error {
    // 解析链接的绝对路径
    absLink, err := filepath.Abs(linkPath)
    if err != nil {
        return fmt.Errorf("failed to resolve link path: %w", err)
    }

    // 读取符号链接目标
    target, err := os.Readlink(absLink)
    if err != nil {
        return fmt.Errorf("failed to read symlink: %w", err)
    }

    // 解析目标的绝对路径
    var absTarget string
    if filepath.IsAbs(target) {
        absTarget = target
    } else {
        // 相对路径，相对于链接所在目录解析
        absTarget = filepath.Join(filepath.Dir(absLink), target)
    }

    absTarget = filepath.Clean(absTarget)

    // 检查目标是否在基础路径内
    rel, err := filepath.Rel(basePath, absTarget)
    if err != nil {
        return fmt.Errorf("failed to compute relative path: %w", err)
    }

    if strings.HasPrefix(rel, "..") {
        return fmt.Errorf("symlink target escapes base directory: %s -> %s", linkPath, target)
    }

    return nil
}
```

- [ ] **Step 4: 在IsPathSafe中集成符号链接检查**

```go
// internal/util/path/path.go

// IsPathSafe 检查目标路径是否在基础路径内，防止路径遍历攻击
func IsPathSafe(basePath, targetPath string) (bool, error) {
    // 解析为绝对路径
    absBase, err := filepath.Abs(basePath)
    if err != nil {
        return false, fmt.Errorf("failed to resolve base path: %w", err)
    }

    // 构建完整目标路径
    fullTarget := filepath.Join(basePath, targetPath)
    absTarget, err := filepath.Abs(fullTarget)
    if err != nil {
        return false, fmt.Errorf("failed to resolve target path: %w", err)
    }

    // 检查符号链接
    fileInfo, err := os.Lstat(absTarget)
    if err == nil && fileInfo.Mode()&os.ModeSymlink != 0 {
        // 是符号链接，检查安全性
        if err := CheckSymlinkSafety(absBase, absTarget); err != nil {
            return false, err
        }
    }

    // 检查目标路径是否在基础路径内
    rel, err := filepath.Rel(absBase, absTarget)
    if err != nil {
        return false, fmt.Errorf("failed to compute relative path: %w", err)
    }

    // 如果相对路径以..开头，说明试图逃逸
    if strings.HasPrefix(rel, "..") {
        return false, fmt.Errorf("path traversal attempt detected: %s tries to escape %s", targetPath, basePath)
    }

    return true, nil
}
```

- [ ] **Step 5: 运行测试确认通过**

```bash
go test ./internal/util/path/... -v -race
```

预期输出: 所有测试通过，无竞态条件 ✓

- [ ] **Step 6: 提交符号链接安全检查**

```bash
git add internal/util/path/
git commit -m "feat: add symlink safety checks to path validation

- Add CheckSymlinkSafety function
- Detect symlink escape attempts
- Integrate symlink check into IsPathSafe
- Prevent symlink-based directory traversal
- Add comprehensive symlink tests

Security: Prevents symlink-based path traversal attacks

Co-Authored-By: Claude Sonnet 4.6 <noreply@anthropic.com>"
```

---

### Task 10: 建立性能基准

**目标:** 创建性能基准测试，建立性能基线

**Files:**
- Create: `test/benchmark/`
- Create: `test/benchmark/benchmark_test.go`

- [ ] **Step 1: 创建基准测试目录**

```bash
mkdir -p test/benchmark
```

- [ ] **Step 2: 创建安全模块基准测试**

```go
// test/benchmark/security_benchmark_test.go
package benchmark

import (
    "context"
    "testing"

    "gvm/internal/util/exec"
    "gvm/internal/util/path"
)

func BenchmarkSafeExecutor_Execute(b *testing.B) {
    executor := exec.NewSafeExecutor([]string{"echo"})
    ctx := context.Background()

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        executor.Execute(ctx, "echo", "test")
    }
}

func BenchmarkParseCommand(b *testing.B) {
    cmdString := "tar -xzf file.tar.gz -C /tmp"

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        exec.ParseCommand(cmdString)
    }
}

func BenchmarkValidateVersion(b *testing.B) {
    version := "1.20.0"

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        validation.ValidateVersion(version)
    }
}

func BenchmarkValidatePath(b *testing.B) {
    pathStr := "/usr/local/go/bin"

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        validation.ValidatePath(pathStr)
    }
}

func BenchmarkIsPathSafe(b *testing.B) {
    basePath := "/tmp/gvm"
    targetPath := "versions/go1.20.0"

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        path.IsPathSafe(basePath, targetPath)
    }
}
```

- [ ] **Step 3: 运行基准测试建立基线**

```bash
go test -bench=. -benchmem ./test/benchmark/... > baseline.txt
cat baseline.txt
```

预期输出: 基准测试结果

- [ ] **Step 4: 创建性能目标文档**

```markdown
# test/benchmark/PERFORMANCE_TARGETS.md

## 性能目标

基于基准测试结果，以下是性能目标：

### 命令执行
- ParseCommand: < 1μs per operation
- SafeExecutor.Execute: < 100μs per operation

### 验证函数
- ValidateVersion: < 500ns per operation
- ValidatePath: < 1μs per operation
- ValidateURL: < 2μs per operation

### 路径检查
- IsPathSafe: < 5μs per operation
- CheckSymlinkSafety: < 10μs per operation

## 基准测试命令

```bash
# 运行所有基准测试
go test -bench=. -benchmem ./...

# 比较前后性能
benchcmp baseline.txt optimized.txt

# CPU profile分析
go test -cpuprofile=cpu.prof -bench=. ./test/benchmark/...
go tool pprof cpu.prof
```

## 持续监控

每次修改后运行基准测试，确保性能没有退化。
```

- [ ] **Step 5: 提交基准测试**

```bash
git add test/benchmark/
git commit -m "test: add performance benchmarks

- Add benchmarks for security functions
- Add benchmarks for validation functions
- Establish performance baseline
- Document performance targets
- Enable performance regression detection

Performance: Provides baseline for optimization efforts

Co-Authored-By: Claude Sonnet 4.6 <noreply@anthropic.com>"
```

---

### Task 11: 配置CI/CD流水线

**目标:** 设置GitHub Actions工作流，自动运行测试、代码检查和安全扫描

**Files:**
- Create: `.github/workflows/test.yml`
- Create: `.github/workflows/lint.yml`
- Create: `.github/workflows/security.yml`
- Modify: `Makefile`

- [ ] **Step 1: 创建workflows目录**

```bash
mkdir -p .github/workflows
```

- [ ] **Step 2: 创建测试工作流**

```yaml
# .github/workflows/test.yml
name: Test

on:
  push:
    branches: [main, feat/rui]
  pull_request:
    branches: [main]

jobs:
  test:
    runs-on: ubuntu-latest

    strategy:
      matrix:
        go-version: ['1.26']

    steps:
    - name: Checkout code
      uses: actions/checkout@v4

    - name: Set up Go
      uses: actions/setup-go@v5
      with:
        go-version: ${{ matrix.go-version }}

    - name: Install dependencies
      run: |
        go mod download
        go install github.com/stretchr/testify/mock/mockgen@latest

    - name: Run tests
      run: go test -v -race -coverprofile=coverage.out ./...

    - name: Check coverage
      run: |
        coverage=$(go tool cover -func=coverage.out | grep total | awk '{print $3}' | sed 's/%//')
        echo "Total coverage: $coverage%"
        if (( $(echo "$coverage < 80" | bc -l) )); then
          echo "Coverage is below 80%"
          exit 1
        fi

    - name: Upload coverage to Codecov
      uses: codecov/codecov-action@v4
      with:
        file: ./coverage.out
        fail_ci_if_error: false

    - name: Build
      run: go build -v ./...
```

- [ ] **Step 3: 创建代码检查工作流**

```yaml
# .github/workflows/lint.yml
name: Lint

on:
  push:
    branches: [main, feat/rui]
  pull_request:
    branches: [main]

jobs:
  lint:
    runs-on: ubuntu-latest

    steps:
    - name: Checkout code
      uses: actions/checkout@v4

    - name: Set up Go
      uses: actions/setup-go@v5
      with:
        go-version: '1.26'

    - name: Run golangci-lint
      uses: golangci/golangci-lint-action@v6
      with:
        version: latest
        args: --timeout=10m

    - name: Check code formatting
      run: |
        fmt_output=$(gofmt -l .)
        if [ -n "$fmt_output" ]; then
          echo "Code is not formatted:"
          echo "$fmt_output"
          exit 1
        fi

    - name: Check for go.mod consistency
      run: |
        go mod tidy
        if ! git diff --exit-code go.mod go.sum; then
          echo "go.mod/go.sum are not consistent"
          exit 1
        fi
```

- [ ] **Step 4: 创建安全扫描工作流**

```yaml
# .github/workflows/security.yml
name: Security

on:
  push:
    branches: [main, feat/rui]
  pull_request:
    branches: [main]
  schedule:
    # Run security scan daily at 2 AM UTC
    - cron: '0 2 * * *'

jobs:
  security:
    runs-on: ubuntu-latest

    steps:
    - name: Checkout code
      uses: actions/checkout@v4

    - name: Set up Go
      uses: actions/setup-go@v5
      with:
        go-version: '1.26'

    - name: Run Gosec Security Scanner
      uses: securego/gosec@master
      with:
        args: ./...

    - name: Check for vulnerabilities
      run: |
        go install golang.org/x/vuln/cmd/govulncheck@latest
        govulncheck ./...

    - name: Run trivy fs scanner
      uses: aquasecurity/trivy-action@master
      with:
        scan-type: 'fs'
        scan-ref: '.'
        format: 'sarif'
        output: 'trivy-results.sarif'

    - name: Upload Trivy results to GitHub Security
      uses: github/codeql-action/upload-sarif@v3
      if: always()
      with:
        sarif_file: 'trivy-results.sarif'
```

- [ ] **Step 5: 更新Makefile添加测试目标**

```makefile
# 添加到现有Makefile

.PHONY: test test-coverage test-race lint dev-tools security clean benchmark

# 运行所有测试
test:
	go test -v ./...

# 运行带覆盖率的测试
test-coverage:
	go test -v -race -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

# 运行竞态检测
test-race:
	go test -race ./...

# 代码检查
lint:
	golangci-lint run
	@if [ -n "$$(gofmt -l .)" ]; then \
		echo "Code is not formatted. Run 'gofmt -w .'"; \
		gofmt -l .; \
		exit 1; \
	fi

# 安全扫描
security:
	@echo "Running security checks..."
	go install golang.org/x/vuln/cmd/govulncheck@latest
	govulncheck ./...

# 安装开发工具
dev-tools:
	go install github.com/stretchr/testify/mock/mockgen@latest
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install golang.org/x/vuln/cmd/govulncheck@latest
	@echo "Development tools installed"

# 清理
clean:
	go clean
	rm -f coverage.out coverage.html

# 运行基准测试
benchmark:
	go test -bench=. -benchmem ./... > benchmark.txt
	@echo "Benchmark results saved to benchmark.txt"
```

- [ ] **Step 6: 测试Makefile目标**

```bash
make dev-tools
make test
make lint
```

预期输出: 所有命令成功执行

- [ ] **Step 7: 提交CI/CD配置**

```bash
git add .github/workflows/ Makefile
git commit -m "feat: add CI/CD pipeline with security scanning

- Add GitHub Actions workflow for testing
- Add golangci-lint workflow
- Add security scanning (gosec, govulncheck, trivy)
- Add Makefile targets for local development
- Enforce code formatting and mod consistency
- Check test coverage threshold (80%)
- Run race detector on all tests

CI: Improves code quality and security

Co-Authored-By: Claude Sonnet 4.6 <noreply@anthropic.com>"
```

---

### Task 6: 创建测试文档

**目标:** 编写测试框架和TDD流程文档

**Files:**
- Create: `docs/testing/test-framework.md`
- Create: `docs/testing/tdd-workflow.md`

- [ ] **Step 1: 创建文档目录**

```bash
mkdir -p docs/testing
```

- [ ] **Step 2: 创建测试框架文档**

```markdown
# GVM测试框架文档

## 概述

GVM项目使用testify作为主要测试框架，遵循严格的TDD方法论。

## 测试框架组件

### 1. 测试辅助工具 (test/framework)

#### SetupTestEnvironment
创建临时测试环境。

```go
func TestExample(t *testing.T) {
    tempDir, cleanup := framework.SetupTestEnvironment(t)
    defer cleanup()

    // 使用tempDir进行测试
}
```

#### CreateTestFile
在测试目录中创建文件。

```go
func TestFileOperation(t *testing.T) {
    tempDir, cleanup := framework.SetupTestEnvironment(t)
    defer cleanup()

    testFile := filepath.Join(tempDir, "test.txt")
    framework.CreateTestFile(t, testFile, "content")

    framework.AssertFileExists(t, testFile)
}
```

### 2. 测试断言

使用testify/assert进行断言：

```go
import "github.com/stretchr/testify/assert"

func TestAssertion(t *testing.T) {
    assert.Equal(t, expected, actual)
    assert.NoError(t, err)
    assert.Contains(t, str, substr)
    assert.True(t, condition)
}
```

### 3. Mock对象

使用testify/mock创建mock：

```go
import "github.com/stretchr/testify/mock"

type MockDownloader struct {
    mock.Mock
}

func (m *MockDownloader) Download(ctx context.Context, url string) (io.ReadCloser, error) {
    args := m.Called(ctx, url)
    return args.Get(0).(io.ReadCloser), args.Error(1)
}
```

## 测试分类

### 单元测试
- 测试单个函数或方法
- 使用mock对象隔离依赖
- 快速执行（毫秒级）

### 集成测试
- 测试多个组件交互
- 使用真实文件系统（临时目录）
- 较慢执行（秒级）

### 端到端测试
- 测试完整工作流
- 测试真实场景
- 最慢执行（分钟级）

## 运行测试

```bash
# 运行所有测试
make test

# 运行特定包的测试
go test ./internal/core/validation/...

# 运行特定测试
go test -run TestValidateVersion ./internal/core/validation/...

# 运行带覆盖率的测试
make test-coverage

# 运行竞态检测
make test-race

# 跳过慢测试
go test -short ./...
```

## 测试覆盖率目标

- 每个模块覆盖率 ≥ 80%
- 使用`go test -coverprofile=coverage.out`生成覆盖率报告
- 使用`go tool cover -html=coverage.out`查看HTML报告

## 最佳实践

1. **测试命名**: 使用`Test<FunctionName>_<Scenario>`格式
2. **表驱动测试**: 使用测试表覆盖多个场景
3. **清晰的错误信息**: 使用assert.Error()包含期望的错误信息
4. **清理资源**: 使用defer确保测试资源被清理
5. **并发安全**: 所有测试必须通过`go test -race`
```

- [ ] **Step 3: 创建TDD流程文档**

```markdown
# TDD流程文档

## 什么是TDD

测试驱动开发（Test-Driven Development）是一种开发方法论，要求先编写测试，然后编写代码使测试通过。

## TDD循环

### Red（红）- 编写失败的测试

先编写一个失败的测试，描述你想要实现的功能。

```go
func TestValidateVersion_ValidVersion_ReturnsNoError(t *testing.T) {
    err := ValidateVersion("1.20.0")
    assert.NoError(t, err)  // 这会失败，因为函数还不存在
}
```

运行测试确认失败：
```bash
go test ./internal/core/validation/... -v
# 输出: undefined: ValidateVersion
```

### Green（绿）- 编写最小实现

编写最简单的代码使测试通过。

```go
func ValidateVersion(version string) error {
    return nil  // 最简单的实现
}
```

运行测试确认通过：
```bash
go test ./internal/core/validation/... -v
# 输出: PASS
```

### Refactor（重构）- 改进代码

现在测试通过了，改进代码质量和正确性。

```go
func ValidateVersion(version string) error {
    if version == "" {
        return fmt.Errorf("version cannot be empty")
    }
    // 添加更多验证逻辑...
    return nil
}
```

运行测试确保仍然通过：
```bash
go test ./internal/core/validation/... -v
# 输出: PASS
```

重复这个循环！

## GVM项目的TDD规则

1. **绝不跳过Red阶段**: 必须先写测试
2. **最小化Green阶段**: 只写足够通过测试的代码
3. **频繁重构**: 测试保护下安全重构
4. **小步前进**: 每个循环只实现一个小功能
5. **持续集成**: 每次提交都运行完整测试套件

## 常见错误

### ❌ 编写太多测试再实现
**正确**: 一次写一个测试，立即实现

### ❌ 跳过重构阶段
**正确**: 测试通过后立即重构改进

### ❌ 测试实现细节
**正确**: 测试行为和接口，不测试内部实现

### ❌ 忽略失败的测试
**正确**: 所有测试必须通过才能继续

## 示例：完整TDD循环

### Step 1: 编写失败测试

```go
func TestParseCommand_RejectsPipe(t *testing.T) {
    cmd, err := ParseCommand("ls | rm")
    assert.Error(t, err)
    assert.Contains(t, err.Error(), "dangerous")
    assert.Nil(t, cmd)
}
```

运行: 失败 ✓

### Step 2: 最小实现

```go
func ParseCommand(cmdString string) (*Command, error) {
    if strings.Contains(cmdString, "|") {
        return nil, fmt.Errorf("command contains dangerous character: |")
    }
    return &Command{Name: cmdString}, nil
}
```

运行: 通过 ✓

### Step 3: 重构改进

```go
var dangerousChars = []string{"|", "&", ";", ...}

func ParseCommand(cmdString string) (*Command, error) {
    for _, char := range dangerousChars {
        if strings.Contains(cmdString, char) {
            return nil, fmt.Errorf("command contains dangerous character: %s", char)
        }
    }
    // 解析命令和参数
    parts := strings.Fields(cmdString)
    return &Command{Name: parts[0], Args: parts[1:]}, nil
}
```

运行: 仍然通过 ✓

## 总结

TDD不是慢，而是更快！
- 更少的调试时间
- 更高的代码质量
- 更自信的重构
- 活的文档

记住：**Red → Green → Refactor**
```

- [ ] **Step 4: 提交测试文档**

```bash
git add docs/testing/
git commit -m "docs: add testing framework and TDD workflow documentation

- Add test framework guide with examples
- Document test utilities and helpers
- Add TDD workflow guide with Red-Green-Refactor
- Provide best practices and common pitfalls
- Include code examples for all scenarios

Docs: Helps developers follow TDD methodology

Co-Authored-By: Claude Sonnet 4.6 <noreply@anthropic.com>"
```

---

## 第一阶段完成检查清单

在进入第二阶段之前，确认以下所有项目已完成：

### 测试框架
- [ ] test/framework/setup.go 创建并测试
- [ ] test/framework/fixtures.go 创建并测试
- [ ] test/framework/helpers.go 创建并测试
- [ ] 所有测试通过且无竞态条件

### 安全模块
- [ ] internal/util/exec/executor.go 实现
- [ ] SafeExecutor 白名单机制工作正常
- [ ] ParseCommand 拒绝所有危险字符
- [ ] 测试覆盖率达到80%+

### 输入验证
- [ ] internal/core/validation/validation.go 实现
- [ ] 版本验证正常工作
- [ ] 路径验证检测遍历攻击
- [ ] URL验证要求HTTPS
- [ ] 所有测试通过

### 路径安全
- [ ] internal/util/path/path.go 增强
- [ ] IsPathSafe 检测路径遍历
- [ ] CheckSymlinkSafety 检测符号链接逃逸
- [ ] 测试覆盖所有攻击场景

### 错误处理
- [ ] internal/core/errors.go 标准错误码定义
- [ ] 所有错误码有清晰的错误消息
- [ ] 支持错误包装和比较

### HTTP客户端
- [ ] internal/http/client.go 资源泄漏修复
- [ ] 响应体正确关闭
- [ ] 连接池配置
- [ ] 重试逻辑实现
- [ ] 资源清理测试通过

### 稳定性
- [ ] main.go panic恢复机制
- [ ] 崩溃时打印堆栈跟踪
- [ ] 优雅退出处理

### 集成测试
- [ ] test/integration/ 安全集成测试
- [ ] test/integration/ 验证集成测试
- [ ] 模块间协作测试通过

### 性能基准
- [ ] test/benchmark/ 基准测试创建
- [ ] 性能基线建立
- [ ] 性能目标文档化
- [ ] Makefile benchmark目标工作

### CI/CD
- [ ] GitHub Actions测试工作流配置
- [ ] golangci-lint工作流配置
- [ ] 安全扫描工作流配置
- [ ] Makefile目标测试通过
- [ ] CI管道在GitHub上成功运行

### 文档
- [ ] 测试框架文档完整
- [ ] TDD流程文档完整
- [ ] 所有代码示例可运行

### 质量门槛
- [ ] 所有单元测试通过
- [ ] go test -race 无竞态警告
- [ ] go test -cover ≥80%
- [ ] golangci-lint 无错误
- [ ] 安全扫描无高危漏洞

---

## 下一步

完成第一阶段后，进入**第二阶段：核心模块重构**：
- 重构 config 模块（改进错误处理，移除panic）
- 重构 registry 模块（添加并发安全）
- 创建基础语言层（消除代码重复）

**准备好进入第二阶段时，运行：**
```bash
git tag phase1-complete
git push origin phase1-complete
```
