#!/bin/bash

# ============================================================
# NoFX 本地测试数据库快速设置脚本
# ============================================================

set -e

echo "🗄️  NoFX 本地数据库设置"
echo "======================================"

# 默认配置
DB_NAME=${DB_NAME:-nofx_test}
DB_USER=${DB_USER:-postgres}
DB_HOST=${DB_HOST:-localhost}
DB_PORT=${DB_PORT:-5432}

echo "📋 数据库配置:"
echo "  名称: $DB_NAME"
echo "  用户: $DB_USER"
echo "  主机: $DB_HOST"
echo "  端口: $DB_PORT"
echo ""

# 检查PostgreSQL是否运行
echo "🔍 检查 PostgreSQL 服务..."
if ! command -v psql &> /dev/null; then
    echo "❌ 错误: 未找到 psql 命令"
    echo ""
    echo "请先安装 PostgreSQL:"
    echo "  macOS:   brew install postgresql@14"
    echo "  Ubuntu:  sudo apt-get install postgresql"
    echo "  CentOS:  sudo yum install postgresql-server"
    exit 1
fi

# 检查PostgreSQL服务状态
if ! pg_isready -h $DB_HOST -p $DB_PORT > /dev/null 2>&1; then
    echo "⚠️  PostgreSQL 服务未运行"
    echo ""
    echo "启动 PostgreSQL:"
    echo "  macOS:   brew services start postgresql@14"
    echo "  Ubuntu:  sudo systemctl start postgresql"
    echo "  CentOS:  sudo systemctl start postgresql"
    echo ""
    read -p "是否现在启动? (y/n) " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        if [[ "$OSTYPE" == "darwin"* ]]; then
            brew services start postgresql@14 || brew services start postgresql
        else
            sudo systemctl start postgresql
        fi
    else
        exit 1
    fi
fi

echo "✅ PostgreSQL 服务正在运行"
echo ""

# 创建数据库
echo "🏗️  创建数据库 '$DB_NAME'..."
if psql -h $DB_HOST -p $DB_PORT -U $DB_USER -lqt | cut -d \| -f 1 | grep -qw $DB_NAME; then
    echo "⚠️  数据库 '$DB_NAME' 已存在"
    read -p "是否删除并重新创建? (y/n) " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        dropdb -h $DB_HOST -p $DB_PORT -U $DB_USER $DB_NAME
        createdb -h $DB_HOST -p $DB_PORT -U $DB_USER $DB_NAME
        echo "✅ 数据库已重新创建"
    else
        echo "ℹ️  使用现有数据库"
    fi
else
    createdb -h $DB_HOST -p $DB_PORT -U $DB_USER $DB_NAME
    echo "✅ 数据库创建成功"
fi

echo ""

# 初始化数据库表
echo "📦 初始化数据库表..."
PROJECT_ROOT="$( cd "$( dirname "${BASH_SOURCE[0]}" )/.." && pwd )"
SQL_FILE="$PROJECT_ROOT/scripts/init_database.sql"

if [ ! -f "$SQL_FILE" ]; then
    echo "❌ 错误: 找不到 init_database.sql"
    exit 1
fi

DATABASE_URL="postgresql://$DB_USER@$DB_HOST:$DB_PORT/$DB_NAME?sslmode=disable"
psql "$DATABASE_URL" -f "$SQL_FILE"

echo ""
echo "======================================"
echo "✅ 本地数据库设置完成！"
echo ""
echo "📝 连接信息:"
echo "  DATABASE_URL=$DATABASE_URL"
echo ""
echo "💡 下一步:"
echo "  1. 编辑 .env.local 文件，填入数据库密码"
echo "  2. 运行: ./scripts/run_local.sh"
echo ""
echo "🔧 或者直接使用环境变量:"
echo "  export DATABASE_URL='$DATABASE_URL'"
echo "  go run main.go"
