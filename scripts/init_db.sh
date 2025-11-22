#!/bin/bash

# ============================================================
# NoFX Trading System - Database Initialization Script
# ============================================================
# 此脚本用于初始化PostgreSQL数据库
# ============================================================

set -e  # 遇到错误立即退出

echo "🚀 NoFX 数据库初始化脚本"
echo "======================================"

# 检查环境变量
if [ -z "$DATABASE_URL" ]; then
    echo "❌ 错误: 未设置 DATABASE_URL 环境变量"
    echo ""
    echo "请设置环境变量，例如："
    echo "export DATABASE_URL='postgresql://user:password@host:port/database'"
    echo ""
    echo "或者使用单独的环境变量："
    echo "export DB_HOST=localhost"
    echo "export DB_PORT=5432"
    echo "export DB_USER=postgres"
    echo "export DB_PASSWORD=your_password"
    echo "export DB_NAME=nofx"
    exit 1
fi

# 解析 DATABASE_URL
if [ -n "$DATABASE_URL" ]; then
    echo "📋 使用 DATABASE_URL 连接数据库"
    DB_CONNECTION="$DATABASE_URL"
else
    # 使用单独的环境变量
    DB_HOST=${DB_HOST:-localhost}
    DB_PORT=${DB_PORT:-5432}
    DB_USER=${DB_USER:-postgres}
    DB_PASSWORD=${DB_PASSWORD}
    DB_NAME=${DB_NAME:-nofx}
    
    echo "📋 连接信息:"
    echo "  Host: $DB_HOST"
    echo "  Port: $DB_PORT"
    echo "  User: $DB_USER"
    echo "  Database: $DB_NAME"
    
    DB_CONNECTION="postgresql://$DB_USER:$DB_PASSWORD@$DB_HOST:$DB_PORT/$DB_NAME"
fi

echo ""
echo "🔍 检查数据库连接..."

# 测试连接
if psql "$DB_CONNECTION" -c "SELECT 1;" > /dev/null 2>&1; then
    echo "✅ 数据库连接成功"
else
    echo "❌ 数据库连接失败，请检查连接信息"
    exit 1
fi

echo ""
echo "📦 开始执行数据库初始化..."
echo ""

# 获取脚本所在目录
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
SQL_FILE="$SCRIPT_DIR/init_database.sql"

# 检查SQL文件是否存在
if [ ! -f "$SQL_FILE" ]; then
    echo "❌ 错误: 找不到 init_database.sql 文件"
    echo "路径: $SQL_FILE"
    exit 1
fi

# 执行SQL脚本
psql "$DB_CONNECTION" -f "$SQL_FILE"

echo ""
echo "======================================"
echo "✅ 数据库初始化完成！"
echo ""
echo "📊 已创建的表:"
echo "  1. system_ai_models - 系统AI模型模板"
echo "  2. system_exchanges - 系统交易所模板"
echo "  3. ai_models - AI模型配置"
echo "  4. exchanges - 交易所配置"
echo "  5. user_signal_sources - 用户信号源"
echo "  6. traders - 交易员配置"
echo "  7. users - 用户表"
echo "  8. system_config - 系统配置"
echo "  9. beta_codes - 内测码"
echo "  10. api_keys - API密钥"
echo "  11. api_usage_logs - API使用日志"
echo "  12. user_trading_config - 用户交易配置"
echo "  13. decision_logs - 决策日志 (新增)"
echo "  14. equity_history - 权益历史 (新增)"
echo ""
echo "🎉 现在可以启动应用了: go run main.go"
