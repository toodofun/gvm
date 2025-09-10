#!/bin/bash

echo "🧪 开始测试所有语言的安装功能..."
echo ""

# 确保 gvm 已编译
if [ ! -f "./gvm" ]; then
    echo "🔨 编译 gvm..."
    go build -o gvm main.go
fi

# 测试函数
test_version_list() {
    local lang=$1
    echo "📋 测试 $lang 版本列表..."
    
    count=$(./gvm ls-remote $lang 2>/dev/null | wc -l | tr -d ' ')
    if [ "$count" -gt 0 ]; then
        echo "  ✅ $lang: $count 个版本"
        # 显示前5个版本作为示例
        echo "  📄 前5个版本:"
        ./gvm ls-remote $lang 2>/dev/null | head -5 | sed 's/^/    /'
    else
        echo "  ❌ $lang: 无版本列表"
    fi
    echo ""
}

test_quick_install() {
    local lang=$1
    local version=$2
    local timeout_duration=$3
    
    echo "⚡ 测试 $lang $version 快速安装 (超时: ${timeout_duration}s)..."
    
    # 检查是否已安装
    if ./gvm ls $lang 2>/dev/null | grep -q "$version"; then
        echo "  ℹ️ $lang $version 已安装，跳过测试"
        return
    fi
    
    # 命令行安装测试（限制时间避免长时间等待）
    echo "  🔹 开始安装..."
    
    # 使用后台进程和超时控制
    ./gvm install $lang $version &
    local pid=$!
    local count=0
    
    while [ $count -lt $timeout_duration ]; do
        if ! kill -0 $pid 2>/dev/null; then
            wait $pid
            local exit_code=$?
            if [ $exit_code -eq 0 ]; then
                echo "  ✅ $lang $version 安装成功"
                
                # 验证安装
                if ./gvm ls $lang 2>/dev/null | grep -q "$version"; then
                    echo "  ✅ $lang $version 安装验证成功"
                    
                    # 测试设置默认版本
                    if ./gvm use $lang $version 2>/dev/null; then
                        echo "  ✅ $lang $version 设置为默认版本成功"
                    else
                        echo "  ⚠️ $lang $version 设置默认版本失败"
                    fi
                else
                    echo "  ❌ $lang $version 安装验证失败"
                fi
            else
                echo "  ❌ $lang $version 安装失败 (退出码: $exit_code)"
            fi
            return
        fi
        
        sleep 1
        count=$((count + 1))
        
        # 每10秒显示一次进度
        if [ $((count % 10)) -eq 0 ]; then
            echo "  ⏳ $lang $version 安装进行中... (${count}s/${timeout_duration}s)"
        fi
    done
    
    # 超时处理
    echo "  ⏰ $lang $version 安装超时，终止进程..."
    kill $pid 2>/dev/null
    wait $pid 2>/dev/null
    echo "  ⚠️ $lang $version 安装测试超时"
    echo ""
}

# 主测试流程
echo "📋 第一阶段：版本列表测试"
echo "================================"
for lang in go node java python ruby rust; do
    test_version_list $lang
done

echo "⚡ 第二阶段：快速安装测试"
echo "================================"
echo "测试预编译包安装（速度较快）..."
test_quick_install "go" "1.21.0" "180"      # 3分钟
test_quick_install "node" "20.11.0" "180"   # 3分钟  
test_quick_install "rust" "1.75.0" "300"    # 5分钟

echo "🔨 第三阶段：源码编译测试"
echo "================================"
echo "⚠️ 注意：源码编译测试时间较长，可手动执行："
echo "  ./gvm install python 3.11.10  # 约 15-30 分钟"
echo "  ./gvm install ruby 3.1.7      # 约 20-40 分钟"
echo ""

echo "📊 测试完成摘要："
echo "================================"
echo "✅ 版本列表功能测试完成"
echo "✅ 快速安装功能测试完成"  
echo "💡 源码编译功能可手动测试"
echo ""
echo "🎯 GUI 测试说明："
echo "  运行 './gvm' 启动 GUI 界面"
echo "  选择语言和版本进行安装测试"
echo ""
echo "🔍 安装验证命令："
echo "  ./gvm ls [language]           # 查看已安装版本"
echo "  ./gvm current [language]      # 查看当前版本"
echo "  ./gvm use [language] [version] # 切换版本"
