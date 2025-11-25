#!/bin/bash

# Trader 诊断脚本
# 用于快速排查为什么没有新的决策周期数据产生

echo "=========================================="
echo "🔍 Trader 诊断工具"
echo "=========================================="
echo ""

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 1. 检查进程是否运行
echo -e "${BLUE}[1/7] 检查进程状态...${NC}"
if pgrep -f "nofx" > /dev/null; then
    echo -e "${GREEN}✅ nofx 进程正在运行${NC}"
    ps aux | grep nofx | grep -v grep
else
    echo -e "${RED}❌ nofx 进程未运行！${NC}"
    echo "   请先启动服务："
    echo "   - systemctl start nofx.service"
    echo "   - 或 pm2 start nofx"
    echo "   - 或 ./start.sh"
    exit 1
fi
echo ""

# 2. 检查数据库中 trader 的状态
echo -e "${BLUE}[2/7] 检查数据库中的 Trader 状态...${NC}"
if command -v psql &> /dev/null; then
    # 假设数据库连接信息在环境变量中
    TRADERS=$(psql -t -c "SELECT id, name, is_running, scan_interval_minutes FROM traders;" 2>/dev/null)
    if [ $? -eq 0 ]; then
        echo "$TRADERS" | while read -r line; do
            if [[ ! -z "$line" ]]; then
                ID=$(echo $line | awk '{print $1}')
                NAME=$(echo $line | awk '{print $2}')
                IS_RUNNING=$(echo $line | awk '{print $3}')
                INTERVAL=$(echo $line | awk '{print $4}')
                
                if [ "$IS_RUNNING" = "t" ]; then
                    echo -e "${GREEN}✅ Trader: $NAME (ID: $ID)${NC}"
                    echo "   状态: 运行中"
                    echo "   扫描间隔: ${INTERVAL} 分钟"
                else
                    echo -e "${RED}❌ Trader: $NAME (ID: $ID)${NC}"
                    echo "   状态: 已停止"
                    echo "   扫描间隔: ${INTERVAL} 分钟"
                    echo -e "${YELLOW}   💡 提示: 请在页面上启动此 Trader${NC}"
                fi
            fi
        done
    else
        echo -e "${YELLOW}⚠️  无法连接数据库，跳过此检查${NC}"
    fi
else
    echo -e "${YELLOW}⚠️  psql 未安装，跳过数据库检查${NC}"
fi
echo ""

# 3. 检查最近的日志
echo -e "${BLUE}[3/7] 检查最近的日志...${NC}"
if [ -d "logs" ]; then
    LATEST_LOG=$(ls -t logs/trader_*.log 2>/dev/null | head -1)
    if [ ! -z "$LATEST_LOG" ]; then
        echo "最新日志文件: $LATEST_LOG"
        
        # 检查最后一条日志的时间
        LAST_LOG_TIME=$(tail -1 "$LATEST_LOG" 2>/dev/null | grep -oP '\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2}' | head -1)
        if [ ! -z "$LAST_LOG_TIME" ]; then
            echo "最后日志时间: $LAST_LOG_TIME"
            
            # 计算距离现在多久
            LAST_TIMESTAMP=$(date -d "$LAST_LOG_TIME" +%s 2>/dev/null || date -j -f "%Y-%m-%d %H:%M:%S" "$LAST_LOG_TIME" +%s 2>/dev/null)
            NOW_TIMESTAMP=$(date +%s)
            DIFF=$((NOW_TIMESTAMP - LAST_TIMESTAMP))
            DIFF_MINUTES=$((DIFF / 60))
            
            if [ $DIFF_MINUTES -lt 5 ]; then
                echo -e "${GREEN}✅ 日志很新 (${DIFF_MINUTES} 分钟前)${NC}"
            elif [ $DIFF_MINUTES -lt 30 ]; then
                echo -e "${YELLOW}⚠️  日志有点旧 (${DIFF_MINUTES} 分钟前)${NC}"
            else
                echo -e "${RED}❌ 日志很旧 (${DIFF_MINUTES} 分钟前)${NC}"
                echo -e "${YELLOW}   💡 Trader 可能已经停止运行${NC}"
            fi
        fi
        
        # 检查最近的错误
        echo ""
        echo "最近的错误日志 (最后 10 条):"
        grep -i "error\|failed\|失败\|❌" "$LATEST_LOG" | tail -10
        
        # 检查是否有决策周期
        echo ""
        echo "最近的决策周期 (最后 3 个):"
        grep "AI决策周期" "$LATEST_LOG" | tail -3
        
    else
        echo -e "${RED}❌ 未找到日志文件${NC}"
    fi
else
    echo -e "${RED}❌ logs 目录不存在${NC}"
fi
echo ""

# 4. 检查最近的决策记录
echo -e "${BLUE}[4/7] 检查数据库中的最近决策记录...${NC}"
if command -v psql &> /dev/null; then
    RECENT_DECISIONS=$(psql -t -c "
        SELECT 
            trader_id,
            cycle_number,
            timestamp,
            success,
            error_message
        FROM decision_logs
        ORDER BY timestamp DESC
        LIMIT 5;
    " 2>/dev/null)
    
    if [ $? -eq 0 ]; then
        if [ ! -z "$RECENT_DECISIONS" ]; then
            echo "$RECENT_DECISIONS"
            
            # 检查最后一条记录的时间
            LAST_DECISION_TIME=$(psql -t -c "SELECT MAX(timestamp) FROM decision_logs;" 2>/dev/null | xargs)
            if [ ! -z "$LAST_DECISION_TIME" ]; then
                echo ""
                echo "最后决策时间: $LAST_DECISION_TIME"
                
                # 计算距离现在多久
                LAST_TS=$(date -d "$LAST_DECISION_TIME" +%s 2>/dev/null || date -j -f "%Y-%m-%d %H:%M:%S" "$LAST_DECISION_TIME" +%s 2>/dev/null)
                NOW_TS=$(date +%s)
                DIFF=$((NOW_TS - LAST_TS))
                DIFF_MIN=$((DIFF / 60))
                
                if [ $DIFF_MIN -lt 5 ]; then
                    echo -e "${GREEN}✅ 决策记录很新 (${DIFF_MIN} 分钟前)${NC}"
                elif [ $DIFF_MIN -lt 30 ]; then
                    echo -e "${YELLOW}⚠️  决策记录有点旧 (${DIFF_MIN} 分钟前)${NC}"
                else
                    echo -e "${RED}❌ 决策记录很旧 (${DIFF_MIN} 分钟前)${NC}"
                    echo -e "${YELLOW}   💡 Trader 可能已经停止决策${NC}"
                fi
            fi
        else
            echo -e "${YELLOW}⚠️  数据库中没有决策记录${NC}"
        fi
    fi
fi
echo ""

# 5. 检查是否有风险控制暂停
echo -e "${BLUE}[5/7] 检查是否触发风险控制暂停...${NC}"
if [ ! -z "$LATEST_LOG" ]; then
    RISK_PAUSE=$(grep "风险控制：暂停交易中" "$LATEST_LOG" | tail -1)
    if [ ! -z "$RISK_PAUSE" ]; then
        echo -e "${YELLOW}⚠️  检测到风险控制暂停:${NC}"
        echo "$RISK_PAUSE"
        echo -e "${YELLOW}   💡 等待暂停时间结束后会自动恢复${NC}"
    else
        echo -e "${GREEN}✅ 未检测到风险控制暂停${NC}"
    fi
fi
echo ""

# 6. 检查 API 连接
echo -e "${BLUE}[6/7] 检查关键服务连接...${NC}"

# 检查 API 端口
if netstat -tlnp 2>/dev/null | grep -q ":8080"; then
    echo -e "${GREEN}✅ API 服务正在监听 8080 端口${NC}"
else
    echo -e "${RED}❌ API 服务未监听 8080 端口${NC}"
fi

# 检查数据库连接
if command -v psql &> /dev/null; then
    if psql -c "SELECT 1;" &>/dev/null; then
        echo -e "${GREEN}✅ 数据库连接正常${NC}"
    else
        echo -e "${RED}❌ 数据库连接失败${NC}"
    fi
fi
echo ""

# 7. 提供诊断建议
echo -e "${BLUE}[7/7] 诊断建议${NC}"
echo ""

# 检查是否有明显问题
HAS_ISSUE=false

# 检查进程
if ! pgrep -f "nofx" > /dev/null; then
    echo -e "${RED}❌ 问题: nofx 进程未运行${NC}"
    echo "   解决方案: 启动服务"
    echo "   命令: systemctl start nofx.service 或 pm2 start nofx"
    HAS_ISSUE=true
fi

# 检查日志时间
if [ ! -z "$DIFF_MINUTES" ] && [ $DIFF_MINUTES -gt 10 ]; then
    echo -e "${RED}❌ 问题: 日志超过 10 分钟未更新${NC}"
    echo "   可能原因:"
    echo "   1. Trader 已停止 (检查数据库 is_running 字段)"
    echo "   2. 程序崩溃 (检查系统日志)"
    echo "   3. 卡在某个操作 (检查日志最后的错误)"
    echo ""
    echo "   解决方案:"
    echo "   1. 重启服务: systemctl restart nofx.service"
    echo "   2. 检查日志: tail -f logs/trader_*.log"
    echo "   3. 在页面上重新启动 Trader"
    HAS_ISSUE=true
fi

if [ "$HAS_ISSUE" = false ]; then
    echo -e "${GREEN}✅ 未发现明显问题${NC}"
    echo ""
    echo "如果仍然没有新数据，请检查:"
    echo "1. 页面上 Trader 的状态是否显示为 '运行中'"
    echo "2. 浏览器控制台是否有错误"
    echo "3. 网络连接是否正常"
fi

echo ""
echo "=========================================="
echo "🔍 诊断完成"
echo "=========================================="
echo ""
echo "💡 快速操作命令:"
echo ""
echo "查看实时日志:"
echo "  tail -f logs/trader_*.log"
echo ""
echo "查看最近的错误:"
echo "  grep -i 'error\|failed\|失败' logs/trader_*.log | tail -20"
echo ""
echo "查看最近的决策周期:"
echo "  grep 'AI决策周期' logs/trader_*.log | tail -10"
echo ""
echo "重启服务:"
echo "  systemctl restart nofx.service"
echo "  或 pm2 restart nofx"
echo ""
