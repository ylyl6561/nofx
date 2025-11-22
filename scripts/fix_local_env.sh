#!/bin/bash

# ============================================================
# 修复本地环境配置
# ============================================================

set -e

echo "🔧 修复本地环境配置"
echo "======================================"

# 获取项目根目录
PROJECT_ROOT="$( cd "$( dirname "${BASH_SOURCE[0]}" )/.." && pwd )"
ENV_FILE="$PROJECT_ROOT/.env.local"

# 检测当前系统用户
CURRENT_USER=$(whoami)
echo "📋 检测到系统用户: $CURRENT_USER"

# 检测PostgreSQL是否运行
if ! pg_isready > /dev/null 2>&1; then
    echo "⚠️  PostgreSQL服务未运行，正在启动..."
    brew services start postgresql@14
    sleep 2
fi

# 检测PostgreSQL用户
PG_USER=$(psql postgres -t -c "SELECT current_user;" 2>/dev/null | xargs)
if [ -z "$PG_USER" ]; then
    PG_USER=$CURRENT_USER
fi

echo "📋 PostgreSQL用户: $PG_USER"

# 检查数据库是否存在
if psql -lqt | cut -d \| -f 1 | grep -qw nofx_test; then
    echo "✅ 数据库 nofx_test 已存在"
else
    echo "📦 创建数据库 nofx_test..."
    createdb nofx_test
    echo "✅ 数据库创建成功"
fi

# 生成正确的DATABASE_URL
DATABASE_URL="postgresql://${PG_USER}@localhost:5432/nofx_test?sslmode=disable"

# 测试连接
echo ""
echo "🔍 测试数据库连接..."
if psql "$DATABASE_URL" -c "SELECT 1;" > /dev/null 2>&1; then
    echo "✅ 数据库连接成功"
else
    echo "❌ 数据库连接失败"
    exit 1
fi

# 更新或创建 .env.local
echo ""
echo "📝 更新 .env.local 配置..."

cat > "$ENV_FILE" << EOF
# ============================================================
# NoFX 本地测试环境配置
# ============================================================
# 自动生成于: $(date)
# ============================================================

# 环境标识
APP_ENV=local

# 本地PostgreSQL数据库配置
DATABASE_URL=$DATABASE_URL

# API服务器端口
API_SERVER_PORT=8080

# 日志级别
LOG_LEVEL=debug
EOF

echo "✅ 配置文件已更新"

# 初始化数据库表
echo ""
echo "🏗️  初始化数据库表..."
if [ -f "$PROJECT_ROOT/scripts/init_database.sql" ]; then
    psql "$DATABASE_URL" -f "$PROJECT_ROOT/scripts/init_database.sql" > /dev/null 2>&1
    echo "✅ 数据库表初始化完成"
else
    echo "⚠️  未找到 init_database.sql，跳过表初始化"
fi

echo ""
echo "======================================"
echo "✅ 本地环境配置完成！"
echo "======================================"
echo ""
echo "📊 配置信息:"
echo "  用户: $PG_USER"
echo "  数据库: nofx_test"
echo "  连接: $DATABASE_URL"
echo ""
echo "🚀 现在可以启动应用:"
echo "  ./scripts/run_local.sh"
echo ""
