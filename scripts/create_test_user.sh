#!/bin/bash

# ============================================================
# 创建本地测试用户
# ============================================================

set -e

echo "👤 创建本地测试用户"
echo "======================================"

# 获取项目根目录
PROJECT_ROOT="$( cd "$( dirname "${BASH_SOURCE[0]}" )/.." && pwd )"

# 加载本地环境变量
if [ -f "$PROJECT_ROOT/.env.local" ]; then
    export $(cat "$PROJECT_ROOT/.env.local" | grep DATABASE_URL | xargs)
else
    echo "❌ 错误: 未找到 .env.local 文件"
    exit 1
fi

# 去除引号
DATABASE_URL=$(echo $DATABASE_URL | tr -d "'\"")

echo "📋 数据库: ${DATABASE_URL:0:50}..."
echo ""

# 执行SQL脚本
echo "🔧 创建测试用户..."
psql "$DATABASE_URL" -f "$PROJECT_ROOT/scripts/create_test_user.sql"

echo ""
echo "======================================"
echo "✅ 测试用户创建完成！"
echo "======================================"
echo ""
echo "📊 登录信息:"
echo "  邮箱: test@local.dev"
echo "  密码: test123"
echo ""
echo "💡 现在可以使用这个账号登录前端了"
echo ""
