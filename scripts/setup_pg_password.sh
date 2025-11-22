#!/bin/bash

# ============================================================
# 为PostgreSQL设置密码（用于DBeaver等客户端连接）
# ============================================================

set -e

echo "🔐 PostgreSQL密码设置工具"
echo "======================================"
echo ""

# 获取当前用户
CURRENT_USER=$(whoami)
echo "📋 当前用户: $CURRENT_USER"

# 检查PostgreSQL是否运行
if ! pg_isready > /dev/null 2>&1; then
    echo "⚠️  PostgreSQL服务未运行，正在启动..."
    brew services start postgresql@14
    sleep 2
fi

echo "✅ PostgreSQL服务正在运行"
echo ""

# 提示输入密码
echo "请设置PostgreSQL密码（用于DBeaver等客户端连接）"
echo "建议使用简单密码，如: local123"
echo ""
read -p "输入密码: " -s PG_PASSWORD
echo ""
read -p "确认密码: " -s PG_PASSWORD_CONFIRM
echo ""

if [ "$PG_PASSWORD" != "$PG_PASSWORD_CONFIRM" ]; then
    echo "❌ 两次密码不一致"
    exit 1
fi

if [ -z "$PG_PASSWORD" ]; then
    echo "❌ 密码不能为空"
    exit 1
fi

# 设置密码
echo "🔧 正在设置密码..."
psql postgres -c "ALTER USER $CURRENT_USER WITH PASSWORD '$PG_PASSWORD';" > /dev/null 2>&1

if [ $? -eq 0 ]; then
    echo "✅ 密码设置成功"
else
    echo "❌ 密码设置失败"
    exit 1
fi

# 配置PostgreSQL允许密码认证
echo ""
echo "🔧 配置PostgreSQL认证方式..."

PG_HBA_CONF="/opt/homebrew/var/postgresql@14/pg_hba.conf"

if [ -f "$PG_HBA_CONF" ]; then
    # 备份配置文件
    cp "$PG_HBA_CONF" "$PG_HBA_CONF.backup"
    
    # 修改认证方式
    sed -i '' 's/host    all             all             127.0.0.1\/32            trust/host    all             all             127.0.0.1\/32            md5/g' "$PG_HBA_CONF"
    
    echo "✅ 配置文件已更新"
    echo "📋 备份文件: $PG_HBA_CONF.backup"
else
    echo "⚠️  未找到配置文件，跳过"
fi

# 重启PostgreSQL
echo ""
echo "🔄 重启PostgreSQL服务..."
brew services restart postgresql@14
sleep 2

# 测试连接
echo ""
echo "🔍 测试密码连接..."
PGPASSWORD=$PG_PASSWORD psql -h localhost -U $CURRENT_USER -d postgres -c "SELECT 1;" > /dev/null 2>&1

if [ $? -eq 0 ]; then
    echo "✅ 密码连接测试成功"
else
    echo "❌ 密码连接测试失败"
    exit 1
fi

# 更新 .env.local
echo ""
echo "📝 更新 .env.local 配置..."

PROJECT_ROOT="$( cd "$( dirname "${BASH_SOURCE[0]}" )/.." && pwd )"
ENV_FILE="$PROJECT_ROOT/.env.local"

# 生成新的DATABASE_URL
NEW_DATABASE_URL="postgresql://$CURRENT_USER:$PG_PASSWORD@localhost:5432/nofx_test?sslmode=disable"

if [ -f "$ENV_FILE" ]; then
    # 备份
    cp "$ENV_FILE" "$ENV_FILE.backup"
    
    # 更新DATABASE_URL
    if grep -q "^DATABASE_URL=" "$ENV_FILE"; then
        sed -i '' "s|^DATABASE_URL=.*|DATABASE_URL=$NEW_DATABASE_URL|" "$ENV_FILE"
        echo "✅ .env.local 已更新"
    else
        echo "DATABASE_URL=$NEW_DATABASE_URL" >> "$ENV_FILE"
        echo "✅ DATABASE_URL 已添加到 .env.local"
    fi
else
    echo "⚠️  未找到 .env.local，请手动创建"
fi

echo ""
echo "======================================"
echo "✅ PostgreSQL密码设置完成！"
echo "======================================"
echo ""
echo "📊 连接信息:"
echo "  Host: localhost"
echo "  Port: 5432"
echo "  Database: nofx_test"
echo "  Username: $CURRENT_USER"
echo "  Password: $PG_PASSWORD"
echo ""
echo "🔧 DBeaver连接配置:"
echo "  1. 新建PostgreSQL连接"
echo "  2. 填入上述连接信息"
echo "  3. 测试连接"
echo ""
echo "📝 .env.local 配置:"
echo "  DATABASE_URL=$NEW_DATABASE_URL"
echo ""
