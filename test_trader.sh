#!/bin/bash

echo "🔍 测试交易员接口"
echo "================================"
echo ""

# 从浏览器复制你的 token 到这里
TOKEN="YOUR_TOKEN_HERE"

TRADER_ID="binance_edad5285-6a9f-4612-b963-034186a257d9_deepseek_1763310123"

echo "1️⃣ 测试获取交易员列表..."
curl -s -H "Authorization: Bearer $TOKEN" \
     http://localhost:8080/api/my-traders | jq .

echo ""
echo "2️⃣ 测试获取交易员配置..."
curl -s -H "Authorization: Bearer $TOKEN" \
     http://localhost:8080/api/traders/$TRADER_ID/config | jq .

echo ""
echo "3️⃣ 测试启动交易员..."
curl -s -X POST -H "Authorization: Bearer $TOKEN" \
     http://localhost:8080/api/traders/$TRADER_ID/start | jq .

echo ""
echo "================================"
echo "✅ 测试完成"
