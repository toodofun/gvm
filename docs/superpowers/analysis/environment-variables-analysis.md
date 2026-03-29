# 环境变量设置分析报告

**日期:** 2026-03-29
**目标:** 检查各语言实现中的环境变量设置，识别遗漏

## 当前实现分析

### 1. Golang (languages/golang/golang.go:108-131)

**已设置:**
- ✅ GOROOT - Go安装目录
- ✅ GOPATH - Go工作目录
- ✅ PATH - 追加 $GOROOT/bin 和 $GOPATH/bin

**遗漏:**
- ❌ GO111MODULE - Go模块支持 (应设为 `on` 或 `auto`)
- ❌ GOPROXY - Go模块代理 (中国用户需要: `https://goproxy.cn,direct`)
- ❌ GOSUMDB - Go校验和数据库 (应设为 `sum.golang.google.cn` 或 `off`)
- ❌ GOTOOLCHAIN - Go工具链版本管理 (Go 1.21+)
- ⚠️ GOBIN - 可执行文件安装目录 (可选，但推荐)
- ❌ CGO_ENABLED - CGO支持 (某些场景需要)

### 2. Python (languages/python/python.go:229-246)

**已设置:**
- ✅ PATH - 追加 $PYTHONHOME/bin
- ✅ LD_LIBRARY_PATH - 追加 lib目录 (Linux/Unix共享库)

**遗漏:**
- ❌ PYTHONHOME - **被注释掉了！** (这是Python的关键环境变量)
- ❌ PYTHONPATH - Python模块搜索路径
- ❌ VIRTUAL_ENV - 虚拟环境标识 (如果使用虚拟环境)
- ❌ PIP_CONFIG_FILE - pip配置文件路径
- ❌ PIP_INDEX_URL - pip包索引 (中国用户需要: `https://pypi.tuna.tsinghua.edu.cn/simple`)
- ⚠️ PYTHONDONTWRITEBYTECODE - 禁止生成.pyc文件 (推荐)
- ❌ PYTHONUNBUFFERED - 非缓冲输出 (推荐用于开发)
- ❌ PYTHONIOENCODING - Python I/O编码 (推荐: `utf-8`)

**平台特定:**
- ❌ Windows: PYTHONPATH 可能需要使用分号分隔
- ❌ macOS: 没有设置 Frameworks 相关路径

### 3. Node.js (languages/node/node.go:123-135)

**已设置:**
- ✅ PATH - 追加 node bin目录

**遗漏:**
- ❌ NODE_PATH - Node模块搜索路径
- ⚠️ npm配置没有通过环境变量设置
- ❌ NPM_CONFIG_PREFIX - npm全局安装路径
- ❌ NPM_CONFIG_REGISTRY - npm注册表 (中国用户需要)
- ❌ NODE_ENV - 运行环境 (production/development)
- ❌ YARN_GLOBAL_FOLDER (如果使用yarn)
- ❌ COREPACK_ENABLE_DOWNLOAD_PROMPT (Node 16.9+)

**npm配置文件 (推荐但不是环境变量):**
- 建议创建 .npmrc 文件设置:
  - prefix = ${GVM_ROOT}/nodejs/current
  - registry = https://registry.npmmirror.com

### 4. Java (languages/java/java.go:71-79)

**已设置:**
- ✅ PATH - 追加 $JAVA_HOME/bin

**遗漏:**
- ❌ JAVA_HOME - **关键！** Java安装目录根本没设置！
- ❌ CLASSPATH - Java类路径
- ❌ JAVA_TOOL_OPTIONS - JVM参数
- ❌ _JAVA_OPTIONS - JVM选项 (某些JVM使用)
- ❌ JDK_HOME - 某些工具使用 (替代JAVA_HOME)

**平台特定:**
- ❌ macOS: 可能需要设置 Info.plist 中的 JVMVersion
- ❌ Windows: 可能需要设置 CLASSPATH 使用分号

### 5. Ruby (languages/ruby/ruby.go)

**需要检查:** 未找到 SetDefaultVersion 方法

**建议设置:**
- ❌ RUBY_HOME (或 RUBY_ROOT)
- ❌ GEM_HOME - gem安装目录
- ❌ GEM_PATH - gem搜索路径
- ❌ PATH - 追加 $GEM_HOME/bin

### 6. Rust (languages/rust/rust.go)

**需要检查:** 未找到 SetDefaultVersion 方法

**建议设置:**
- ❌ RUSTUP_HOME - rustup安装目录
- ❌ CARGO_HOME - cargo主目录
- ❌ PATH - 追加 $CARGO_HOME/bin
- ❌ RUSTFLAGS - Rust编译器标志

### 7. GVM (languages/gvm/gvm.go)

**需要检查:** 未找到 SetDefaultVersion 方法

**建议设置:**
- ❌ GVM_ROOT - GVM根目录
- ❌ GVM_VERSIONS - 版本安装目录

## 关键问题汇总

### 🔴 严重问题 (必须修复)

1. **Java缺少JAVA_HOME** - 这会导致很多Java工具无法工作
2. **Python的PYTHONHOME被注释** - 这会导致Python无法找到标准库
3. **Go缺少模块代理设置** - 中国用户无法下载go模块

### 🟡 重要问题 (强烈推荐)

4. **Node.js缺少npm配置** - npm全局包管理混乱
5. **Python缺少PYTHONPATH** - 自定义模块无法导入
6. **Go缺少GO111MODULE** - 影响模块行为
7. **所有语言缺少平台特定处理**

### 🟢 建议改进

8. **缺少中国镜像配置** - 影响国内用户体验
9. **缺少编码相关设置** - 可能导致中文环境问题
10. **缺少性能优化设置** - 如 PYTHONDONTWRITEBYTECODE

## 推荐的修复策略

### 阶段1: 修复关键问题 (立即)

1. Java: 添加 JAVA_HOME
2. Python: 取消注释 PYTHONHOME
3. Go: 添加 GO111MODULE=on

### 阶段2: 添加平台支持 (短期)

4. 检测操作系统并设置平台特定环境变量
5. 处理Windows路径分隔符问题

### 阶段3: 优化用户体验 (中期)

6. 添加中国镜像配置支持
7. 添加配置文件支持 (.gvm/config.yml)
8. 允许用户自定义环境变量

## 实施建议

### 选项 A: 立即修复 (推荐)
在第一阶段实施计划中添加新任务，修复所有关键问题

### 选项 B: 分阶段修复
- 第一阶段: 修复严重问题
- 第五阶段 (语言重构): 完整重构所有语言的环境变量设置

### 选项 C: 配置驱动
创建统一的配置系统，让用户自定义环境变量

## 测试建议

为每个语言添加环境变量测试:

```go
func TestGolang_SetDefaultVersion_SetsCorrectEnv(t *testing.T) {
    // 设置默认版本
    g := &Golang{}
    err := g.SetDefaultVersion(ctx, "1.20.0")

    // 验证环境变量
    assert.NoError(t, err)
    // 验证 GOROOT, GOPATH, GO111MODULE 等已正确设置
}
```

## 优先级矩阵

| 语言 | 严重度 | 用户影响 | 修复难度 | 优先级 |
|------|--------|----------|----------|--------|
| Java (JAVA_HOME) | 🔴 高 | 🔴 高 | 🟢 低 | P0 |
| Python (PYTHONHOME) | 🔴 高 | 🔴 高 | 🟢 低 | P0 |
| Go (GO111MODULE) | 🟡 中 | 🟡 中 | 🟢 低 | P1 |
| Node.js (npm配置) | 🟡 中 | 🟡 中 | 🟡 中 | P1 |
| Python (镜像) | 🟢 低 | 🟡 中 | 🟢 低 | P2 |
| Go (镜像) | 🟢 低 | 🟡 中 | 🟢 低 | P2 |

## 结论

当前实现存在**多个关键环境变量遗漏**，建议：
1. **立即修复** P0 问题 (JAVA_HOME, PYTHONHOME)
2. 在语言重构阶段（第五阶段）**全面重构**环境变量设置
3. 创建**统一配置系统**以提高可维护性
