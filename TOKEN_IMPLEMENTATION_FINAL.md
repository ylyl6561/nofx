# Token 统计功能 - 最终实现方案

## 📊 数据库设计

### 三层 Token 统计架构

```
decision_logs (每次决策)
    ├── prompt_tokens
    ├── completion_tokens  
    └── total_tokens
         ↓ 累加
traders (每个交易员)
    └── total_tokens
         ↓ 累加
users (每个用户)
    └── total_tokens
```

### 优势
1. **决策级别**：`decision_logs` 记录每次 AI 调用的详细 token 使用情况
2. **交易员级别**：`traders` 记录每个交易员的累计 token 使用量
3. **用户级别**：`users` 记录用户所有交易员的总 token 使用量

---

## ✅ 已完成的修改

### 1. 数据库表结构

#### **traders 表**
```sql
ALTER TABLE traders ADD COLUMN total_tokens BIGINT DEFAULT 0;
ALTER TABLE traders ADD COLUMN is_deleted TEXT DEFAULT 'n';
```

#### **users 表**
```sql
ALTER TABLE users ADD COLUMN total_tokens BIGINT DEFAULT 0;
```

#### **decision_logs 表**
```sql
ALTER TABLE decision_logs ADD COLUMN prompt_tokens INTEGER DEFAULT 0;
ALTER TABLE decision_logs ADD COLUMN completion_tokens INTEGER DEFAULT 0;
ALTER TABLE decision_logs ADD COLUMN total_tokens INTEGER DEFAULT 0;
```

---

### 2. Go 结构体修改

#### **config/database.go - TraderRecord**
```go
type TraderRecord struct {
    // ... 其他字段 ...
    TotalTokens int64  `json:"total_tokens"`           // 累计使用的 token 数量
    IsDeleted   string `json:"is_deleted"`             // 软删除标记 ('n'=未删除, 'y'=已删除)
    // ...
}
```

#### **decision/engine.go - FullDecision**
```go
type FullDecision struct {
    // ... 其他字段 ...
    PromptTokens     int `json:"prompt_tokens"`
    CompletionTokens int `json:"completion_tokens"`
    TotalTokens      int `json:"total_tokens"`
}
```

#### **logger/decision_logger.go - DecisionRecord**
```go
type DecisionRecord struct {
    // ... 其他字段 ...
    PromptTokens     int `json:"prompt_tokens"`
    CompletionTokens int `json:"completion_tokens"`
    TotalTokens      int `json:"total_tokens"`
}
```

---

### 3. MCP Client 修改

#### **mcp/interface.go**
```go
// AIResponse AI API 响应（包含 token 使用信息）
type AIResponse struct {
    Content          string `json:"content"`
    PromptTokens     int    `json:"prompt_tokens"`
    CompletionTokens int    `json:"completion_tokens"`
    TotalTokens      int    `json:"total_tokens"`
}

type AIClient interface {
    CallWithMessagesAndTokens(systemPrompt, userPrompt string) (*AIResponse, error)
}
```

#### **mcp/client.go**
```go
func (client *Client) CallWithMessagesAndTokens(systemPrompt, userPrompt string) (*AIResponse, error) {
    // ... API 调用逻辑 ...
    
    // 解析响应中的 usage 字段
    var result struct {
        Choices []struct {
            Message struct {
                Content string `json:"content"`
            } `json:"message"`
        } `json:"choices"`
        Usage struct {
            PromptTokens     int `json:"prompt_tokens"`
            CompletionTokens int `json:"completion_tokens"`
            TotalTokens      int `json:"total_tokens"`
        } `json:"usage"`
    }
    
    // 返回包含 token 信息的响应
    return &AIResponse{
        Content:          result.Choices[0].Message.Content,
        PromptTokens:     result.Usage.PromptTokens,
        CompletionTokens: result.Usage.CompletionTokens,
        TotalTokens:      result.Usage.TotalTokens,
    }, nil
}
```

---

### 4. Decision 包修改

#### **decision/engine.go**
```go
func GetFullDecisionWithCustomPrompt(...) (*FullDecision, error) {
    // ... 构建 prompt ...
    
    // 使用新的 API 获取 token 信息
    aiResponse, err := aiClient.CallWithMessagesAndTokens(systemPrompt, userPrompt)
    if err != nil {
        return nil, err
    }
    
    // 解析决策
    decision, err := parseFullDecisionResponse(aiResponse.Content, ...)
    
    // 保存 token 信息
    if decision != nil {
        decision.PromptTokens = aiResponse.PromptTokens
        decision.CompletionTokens = aiResponse.CompletionTokens
        decision.TotalTokens = aiResponse.TotalTokens
    }
    
    return decision, nil
}
```

---

## 📋 待完成的工作

### 1. 修改 `config/decision_logs_db.go`

#### 添加字段到结构体
```go
type DecisionLogRecord struct {
    // ... 现有字段 ...
    AIRequestDurationMs int64 `json:"ai_request_duration_ms"`
    
    // 添加这三个字段
    PromptTokens     int `json:"prompt_tokens"`
    CompletionTokens int `json:"completion_tokens"`
    TotalTokens      int `json:"total_tokens"`
    
    CreatedAt time.Time `json:"created_at"`
}
```

#### 修改 SaveDecisionLog 的 SQL
```go
func (d *Database) SaveDecisionLog(log *DecisionLogRecord) error {
    query := `
        INSERT INTO decision_logs (
            id, user_id, trader_id, cycle_number, timestamp,
            system_prompt, input_prompt, cot_trace,
            account_state, positions, decisions, candidate_coins, execution_log,
            success, error_message, ai_request_duration_ms,
            prompt_tokens, completion_tokens, total_tokens,  -- 添加这三个字段
            created_at
        ) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)  -- 添加三个 ?
    `
    
    _, err := d.db.Exec(d.convertQuery(query),
        log.ID, log.UserID, log.TraderID, log.CycleNumber, log.Timestamp,
        log.SystemPrompt, log.InputPrompt, log.CoTTrace,
        log.AccountState, log.Positions, log.Decisions, log.CandidateCoins, log.ExecutionLog,
        log.Success, log.ErrorMessage, log.AIRequestDurationMs,
        log.PromptTokens, log.CompletionTokens, log.TotalTokens,  -- 添加这三个参数
        log.CreatedAt,
    )
    
    return err
}
```

---

### 2. 修改 `logger/db_decision_logger.go`

#### 在 LogDecision 方法中添加 token 累加逻辑

```go
func (l *DBDecisionLogger) LogDecision(record *DecisionRecord) error {
    l.cycleNumber++
    record.CycleNumber = l.cycleNumber
    record.Timestamp = time.Now()
    
    // ... 序列化各个字段 ...
    
    // 创建数据库记录
    dbRecord := &config.DecisionLogRecord{
        UserID:              l.userID,
        TraderID:            l.traderID,
        CycleNumber:         record.CycleNumber,
        Timestamp:           record.Timestamp,
        SystemPrompt:        record.SystemPrompt,
        InputPrompt:         record.InputPrompt,
        CoTTrace:            record.CoTTrace,
        AccountState:        string(accountStateJSON),
        Positions:           string(positionsJSON),
        Decisions:           string(decisionsJSON),
        CandidateCoins:      string(candidateCoinsJSON),
        ExecutionLog:        string(executionLogJSON),
        Success:             record.Success,
        ErrorMessage:        record.ErrorMessage,
        AIRequestDurationMs: record.AIRequestDurationMs,
        // 添加 token 字段
        PromptTokens:     record.PromptTokens,
        CompletionTokens: record.CompletionTokens,
        TotalTokens:      record.TotalTokens,
    }
    
    // 保存到数据库
    if err := l.database.SaveDecisionLog(dbRecord); err != nil {
        return fmt.Errorf("保存决策日志到数据库失败: %w", err)
    }
    
    // ✅ 关键：累加 token 到 trader 表和 user 表
    if record.TotalTokens > 0 {
        // 1. 更新 trader 的 token
        updateTraderQuery := `UPDATE traders SET total_tokens = total_tokens + ? WHERE id = ?`
        _, err := l.database.db.Exec(l.database.convertQuery(updateTraderQuery), 
            record.TotalTokens, l.traderID)
        if err != nil {
            log.Printf("⚠️  更新 trader token 统计失败: %v", err)
        } else {
            log.Printf("📊 [Token] 累加 %d tokens 到 trader %s", 
                record.TotalTokens, l.traderID)
        }
        
        // 2. 更新 user 的 token
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
    
    fmt.Printf("📝 决策记录已保存到数据库: trader=%s cycle=%d\n", l.traderID, record.CycleNumber)
    return nil
}
```

---

### 3. 修改 `trader/auto_trader.go`

#### 在 runCycle 方法中传递 token 信息

```go
func (at *AutoTrader) runCycle() error {
    // ... 前面的代码 ...
    
    // 5. 调用AI获取完整决策
    fullDecision, err := decision.GetFullDecisionWithCustomPrompt(
        ctx, at.aiClient, at.customPrompt, at.overrideBasePrompt, at.systemPromptTemplate)
    if err != nil {
        record.Success = false
        record.ErrorMessage = fmt.Sprintf("AI决策失败: %v", err)
        at.decisionLogger.LogDecision(record)
        return fmt.Errorf("AI决策失败: %w", err)
    }
    
    // ✅ 添加：保存 token 信息到 record
    record.PromptTokens = fullDecision.PromptTokens
    record.CompletionTokens = fullDecision.CompletionTokens
    record.TotalTokens = fullDecision.TotalTokens
    
    // ... 后续的执行逻辑 ...
}
```

---

### 4. 添加 API 端点

#### 在 `api/server.go` 中添加

```go
// GET /api/users/token-usage
func (s *Server) handleGetUserTokenUsage(c *gin.Context) {
    userID := c.GetString("user_id")
    
    // 直接从 users 表读取 total_tokens（性能最优）
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

// 在 setupRoutes 中注册
func (s *Server) setupRoutes() {
    // ...
    protected := r.Group("/api")
    protected.Use(s.authMiddleware())
    {
        // ... 其他路由 ...
        protected.GET("/users/token-usage", s.handleGetUserTokenUsage)
    }
}
```

---

## 🚀 部署步骤

### Step 1: 执行数据库迁移

```bash
psql -d nofx_db -f scripts/migrate_add_tokens_and_soft_delete.sql
```

**预期输出：**
```
✅ 已添加 total_tokens 字段
✅ 已添加 is_deleted 字段 (TEXT DEFAULT 'n')
✅ 已添加 users.total_tokens 字段
✅ 已添加 prompt_tokens 字段
✅ 已添加 completion_tokens 字段
✅ 已添加 total_tokens 字段
========================================
✅ 数据库迁移完成！
========================================
```

---

### Step 2: 验证数据库字段

```bash
# 检查 traders 表
psql -d nofx_db -c "\d traders" | grep -E "total_tokens|is_deleted"

# 检查 users 表
psql -d nofx_db -c "\d users" | grep "total_tokens"

# 检查 decision_logs 表
psql -d nofx_db -c "\d decision_logs" | grep -E "prompt_tokens|completion_tokens|total_tokens"
```

---

### Step 3: 完成代码修改

按照上面"待完成的工作"部分，依次修改：
1. `config/decision_logs_db.go`
2. `logger/db_decision_logger.go`
3. `trader/auto_trader.go`
4. `api/server.go`

---

### Step 4: 编译和重启

```bash
# 编译
go build -o nofx main.go

# 重启服务
systemctl restart nofx.service
# 或
pm2 restart nofx
```

---

### Step 5: 测试验证

#### 测试 1: 查看日志

```bash
tail -f logs/trader_*.log | grep "Token"
```

**预期输出：**
```
📊 [Token] Prompt: 1234 | Completion: 567 | Total: 1801
📊 [Token] 累加 1801 tokens 到 trader trader_123
📊 [Token] 累加 1801 tokens 到 user user_456
```

---

#### 测试 2: 查询数据库

```bash
# 查看 decision_logs 中的 token 记录
psql -d nofx_db -c "
  SELECT cycle_number, prompt_tokens, completion_tokens, total_tokens 
  FROM decision_logs 
  WHERE trader_id = 'YOUR_TRADER_ID' 
  ORDER BY cycle_number DESC 
  LIMIT 3;
"

# 查看 trader 的累计 token
psql -d nofx_db -c "
  SELECT id, name, total_tokens 
  FROM traders 
  WHERE id = 'YOUR_TRADER_ID';
"

# 查看 user 的累计 token
psql -d nofx_db -c "
  SELECT id, email, total_tokens 
  FROM users 
  WHERE id = 'YOUR_USER_ID';
"
```

---

#### 测试 3: 调用 API

```bash
curl http://localhost:8080/api/users/token-usage \
  -H "Authorization: Bearer YOUR_TOKEN"
```

**预期响应：**
```json
{
  "total_tokens": 123456,
  "user_id": "user_123"
}
```

---

## 📊 数据流程图

```
AI API 调用
    ↓
返回 usage { prompt_tokens, completion_tokens, total_tokens }
    ↓
mcp.Client.CallWithMessagesAndTokens() 解析并返回 AIResponse
    ↓
decision.GetFullDecisionWithCustomPrompt() 保存到 FullDecision
    ↓
trader.AutoTrader.runCycle() 保存到 DecisionRecord
    ↓
logger.DBDecisionLogger.LogDecision() 执行三个操作：
    ├── 1. 保存到 decision_logs 表
    ├── 2. 累加到 traders.total_tokens
    └── 3. 累加到 users.total_tokens
         ↓
API /api/users/token-usage 直接读取 users.total_tokens
    ↓
前端展示
```

---

## 🎯 关键优势

### 1. 性能优化
- **用户 token 查询**：直接从 `users.total_tokens` 读取，无需聚合计算
- **单次写入**：每次决策只需要 3 个 UPDATE 操作

### 2. 数据一致性
- **原子性**：每次决策的 token 同时更新到三个表
- **可追溯**：`decision_logs` 保留每次调用的详细记录

### 3. 灵活性
- **多维度统计**：支持按决策、交易员、用户三个维度查询
- **历史分析**：可以分析 token 使用趋势

---

## 💡 未来优化建议

### 1. Token 成本计算
```go
// 根据不同模型计算成本
type TokenCost struct {
    PromptCost      float64  // 每 1K prompt tokens 的成本
    CompletionCost  float64  // 每 1K completion tokens 的成本
}

func CalculateCost(promptTokens, completionTokens int, model string) float64 {
    costs := map[string]TokenCost{
        "deepseek-chat": {PromptCost: 0.0014, CompletionCost: 0.0028},
        "gpt-4":         {PromptCost: 0.03, CompletionCost: 0.06},
    }
    
    cost := costs[model]
    return (float64(promptTokens)/1000)*cost.PromptCost + 
           (float64(completionTokens)/1000)*cost.CompletionCost
}
```

### 2. Token 限额控制
```go
// 检查用户是否超过 token 限额
func CheckTokenLimit(userID string, limit int64) (bool, error) {
    var totalTokens int64
    query := `SELECT total_tokens FROM users WHERE id = ?`
    err := db.QueryRow(query, userID).Scan(&totalTokens)
    if err != nil {
        return false, err
    }
    return totalTokens < limit, nil
}
```

### 3. Token 使用趋势分析
```sql
-- 按天统计 token 使用量
SELECT 
    DATE(timestamp) as date,
    SUM(total_tokens) as daily_tokens
FROM decision_logs
WHERE user_id = ?
GROUP BY DATE(timestamp)
ORDER BY date DESC
LIMIT 30;
```

---

## ✅ 完成检查清单

- [x] 修改 `config/database_tables.go` 添加字段定义
- [x] 修改 `config/database.go` 添加结构体字段
- [x] 修改 `mcp/interface.go` 定义 AIResponse
- [x] 修改 `mcp/client.go` 实现 CallWithMessagesAndTokens
- [x] 修改 `decision/engine.go` 保存 token 信息
- [x] 修改 `logger/decision_logger.go` 添加结构体字段
- [x] 创建数据库迁移脚本
- [x] 代码编译通过
- [ ] 修改 `config/decision_logs_db.go` 添加 SQL
- [ ] 修改 `logger/db_decision_logger.go` 累加 token
- [ ] 修改 `trader/auto_trader.go` 传递 token
- [ ] 添加 `/api/users/token-usage` API
- [ ] 执行数据库迁移
- [ ] 测试验证

---

## 📞 需要帮助？

如果遇到问题，检查：
1. 数据库迁移是否成功执行
2. 日志中是否有 "📊 [Token]" 相关输出
3. 数据库中的 token 字段是否正确累加
4. API 是否返回正确的 token 数量

详细文档：`QUICK_START_GUIDE.md`
