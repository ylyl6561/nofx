#!/bin/bash

# ============================================================
# NoFX 生产环境启动脚本
# ============================================================

set -e

echo "🚀 启动 NoFX (生产环境)"
echo "======================================"

# 获取脚本所在目录的父目录（项目根目录）
PROJECT_ROOT="$( cd "$( dirname "${BASH_SOURCE[0]}" )/.." && pwd )"

# 设置环境标识
export APP_ENV=production

# 加载生产环境变量
if [ -f "$PROJECT_ROOT/.env.production" ]; then
    echo "📋 加载生产环境配置..."
    export $(cat "$PROJECT_ROOT/.env.production" | grep -v '^#' | xargs)
    echo "✅ 已加载 .env.production"
else
    echo "⚠️  警告: 未找到 .env.production 文件"
    echo "请先创建 .env.production 文件并配置生产数据库连接"
    exit 1
fi

echo ""
echo "📊 当前配置:"
echo "  数据库: ${DATABASE_URL:0:30}..."
echo "  端口: ${API_SERVER_PORT:-8080}"
echo ""

# 检查生产数据库连接
echo "🔍 检查数据库连接..."
if psql "$DATABASE_URL" -c "SELECT 1;" > /dev/null 2>&1; then
    echo "✅ 数据库连接成功"
else
    echo "❌ 数据库连接失败"
    echo "请检查 .env.production 中的连接信息"
    exit 1
fi

echo ""
echo "======================================"
echo "⚠️  生产环境模式"
echo "======================================"
echo ""

# 切换到项目根目录
cd "$PROJECT_ROOT"

# 启动应用
go run main.go
