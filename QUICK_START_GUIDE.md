# Token 统计和软删除功能 - 快速实施指南

## 🎯 概述

本指南帮助你快速完成 Token 统计和软删除功能的实施。

---

## ✅ 已完成的工作

1. ✅ 数据库表结构修改
2. ✅ MCP Client 返回 token 信息
3. ✅ Decision 包保存 token 信息
4. ✅ Logger 结构体添加 token 字段
5. ✅ 数据库迁移脚本
6. ✅ 代码编译通过

---

## 🚀 立即执行的步骤

### Step 1: 执行数据库迁移（必须）

```bash
# 执行迁移脚本
psql -d nofx_db -f scripts/migrate_add_tokens_and_soft_delete.sql
```

**验证迁移成功：**
```bash
psql -d nofx_db -c "\d traders" | grep -E "total_tokens|is_deleted"
psql -d nofx_db -c "\d decision_logs" | grep -E "prompt_tokens|completion_tokens|total_tokens"
```

---

### Step 2: 重启服务测试基础功能

```bash
# 编译（已通过）
go build -o nofx main.go

# 重启服务
systemctl restart nofx.service
# 或
pm2 restart nofx
```

---

### Step 3: 验证 Token 统计是否工作

```bash
# 1. 启动一个 trader
curl -X POST http://localhost:8080/api/traders/YOUR_TRADER_ID/start \
  -H "Authorization: Bearer YOUR_TOKEN"

# 2. 等待 3-5 分钟（一个决策周期）

# 3. 查看日志
tail -f logs/trader_*.log | grep "Token"

# 预期看到：
# 📊 [Token] Prompt: 1234 | Completion: 567 | Total: 1801
```

如果看到上面的日志，说明 **Token 统计的核心功能已经工作**！

---

## 📋 剩余工作清单

### 必须完成的（核心功能）

#### 1. 修改 `config/decision_logs_db.go`

在 `DecisionLogRecord` 结构体中添加（约第28行）：

```go
AIRequestDurationMs int64     `json:"ai_request_duration_ms"`
// 添加以下三行
PromptTokens     int `json:"prompt_tokens"`
CompletionTokens int `json:"completion_tokens"`
TotalTokens      int `json:"total_tokens"`
CreatedAt        time.Time `json:"created_at"`
```

在 `SaveDecisionLog` 方法的 SQL 中添加（约第45行）：

```go
query := `
    INSERT INTO decision_logs (
        id, user_id, trader_id, cycle_number, timestamp,
        system_prompt, input_prompt, cot_trace,
        account_state, positions, decisions, candidate_coins, execution_log,
        success, error_message, ai_request_duration_ms,
        prompt_tokens, completion_tokens, total_tokens,  -- 添加这一行
        created_at
    ) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)  -- 添加三个 ?
`
```

在 `Exec` 调用中添加参数（约第60行）：

```go
_, err := d.db.Exec(d.convertQuery(query),
    log.ID, log.UserID, log.TraderID, log.CycleNumber, log.Timestamp,
    log.SystemPrompt, log.InputPrompt, log.CoTTrace,
    log.AccountState, log.Positions, log.Decisions, log.CandidateCoins, log.ExecutionLog,
    log.Success, log.ErrorMessage, log.AIRequestDurationMs,
    log.PromptTokens, log.CompletionTokens, log.TotalTokens,  -- 添加这一行
    log.CreatedAt,
)
```

---

#### 2. 修改 `logger/db_decision_logger.go`

在 `LogDecision` 方法中添加（约第77行）：

```go
dbRecord := &config.DecisionLogRecord{
    // ... 现有字段 ...
    AIRequestDurationMs: record.AIRequestDurationMs,
    // 添加以下三行
    PromptTokens:     record.PromptTokens,
    CompletionTokens: record.CompletionTokens,
    TotalTokens:      record.TotalTokens,
}

// ✅ 新增：累加 token 到 trader 表和 user 表
if record.TotalTokens > 0 {
    // 更新 trader 的 token
    updateTraderQuery := `UPDATE traders SET total_tokens = total_tokens + ? WHERE id = ?`
    _, err := l.database.db.Exec(l.database.convertQuery(updateTraderQuery), 
        record.TotalTokens, l.traderID)
    if err != nil {
        log.Printf("⚠️  更新 trader token 统计失败: %v", err)
    } else {
        log.Printf("📊 [Token] 累加 %d tokens 到 trader %s", 
            record.TotalTokens, l.traderID)
    }
    
    // 更新 user 的 token
    updateUserQuery := `UPDATE users SET total_tokens = total_tokens + ? WHERE id = ?`
    _, err = l.database.db.Exec(l.database.convertQuery(updateUserQuery), 
        record.TotalTokens, l.userID)
    if err != nil {
        log.Printf("⚠️  更新 user token 统计失败: %v", err)
    } else {
        log.Printf("📊 [Token] 累加 %d tokens 到 user %s", 
            record.TotalTokens, l.userID)
    }
}
```

---

#### 3. 修改 `trader/auto_trader.go`

在 `runCycle` 方法中，找到调用 AI 的地方（约第485行），添加：

```go
// 5. 调用AI获取完整决策
fullDecision, err := decision.GetFullDecisionWithCustomPrompt(ctx, at.aiClient, at.customPrompt, at.overrideBasePrompt, at.systemPromptTemplate)
if err != nil {
    record.Success = false
    record.ErrorMessage = fmt.Sprintf("AI决策失败: %v", err)
    at.decisionLogger.LogDecision(record)
    return fmt.Errorf("AI决策失败: %w", err)
}

// 添加以下三行
record.PromptTokens = fullDecision.PromptTokens
record.CompletionTokens = fullDecision.CompletionTokens
record.TotalTokens = fullDecision.TotalTokens
```

---

### 可选完成的（增强功能）

#### 4. 添加 Token 汇总 API

在 `api/server.go` 中添加新的 API 端点（约第2500行）：

```go
// GET /api/users/token-usage
func (s *Server) handleGetUserTokenUsage(c *gin.Context) {
    userID := c.GetString("user_id")
    
    // 直接从 users 表读取 total_tokens
    var totalTokens int64
    query := `SELECT COALESCE(total_tokens, 0) FROM users WHERE id = ?`
    err := s.database.db.QueryRow(s.database.convertQuery(query), userID).Scan(&totalTokens)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败"})
        return
    }
    
    c.JSON(http.StatusOK, gin.H{
        "total_tokens": totalTokens,
        "user_id": userID,
    })
}
```

在 `setupRoutes` 方法中注册路由（约第180行）：

```go
protected := r.Group("/api")
protected.Use(s.authMiddleware())
{
    // ... 现有路由 ...
    protected.GET("/users/token-usage", s.handleGetUserTokenUsage)  // 添加这一行
}
```

---

#### 5. 修改删除逻辑为软删除

在 `api/server.go` 的 `handleDeleteTrader` 方法中（约第820行），修改删除逻辑：

```go
// 原来的代码：
// err = s.database.DeleteTrader(userID, traderID)

// 修改为：
query := `UPDATE traders SET is_deleted = 'y', updated_at = CURRENT_TIMESTAMP WHERE id = ? AND user_id = ?`
_, err = s.database.db.Exec(s.database.convertQuery(query), traderID, userID)
if err != nil {
    c.JSON(http.StatusInternalServerError, gin.H{"error": "删除失败"})
    return
}

log.Printf("🗑️  交易员 %s 已软删除", traderID)
```

---

#### 6. 修改查询逻辑过滤已删除的 trader

在 `config/database.go` 中，找到所有查询 traders 的地方，添加 `AND is_deleted = 'n'`：

**GetTraders 方法（约第820行）：**
```go
query := `
    SELECT ... 
    FROM traders
    WHERE user_id = ? AND is_deleted = 'n'  -- 添加这个条件
    ORDER BY created_at DESC
`
```

**GetTraderConfig 方法（约第860行）：**
```go
query := `
    SELECT ... 
    FROM traders
    WHERE id = ? AND user_id = ? AND is_deleted = 'n'  -- 添加这个条件
`
```

**CreateTrader 方法（约第806行）：**
确保插入时 `is_deleted = 'n'`

---

## 🧪 测试步骤

### 测试 1: Token 统计

```bash
# 1. 重新编译
go build -o nofx main.go

# 2. 重启服务
systemctl restart nofx.service

# 3. 启动 trader
curl -X POST http://localhost:8080/api/traders/YOUR_TRADER_ID/start \
  -H "Authorization: Bearer YOUR_TOKEN"

# 4. 等待 3-5 分钟

# 5. 查看数据库
psql -d nofx_db -c "
  SELECT cycle_number, prompt_tokens, completion_tokens, total_tokens 
  FROM decision_logs 
  WHERE trader_id = 'YOUR_TRADER_ID' 
  ORDER BY cycle_number DESC 
  LIMIT 3;
"

# 6. 查看 trader 累计 token
psql -d nofx_db -c "
  SELECT id, name, total_tokens 
  FROM traders 
  WHERE id = 'YOUR_TRADER_ID';
"
```

**预期结果：**
- decision_logs 表中有 token 记录
- traders 表中的 total_tokens 在累加

---

### 测试 2: Token 汇总 API

```bash
curl http://localhost:8080/api/users/token-usage \
  -H "Authorization: Bearer YOUR_TOKEN"
```

**预期响应：**
```json
{
  "total_tokens": 12345,
  "user_id": "user_123"
}
```

---

### 测试 3: 软删除

```bash
# 1. 删除 trader
curl -X DELETE http://localhost:8080/api/traders/YOUR_TRADER_ID \
  -H "Authorization: Bearer YOUR_TOKEN"

# 2. 查看数据库
psql -d nofx_db -c "
  SELECT id, name, is_deleted, total_tokens 
  FROM traders 
  WHERE id = 'YOUR_TRADER_ID';
"

# 3. 验证列表中不显示
curl http://localhost:8080/api/traders \
  -H "Authorization: Bearer YOUR_TOKEN"
```

**预期结果：**
- 数据库中 `is_deleted = 'y'`
- trader 的 total_tokens 仍然保留
- API 列表中不再显示该 trader

---

## 📊 前端实现（可选）

如果需要在前端展示 token 使用量，参考 `IMPLEMENTATION_SUMMARY.md` 中的前端实现部分。

---

## ✅ 完成检查清单

- [ ] 执行数据库迁移
- [ ] 修改 `config/decision_logs_db.go`
- [ ] 修改 `logger/db_decision_logger.go`
- [ ] 修改 `trader/auto_trader.go`
- [ ] 编译通过
- [ ] 重启服务
- [ ] 测试 Token 统计
- [ ] （可选）添加 Token 汇总 API
- [ ] （可选）修改删除逻辑为软删除
- [ ] （可选）修改查询逻辑过滤已删除
- [ ] （可选）前端展示 Token 使用量

---

## 🎯 预计时间

- **必须完成**：约 1-2 小时
- **可选功能**：约 1-2 小时
- **总计**：约 2-4 小时

---

## 📞 需要帮助？

如果遇到问题，检查：
1. 数据库迁移是否成功
2. 编译是否有错误
3. 日志中是否有错误信息
4. 数据库字段是否正确添加

详细文档：
- `IMPLEMENTATION_SUMMARY.md` - 完整实现文档
- `TOKEN_AND_SOFT_DELETE_IMPLEMENTATION.md` - 详细技术文档
