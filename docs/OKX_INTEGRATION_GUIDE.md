# 🏦 OKX 交易所集成指南

## 📋 概述

AI-Trader 现已支持 OKX 交易所！本指南将帮助你快速接入 OKX API 进行量化交易。

---

## 🔑 获取 OKX API 密钥

### 步骤 1：注册 OKX 账户
1. 访问 [OKX 官网](https://www.okx.com)
2. 注册并完成 KYC 认证
3. 充值 USDT 到交易账户

### 步骤 2：创建 API Key
1. 登录 OKX
2. 进入 **个人中心** → **API**
3. 点击 **创建 API**
4. 设置 API 权限：
   - ✅ **读取** - 必需
   - ✅ **交易** - 必需
   - ❌ **提币** - 不建议开启
5. 设置 IP 白名单（推荐）
6. 记录以下信息：
   - **API Key**
   - **Secret Key**
   - **Passphrase**（创建时设置的密码）

⚠️ **重要提示**：
- Secret Key 只显示一次，请妥善保存
- Passphrase 是你创建 API 时设置的密码，不是登录密码
- 建议使用 IP 白名单提高安全性

---

## ⚙️ 配置 AI-Trader

### 方式 1：通过 Web 界面配置

1. 登录 AI-Trader Web 界面
2. 进入 **交易所配置**
3. 添加 OKX 配置：
   ```json
   {
     "id": "okx_main",
     "name": "OKX主账户",
     "type": "okx",
     "api_key": "your_okx_api_key",
     "secret_key": "your_okx_secret_key",
     "passphrase": "your_okx_passphrase"
   }
   ```

### 方式 2：通过配置文件

编辑 `config.json`：

```json
{
  "traders": [
    {
      "id": "okx_trader_001",
      "name": "OKX量化交易员",
      "enabled": true,
      "exchange": "okx",
      
      "okx_api_key": "your_okx_api_key",
      "okx_secret_key": "your_okx_secret_key",
      "okx_passphrase": "your_okx_passphrase",
      
      "ai_model": "deepseek",
      "deepseek_key": "your_deepseek_key",
      
      "initial_balance": 1000,
      "scan_interval_minutes": 60
    }
  ],
  "leverage": {
    "btc_eth_leverage": 10,
    "altcoin_leverage": 5
  }
}
```

---

## 🚀 启动交易

### 本地运行
```bash
cd /Users/yuliang/project/nofx
./ai-trader
```

### Docker 运行
```bash
docker compose up -d ai-trader
docker compose logs -f ai-trader
```

---

## 📊 OKX API 功能说明

### 支持的功能

| 功能 | 状态 | 说明 |
|------|------|------|
| **获取余额** | ✅ | 查询账户USDT余额 |
| **获取持仓** | ✅ | 查询当前持仓信息 |
| **获取价格** | ✅ | 实时市场价格 |
| **开仓** | ✅ | 市价开多/空仓 |
| **平仓** | ✅ | 市价平仓 |
| **设置止损** | ✅ | 条件单止损 |
| **取消订单** | ✅ | 取消止盈止损单 |
| **设置杠杆** | ✅ | 动态调整杠杆倍数 |

### 交易模式
- **仓位模式**：全仓（Cross Margin）
- **持仓模式**：双向持仓（可同时持有多空）
- **订单类型**：市价单 + 条件止损单

---

## 🔍 API 调用示例

### 1. 查询余额
```bash
curl -X GET http://localhost:8080/api/v1/traders/okx_trader_001 \
  -H "Authorization: Bearer aitrader_your_api_key"
```

### 2. 启动交易员
```bash
curl -X POST http://localhost:8080/api/v1/traders/okx_trader_001/start \
  -H "Authorization: Bearer aitrader_your_api_key"
```

### 3. 查看绩效
```bash
curl -X GET http://localhost:8080/api/v1/traders/okx_trader_001/performance \
  -H "Authorization: Bearer aitrader_your_api_key"
```

---

## 🎯 交易币种格式

### OKX 币种格式转换

| 标准格式 | OKX 格式 | 说明 |
|---------|---------|------|
| BTCUSDT | BTC-USDT-SWAP | 永续合约 |
| ETHUSDT | ETH-USDT-SWAP | 永续合约 |
| SOLUSDT | SOL-USDT-SWAP | 永续合约 |

AI-Trader 会自动处理格式转换，你只需使用标准格式（如 BTCUSDT）。

---

## ⚠️ 风险控制

### 建议配置

```json
{
  "initial_balance": 1000,
  "leverage": {
    "btc_eth_leverage": 10,
    "altcoin_leverage": 5
  },
  "max_daily_loss": 50,
  "max_drawdown": 20
}
```

### 安全建议
1. **小额测试**：先用小额资金测试
2. **设置止损**：每笔交易都设置止损
3. **控制杠杆**：新手建议 3-5 倍杠杆
4. **分散投资**：不要将所有资金投入单一币种
5. **定期检查**：监控交易日志和绩效

---

## 🐛 常见问题

### Q1: API Key 无效？
**A:** 检查以下几点：
- API Key、Secret Key、Passphrase 是否正确
- API 权限是否包含"读取"和"交易"
- IP 白名单是否正确配置
- 账户是否完成 KYC 认证

### Q2: 签名错误？
**A:** 
- 确认 Secret Key 没有多余的空格
- 确认 Passphrase 是创建 API 时设置的密码
- 检查系统时间是否准确

### Q3: 余额不足？
**A:**
- 确认交易账户有足够的 USDT
- OKX 需要在"交易账户"而非"资金账户"
- 可以在 OKX 界面进行账户间划转

### Q4: 订单被拒绝？
**A:**
- 检查杠杆设置是否合理
- 确认币种是否支持永续合约
- 查看 OKX 的最小交易量要求

---

## 📈 性能优化

### 缓存机制
- **余额缓存**：15秒
- **持仓缓存**：15秒
- 减少 API 调用，避免触发限流

### 限流说明
OKX API 限流规则：
- 普通接口：20次/2秒
- 交易接口：60次/2秒
- 建议设置 `scan_interval_minutes >= 5`

---

## 🔄 从其他交易所迁移

### 从 Binance 迁移
```json
// Binance 配置
{
  "exchange": "binance",
  "binance_api_key": "...",
  "binance_secret_key": "..."
}

// 改为 OKX
{
  "exchange": "okx",
  "okx_api_key": "...",
  "okx_secret_key": "...",
  "okx_passphrase": "..."
}
```

### 多交易所同时使用
```json
{
  "traders": [
    {
      "id": "binance_trader",
      "exchange": "binance",
      ...
    },
    {
      "id": "okx_trader",
      "exchange": "okx",
      ...
    }
  ]
}
```

---

## 📞 技术支持

### 日志查看
```bash
# Docker
docker compose logs -f ai-trader | grep OKX

# 本地
tail -f ai-trader.log | grep OKX
```

### 调试模式
设置环境变量启用详细日志：
```bash
DEBUG=true ./ai-trader
```

---

## ✅ 检查清单

部署前确认：
- [ ] 已获取 OKX API Key、Secret Key、Passphrase
- [ ] API 权限包含"读取"和"交易"
- [ ] 交易账户有足够的 USDT
- [ ] 配置文件正确填写
- [ ] 已设置合理的杠杆和风险参数
- [ ] 已进行小额测试

---

## 🎉 开始交易！

配置完成后，启动 AI-Trader：

```bash
./ai-trader
```

查看日志确认 OKX 连接成功：
```
🏦 [OKX量化交易员] 使用OKX交易
✓ 已获取OKX账户余额: 总权益=1000.00 USDT
✓ 已获取OKX持仓信息: 0个持仓
🚀 交易员启动成功
```

**祝你交易顺利！** 📈
