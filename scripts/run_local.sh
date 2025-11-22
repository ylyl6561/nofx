#!/bin/bash

# ============================================================
# NoFX 本地测试环境启动脚本
# ============================================================

set -e

echo "🚀 启动 NoFX (本地测试环境)"
echo "======================================"

# 获取脚本所在目录的父目录（项目根目录）
PROJECT_ROOT="$( cd "$( dirname "${BASH_SOURCE[0]}" )/.." && pwd )"

# 设置环境标识
export APP_ENV=local

# 加载本地环境变量
if [ -f "$PROJECT_ROOT/.env.local" ]; then
    echo "📋 加载本地环境配置..."
    export $(cat "$PROJECT_ROOT/.env.local" | grep -v '^#' | xargs)
    echo "✅ 已加载 .env.local"
else
    echo "⚠️  警告: 未找到 .env.local 文件"
    echo "请先创建 .env.local 文件并配置本地数据库连接"
    exit 1
fi

echo ""
echo "📊 当前配置:"
echo "  数据库: $DATABASE_URL"
echo "  端口: ${API_SERVER_PORT:-8080}"
echo ""

# 检查本地数据库连接
echo "🔍 检查数据库连接..."
if psql "$DATABASE_URL" -c "SELECT 1;" > /dev/null 2>&1; then
    echo "✅ 数据库连接成功"
else
    echo "❌ 数据库连接失败"
    echo ""
    echo "请确保:"
    echo "  1. PostgreSQL 服务已启动"
    echo "  2. 数据库 'nofx_test' 已创建"
    echo "  3. .env.local 中的连接信息正确"
    echo ""
    echo "快速创建数据库:"
    echo "  createdb nofx_test"
    exit 1
fi

echo ""
echo "🏗️  初始化数据库表..."
psql "$DATABASE_URL" -f "$PROJECT_ROOT/scripts/init_database.sql" > /dev/null 2>&1
echo "✅ 数据库表已就绪"

echo ""
echo "======================================"
echo "🎯 启动应用..."
echo "======================================"
echo ""

# 切换到项目根目录
cd "$PROJECT_ROOT"

# 启动应用
go run main.go
