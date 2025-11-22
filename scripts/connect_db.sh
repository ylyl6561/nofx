#!/bin/bash

# ============================================================
# NoFX 数据库连接脚本
# ============================================================

set -e

echo "🗄️  NoFX 数据库连接工具"
echo "======================================"
echo ""
echo "选择要连接的数据库:"
echo "  1) 本地测试数据库 (localhost)"
echo "  2) 生产数据库 (zeabur)"
echo "  3) 自定义连接"
echo ""
read -p "请选择 (1-3): " choice

case $choice in
    1)
        echo ""
        echo "📋 连接本地测试数据库..."
        
        # 尝试从 .env.local 读取
        PROJECT_ROOT="$( cd "$( dirname "${BASH_SOURCE[0]}" )/.." && pwd )"
        if [ -f "$PROJECT_ROOT/.env.local" ]; then
            export $(cat "$PROJECT_ROOT/.env.local" | grep DATABASE_URL | xargs)
        else
            # 使用默认配置
            DATABASE_URL="postgresql://postgres@localhost:5432/nofx_test?sslmode=disable"
        fi
        
        echo "🔗 连接: $DATABASE_URL"
        echo ""
        psql "$DATABASE_URL"
        ;;
        
    2)
        echo ""
        echo "📋 连接生产数据库..."
        
        # 尝试从 .env.production 读取
        PROJECT_ROOT="$( cd "$( dirname "${BASH_SOURCE[0]}" )/.." && pwd )"
        if [ -f "$PROJECT_ROOT/.env.production" ]; then
            export $(cat "$PROJECT_ROOT/.env.production" | grep DATABASE_URL | xargs)
        else
            echo "❌ 未找到 .env.production 文件"
            echo "请先创建 .env.production 并配置生产数据库连接"
            exit 1
        fi
        
        echo "🔗 连接: ${DATABASE_URL:0:50}..."
        echo ""
        echo "⚠️  警告: 你正在连接生产数据库，请谨慎操作！"
        read -p "确认继续? (y/n) " -n 1 -r
        echo
        if [[ ! $REPLY =~ ^[Yy]$ ]]; then
            echo "已取消"
            exit 0
        fi
        
        psql "$DATABASE_URL"
        ;;
        
    3)
        echo ""
        echo "📋 自定义连接"
        read -p "Host: " db_host
        read -p "Port (5432): " db_port
        db_port=${db_port:-5432}
        read -p "User: " db_user
        read -sp "Password: " db_password
        echo
        read -p "Database: " db_name
        
        DATABASE_URL="postgresql://$db_user:$db_password@$db_host:$db_port/$db_name?sslmode=disable"
        
        echo ""
        echo "🔗 连接: postgresql://$db_user:***@$db_host:$db_port/$db_name"
        echo ""
        psql "$DATABASE_URL"
        ;;
        
    *)
        echo "❌ 无效选择"
        exit 1
        ;;
esac
