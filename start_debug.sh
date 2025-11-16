#!/bin/bash

echo "🔍 检查环境..."
echo ""

# 检查 Go 版本
echo "Go 版本:"
go version
echo ""

# 检查端口占用
echo "检查端口 8080:"
lsof -ti:8080 && echo "⚠️  端口 8080 被占用" || echo "✅ 端口 8080 可用"
echo ""

# 清理端口
echo "清理端口..."
lsof -ti:8080 | xargs kill -9 2>/dev/null
echo ""

# 启动后端
echo "🚀 启动后端服务..."
echo "================================================"
go run main.go
