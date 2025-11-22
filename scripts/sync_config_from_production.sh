#!/bin/bash

# ============================================================
# 从生产环境同步配置数据到本地测试环境
# ============================================================

set -e

echo "🔄 配置数据同步工具"
echo "======================================"
echo ""

# 获取脚本所在目录的父目录（项目根目录）
PROJECT_ROOT="$( cd "$( dirname "${BASH_SOURCE[0]}" )/.." && pwd )"

# 加载环境配置
if [ -f "$PROJECT_ROOT/.env.production" ]; then
    echo "📋 加载生产环境配置..."
    export PROD_DATABASE_URL=$(cat "$PROJECT_ROOT/.env.production" | grep DATABASE_URL | cut -d '=' -f 2-)
else
    echo "❌ 错误: 未找到 .env.production 文件"
    exit 1
fi

if [ -f "$PROJECT_ROOT/.env.local" ]; then
    echo "📋 加载本地环境配置..."
    export LOCAL_DATABASE_URL=$(cat "$PROJECT_ROOT/.env.local" | grep DATABASE_URL | cut -d '=' -f 2-)
else
    echo "❌ 错误: 未找到 .env.local 文件"
    exit 1
fi

# 去除引号
PROD_DATABASE_URL=$(echo $PROD_DATABASE_URL | tr -d "'\"")
LOCAL_DATABASE_URL=$(echo $LOCAL_DATABASE_URL | tr -d "'\"")

echo ""
echo "📊 数据库连接信息:"
echo "  生产: ${PROD_DATABASE_URL:0:50}..."
echo "  本地: ${LOCAL_DATABASE_URL:0:50}..."
echo ""

# 创建临时目录
TEMP_DIR="$PROJECT_ROOT/temp_sync"
mkdir -p "$TEMP_DIR"

echo "======================================"
echo "📦 开始导出生产环境配置数据..."
echo "======================================"
echo ""

# 导出各个配置表的数据
TABLES=(
    "system_ai_models"
    "system_exchanges"
    "system_config"
    "ai_models"
    "exchanges"
    "user_signal_sources"
)

for table in "${TABLES[@]}"; do
    echo "📤 导出表: $table"
    
    # 导出为 SQL INSERT 语句
    psql "$PROD_DATABASE_URL" -c "\copy (SELECT * FROM $table) TO '$TEMP_DIR/${table}.csv' WITH CSV HEADER" 2>/dev/null || {
        echo "⚠️  表 $table 可能为空或不存在，跳过"
        continue
    }
    
    if [ -f "$TEMP_DIR/${table}.csv" ]; then
        # 统计记录数
        count=$(wc -l < "$TEMP_DIR/${table}.csv")
        count=$((count - 1))  # 减去表头
        echo "   ✓ 导出 $count 条记录"
    fi
done

echo ""
echo "======================================"
echo "📥 开始导入到本地测试环境..."
echo "======================================"
echo ""

# 确认导入
read -p "⚠️  这将覆盖本地数据库中的配置数据，是否继续? (y/n) " -n 1 -r
echo
if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    echo "❌ 已取消"
    rm -rf "$TEMP_DIR"
    exit 0
fi

for table in "${TABLES[@]}"; do
    if [ ! -f "$TEMP_DIR/${table}.csv" ]; then
        echo "⚠️  跳过表 $table (未导出)"
        continue
    fi
    
    echo "📥 导入表: $table"
    
    # 清空表（保留结构）
    psql "$LOCAL_DATABASE_URL" -c "TRUNCATE TABLE $table CASCADE;" 2>/dev/null || {
        echo "⚠️  清空表 $table 失败，可能不存在"
        continue
    }
    
    # 导入数据
    psql "$LOCAL_DATABASE_URL" -c "\copy $table FROM '$TEMP_DIR/${table}.csv' WITH CSV HEADER" 2>/dev/null && {
        count=$(wc -l < "$TEMP_DIR/${table}.csv")
        count=$((count - 1))
        echo "   ✓ 导入 $count 条记录"
    } || {
        echo "   ❌ 导入失败"
    }
done

echo ""
echo "======================================"
echo "🧹 清理临时文件..."
echo "======================================"
rm -rf "$TEMP_DIR"
echo "✓ 清理完成"

echo ""
echo "======================================"
echo "✅ 配置数据同步完成！"
echo "======================================"
echo ""
echo "📊 已同步的表:"
for table in "${TABLES[@]}"; do
    echo "  - $table"
done
echo ""
echo "💡 提示:"
echo "  1. 用户数据(users)和交易员配置(traders)未同步，避免冲突"
echo "  2. 决策日志(decision_logs)和权益历史(equity_history)未同步"
echo "  3. 如需同步其他数据，请手动修改此脚本"
echo ""
