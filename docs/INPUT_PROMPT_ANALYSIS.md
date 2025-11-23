# Input Prompt 数据缺失问题分析

## 🔍 问题描述

`decision_logs` 表中的 `input_prompt` 字段有时候包含详细的市场数据（价格、MACD、RSI等指标），但有时候却很简单，只有基本的账户信息和币种列表，缺少市场数据。

## 📊 Input Prompt 的构建逻辑

### **代码位置**
`decision/engine.go` 第360-460行：`buildUserPrompt(ctx *Context)` 函数

### **完整的 Input Prompt 应该包含**

1. ✅ **系统状态**（总是有）
   ```
   时间: 2025-11-23 23:00:00 | 周期: #5 | 运行: 15分钟
   ```

2. ✅ **BTC市场数据**（如果有 `ctx.MarketDataMap["BTCUSDT"]`）
   ```
   BTC: 98765.43 (1h: +1.23%, 4h: +2.45%) | MACD: 0.1234 | RSI: 65.43
   ```

3. ✅ **账户信息**（总是有）
   ```
   账户: 净值1000.00 | 余额950.00 (95.0%) | 盈亏+5.00% | 保证金15.0% | 持仓2个
   ```

4. ⚠️ **持仓详情 + 市场数据**（如果有持仓 + `ctx.MarketDataMap`）
   ```
   ## 当前持仓
   1. BTCUSDT LONG | 入场价98000 当前价98500 | 数量0.1 | ...
   
   价格: 98500.00 | 1h: +1.2% | 4h: +2.3% | 24h: +5.6%
   MACD: 0.1234 | 信号线: 0.1100 | 柱状图: 0.0134
   RSI(7): 65.43 | RSI(14): 62.15
   成交量: 1234.56 BTC | 24h成交量: 98765.43 BTC
   资金费率: 0.01% | 持仓量: 12345.67 BTC
   ```

5. ⚠️ **候选币种 + 市场数据**（如果有 `ctx.MarketDataMap`）
   ```
   ## 候选币种 (8个)
   
   ### 1. ETHUSDT
   
   价格: 3500.00 | 1h: +0.8% | 4h: +1.5% | 24h: +3.2%
   MACD: 0.0567 | 信号线: 0.0500 | 柱状图: 0.0067
   ...
   ```

6. ✅ **夏普比率**（如果有 `ctx.Performance`）
   ```
   ## 📊 夏普比率: 1.25
   ```

## 🐛 问题根源

### **关键发现：`MarketDataMap` 未被填充**

**文件：** `trader/auto_trader.go` 第738-758行

```go
// 6. 构建上下文
ctx := &decision.Context{
    CurrentTime:     time.Now().Format("2006-01-02 15:04:05"),
    RuntimeMinutes:  int(time.Since(at.startTime).Minutes()),
    CallCount:       at.callCount,
    BTCETHLeverage:  at.config.BTCETHLeverage,
    AltcoinLeverage: at.config.AltcoinLeverage,
    Account: decision.AccountInfo{...},
    Positions:      positionInfos,
    CandidateCoins: candidateCoins,
    Performance:    performance,
    // ❌ 缺少：MarketDataMap 没有被赋值！
}
```

### **MarketDataMap 的定义**

**文件：** `decision/engine.go` 第83行

```go
type Context struct {
    ...
    MarketDataMap   map[string]*market.Data `json:"-"` // 不序列化，但内部使用
    ...
}
```

### **buildUserPrompt 依赖 MarketDataMap**

**文件：** `decision/engine.go` 第369-373行（BTC数据）

```go
// BTC 市场
if btcData, hasBTC := ctx.MarketDataMap["BTCUSDT"]; hasBTC {
    sb.WriteString(fmt.Sprintf("BTC: %.2f (1h: %+.2f%%, 4h: %+.2f%%) | MACD: %.4f | RSI: %.2f\n\n",
        btcData.CurrentPrice, btcData.PriceChange1h, btcData.PriceChange4h,
        btcData.CurrentMACD, btcData.CurrentRSI7))
}
```

**第411-414行（持仓市场数据）**

```go
// 使用FormatMarketData输出完整市场数据
if marketData, ok := ctx.MarketDataMap[pos.Symbol]; ok {
    sb.WriteString(market.Format(marketData))
    sb.WriteString("\n")
}
```

**第424-440行（候选币种市场数据）**

```go
for _, coin := range ctx.CandidateCoins {
    marketData, hasData := ctx.MarketDataMap[coin.Symbol]
    if !hasData {
        continue  // ❌ 如果没有市场数据，跳过这个币种
    }
    displayedCount++
    
    // 使用FormatMarketData输出完整市场数据
    sb.WriteString(fmt.Sprintf("### %d. %s%s\n\n", displayedCount, coin.Symbol, sourceTags))
    sb.WriteString(market.Format(marketData))
    sb.WriteString("\n")
}
```

## 📋 两种情况对比

### **情况1：MarketDataMap 为空或 nil**

**Input Prompt 内容：**
```
时间: 2025-11-23 23:00:00 | 周期: #5 | 运行: 15分钟

账户: 净值1000.00 | 余额950.00 (95.0%) | 盈亏+5.00% | 保证金15.0% | 持仓0个

当前持仓: 无

## 候选币种 (0个)

---

现在请分析并输出决策（思维链 + JSON）
```

**特征：**
- ❌ 没有 BTC 市场数据
- ❌ 没有持仓的市场数据
- ❌ 候选币种数量显示为 0（因为没有市场数据就跳过）
- ✅ 只有基本的账户信息

### **情况2：MarketDataMap 已填充**

**Input Prompt 内容：**
```
时间: 2025-11-23 23:00:00 | 周期: #5 | 运行: 15分钟

BTC: 98765.43 (1h: +1.23%, 4h: +2.45%) | MACD: 0.1234 | RSI: 65.43

账户: 净值1000.00 | 余额950.00 (95.0%) | 盈亏+5.00% | 保证金15.0% | 持仓1个

## 当前持仓
1. BTCUSDT LONG | 入场价98000 当前价98500 | 数量0.1 | ...

价格: 98500.00 | 1h: +1.2% | 4h: +2.3% | 24h: +5.6%
MACD: 0.1234 | 信号线: 0.1100 | 柱状图: 0.0134
RSI(7): 65.43 | RSI(14): 62.15
成交量: 1234.56 BTC | 24h成交量: 98765.43 BTC
资金费率: 0.01% | 持仓量: 12345.67 BTC

## 候选币种 (8个)

### 1. ETHUSDT

价格: 3500.00 | 1h: +0.8% | 4h: +1.5% | 24h: +3.2%
MACD: 0.0567 | 信号线: 0.0500 | 柱状图: 0.0067
RSI(7): 58.23 | RSI(14): 55.67
...

### 2. SOLUSDT
...

## 📊 夏普比率: 1.25

---

现在请分析并输出决策（思维链 + JSON）
```

**特征：**
- ✅ 有 BTC 市场数据
- ✅ 持仓包含详细的市场指标
- ✅ 候选币种包含完整的技术分析数据
- ✅ AI 可以基于这些数据做出更精准的决策

## 🔍 为什么会出现两种情况？

### **可能的原因**

1. **市场数据获取失败**
   - API 调用超时
   - 网络问题
   - 交易所限流

2. **市场数据未被传递到 Context**
   - 代码中没有调用市场数据获取函数
   - 获取了但没有赋值给 `ctx.MarketDataMap`

3. **条件判断导致跳过**
   - 某些条件下不获取市场数据
   - 例如：测试模式、特定配置等

## 🔧 如何定位问题

### **检查步骤**

1. **查看日志**
   ```bash
   # 搜索市场数据相关的日志
   grep "市场数据\|MarketData\|获取行情" logs/*.log
   ```

2. **检查是否有市场数据获取代码**
   ```bash
   # 在 auto_trader.go 中搜索市场数据获取
   grep -n "GetMarketData\|FetchMarketData\|market\.Get" trader/auto_trader.go
   ```

3. **检查 market 包的使用**
   ```bash
   # 查看 market 包是否被调用
   grep -r "market\." trader/auto_trader.go
   ```

## 💡 解决方案

### **方案1：在构建 Context 时获取市场数据**

**位置：** `trader/auto_trader.go` 第738行之前

```go
// 5.5. 获取市场数据
marketDataMap := make(map[string]*market.Data)

// 获取 BTC 数据
if btcData, err := at.marketMonitor.GetMarketData("BTCUSDT"); err == nil {
    marketDataMap["BTCUSDT"] = btcData
}

// 获取所有候选币种的市场数据
for _, coin := range candidateCoins {
    if data, err := at.marketMonitor.GetMarketData(coin.Symbol); err == nil {
        marketDataMap[coin.Symbol] = data
    }
}

// 6. 构建上下文
ctx := &decision.Context{
    CurrentTime:     time.Now().Format("2006-01-02 15:04:05"),
    RuntimeMinutes:  int(time.Since(at.startTime).Minutes()),
    CallCount:       at.callCount,
    BTCETHLeverage:  at.config.BTCETHLeverage,
    AltcoinLeverage: at.config.AltcoinLeverage,
    Account:         decision.AccountInfo{...},
    Positions:       positionInfos,
    CandidateCoins:  candidateCoins,
    MarketDataMap:   marketDataMap,  // ✅ 添加市场数据
    Performance:     performance,
}
```

### **方案2：检查 market.Monitor 是否正常工作**

**检查：**
1. `at.marketMonitor` 是否已初始化
2. WebSocket 连接是否正常
3. 数据是否在实时更新

### **方案3：添加日志和错误处理**

```go
// 获取市场数据时添加日志
log.Printf("🔍 开始获取市场数据，候选币种数量: %d", len(candidateCoins))

marketDataMap := make(map[string]*market.Data)
successCount := 0

for _, coin := range candidateCoins {
    if data, err := at.marketMonitor.GetMarketData(coin.Symbol); err == nil {
        marketDataMap[coin.Symbol] = data
        successCount++
    } else {
        log.Printf("⚠️  获取 %s 市场数据失败: %v", coin.Symbol, err)
    }
}

log.Printf("✅ 成功获取 %d/%d 个币种的市场数据", successCount, len(candidateCoins))
```

## 📊 影响分析

### **对 AI 决策的影响**

| 有市场数据 | 无市场数据 |
|-----------|-----------|
| ✅ AI 可以看到价格趋势 | ❌ AI 只能基于账户状态决策 |
| ✅ AI 可以分析 MACD、RSI 等指标 | ❌ AI 无法进行技术分析 |
| ✅ AI 可以判断市场情绪 | ❌ AI 缺少关键信息 |
| ✅ 决策更精准 | ❌ 决策质量下降 |
| ✅ 可以识别交易机会 | ❌ 可能错过机会 |

### **对系统的影响**

- **Input Prompt 大小差异**
  - 有数据：5-10KB
  - 无数据：1-2KB

- **AI 响应质量**
  - 有数据：详细的思维链和理由
  - 无数据：简单的 hold/wait 决策

- **交易频率**
  - 有数据：正常交易
  - 无数据：大部分时间 wait

## 🎯 建议

1. **立即检查**
   - 查看最近的决策日志
   - 确认是否有市场数据获取代码
   - 检查 `marketMonitor` 的状态

2. **添加监控**
   - 记录每次市场数据获取的成功率
   - 在 Input Prompt 为空时发出警告

3. **优化代码**
   - 确保 `MarketDataMap` 总是被填充
   - 添加重试机制
   - 缓存市场数据以应对临时故障

4. **测试验证**
   - 对比有/无市场数据时的 AI 决策质量
   - 确认修复后所有 Input Prompt 都包含市场数据

## 📝 总结

**问题根源：** `ctx.MarketDataMap` 在构建 Context 时没有被赋值，导致 `buildUserPrompt` 函数无法输出市场数据。

**解决方向：** 在 `trader/auto_trader.go` 的 `buildContext` 或类似函数中，添加市场数据获取和赋值逻辑。

**优先级：** 🔴 高 - 这直接影响 AI 的决策质量和系统的交易效果。
