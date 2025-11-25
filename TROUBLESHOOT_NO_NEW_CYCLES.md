# 决策页面没有新周期数据的排查指南

## 🔍 问题描述

决策页面上没有了 3 分钟一次的新数据产生，已经过了很久都没有新的周期数据。

## 📋 可能的原因

### **1. Trader 已停止运行** ⭐⭐⭐⭐⭐ (最常见)

**原因：**
- 用户手动停止了 Trader
- 服务器重启后 Trader 未自动启动（如果 `is_running = false`）
- Trader 运行时遇到错误自动退出

**检查方法：**

#### **方法 A: 通过数据库检查**

```sql
-- 检查 Trader 状态
SELECT 
    id,
    name,
    is_running,
    scan_interval_minutes,
    updated_at
FROM traders
ORDER BY updated_at DESC;
```

**预期结果：**
- 如果 `is_running = false` → Trader 已停止
- 如果 `is_running = true` → Trader 应该在运行

#### **方法 B: 通过 API 检查**

```bash
curl http://localhost:8080/api/traders/your_trader_id \
  -H "Authorization: Bearer your_token"
```

**查看响应中的 `is_running` 字段：**
```json
{
  "id": "trader_123",
  "name": "My Trader",
  "is_running": false,  // ❌ 如果是 false，说明已停止
  ...
}
```

#### **方法 C: 通过日志检查**

```bash
# 查看最近的日志
tail -50 logs/trader_*.log

# 查找停止信号
grep -i "停止\|stopped\|⏹" logs/trader_*.log | tail -5
```

**如果看到：**
```
⏹ [My Trader] Trader已停止，跳过本次决策周期
⏹ 自动交易系统停止
```

说明 Trader 已停止。

**解决方案：**
1. 在页面上点击 "启动" 按钮
2. 或通过 API 启动：
   ```bash
   curl -X POST http://localhost:8080/api/traders/your_trader_id/start \
     -H "Authorization: Bearer your_token"
   ```

---

### **2. 风险控制暂停** ⭐⭐⭐⭐

**原因：**
- 触发了最大日亏损限制
- 触发了最大回撤限制
- 系统自动暂停交易

**检查方法：**

```bash
# 查找风险控制日志
grep "风险控制" logs/trader_*.log | tail -10
```

**如果看到：**
```
⏸ 风险控制：暂停交易中，剩余 45 分钟
```

说明触发了风险控制暂停。

**特征：**
- Trader 状态仍然是 `is_running = true`
- 但不会执行新的决策
- 每个周期会记录一条 "暂停交易中" 的日志
- 暂停时间结束后会自动恢复

**解决方案：**
- 等待暂停时间结束（会自动恢复）
- 或者调整风险控制参数（在用户配置中）

---

### **3. 程序崩溃或卡住** ⭐⭐⭐

**原因：**
- AI API 调用超时（默认 180 秒）
- 数据库连接失败
- 交易所 API 调用失败
- 内存不足导致程序崩溃

**检查方法：**

#### **检查进程是否运行**

```bash
# 检查进程
ps aux | grep nofx

# 或
systemctl status nofx.service

# 或
pm2 list
```

**如果进程不存在：**
→ 程序已崩溃，需要重启

#### **检查日志中的错误**

```bash
# 查找错误日志
grep -i "error\|failed\|失败\|panic\|timeout" logs/trader_*.log | tail -20

# 查看最后的日志
tail -100 logs/trader_*.log
```

**常见错误：**

**A. AI API 超时**
```
❌ 调用AI API失败: context deadline exceeded (Client.Timeout exceeded)
```

**B. 数据库连接失败**
```
❌ 构建交易上下文失败: 获取账户余额失败: database connection failed
```

**C. 交易所 API 失败**
```
❌ 获取账户余额失败: API key invalid
```

**解决方案：**
1. 重启服务：
   ```bash
   systemctl restart nofx.service
   # 或
   pm2 restart nofx
   ```

2. 检查 API 密钥是否有效
3. 检查网络连接
4. 检查数据库是否正常运行

---

### **4. 数据库连接问题** ⭐⭐⭐

**原因：**
- PostgreSQL 数据库宕机
- 数据库连接池耗尽
- 网络连接中断

**检查方法：**

```bash
# 检查 PostgreSQL 状态
systemctl status postgresql

# 测试数据库连接
psql -h localhost -U your_user -d nofx_db -c "SELECT 1;"

# 查看数据库日志
tail -50 /var/log/postgresql/postgresql-*.log
```

**解决方案：**
1. 重启 PostgreSQL：
   ```bash
   systemctl restart postgresql
   ```

2. 检查数据库配置
3. 增加连接池大小

---

### **5. 决策周期正常但前端未刷新** ⭐⭐

**原因：**
- 前端轮询停止
- WebSocket 连接断开
- 浏览器缓存问题

**检查方法：**

#### **A. 检查数据库中是否有新记录**

```sql
-- 查看最近的决策记录
SELECT 
    cycle_number,
    timestamp,
    success,
    error_message
FROM decision_logs
WHERE trader_id = 'your_trader_id'
ORDER BY cycle_number DESC
LIMIT 10;
```

**如果有新记录：**
→ 后端正常，问题在前端

**如果没有新记录：**
→ 后端有问题，继续排查

#### **B. 检查浏览器控制台**

1. 打开浏览器开发者工具（F12）
2. 查看 Console 标签页
3. 查找错误信息

**常见错误：**
```
Failed to fetch
Network request failed
401 Unauthorized
```

**解决方案：**
1. 刷新页面（Ctrl+F5 强制刷新）
2. 清除浏览器缓存
3. 重新登录
4. 检查网络连接

---

### **6. 扫描间隔设置过长** ⭐

**原因：**
- 扫描间隔被设置为很长时间（例如 60 分钟）
- 看起来像是没有新数据，实际上是间隔太长

**检查方法：**

```sql
-- 检查扫描间隔
SELECT 
    id,
    name,
    scan_interval_minutes
FROM traders;
```

**如果 `scan_interval_minutes > 10`：**
→ 间隔可能太长

**解决方案：**
1. 在页面上修改扫描间隔为 3 分钟
2. 或通过 API 更新：
   ```bash
   curl -X PUT http://localhost:8080/api/traders/your_trader_id \
     -H "Authorization: Bearer your_token" \
     -H "Content-Type: application/json" \
     -d '{"scan_interval_minutes": 3}'
   ```

---

## 🔧 快速诊断工具

我已经创建了一个诊断脚本，可以自动检查大部分问题：

```bash
# 运行诊断脚本
./diagnose_trader.sh
```

**脚本会检查：**
1. ✅ 进程是否运行
2. ✅ 数据库中 Trader 的状态
3. ✅ 最近的日志
4. ✅ 最近的决策记录
5. ✅ 是否有风险控制暂停
6. ✅ API 服务是否正常
7. ✅ 提供诊断建议

---

## 📝 逐步排查流程

### **Step 1: 检查 Trader 状态**

```sql
SELECT id, name, is_running, scan_interval_minutes 
FROM traders;
```

**如果 `is_running = false`：**
→ 在页面上启动 Trader
→ 问题解决 ✅

**如果 `is_running = true`：**
→ 继续 Step 2

---

### **Step 2: 检查最近的日志**

```bash
tail -100 logs/trader_*.log
```

**查找关键信息：**
- 最后一条日志的时间
- 是否有错误信息
- 是否有 "AI决策周期" 日志

**如果最后日志时间 < 5 分钟：**
→ Trader 正在运行，继续 Step 3

**如果最后日志时间 > 10 分钟：**
→ Trader 可能卡住或崩溃，重启服务
→ `systemctl restart nofx.service`

---

### **Step 3: 检查决策记录**

```sql
SELECT 
    cycle_number,
    timestamp,
    success,
    error_message
FROM decision_logs
WHERE trader_id = 'your_trader_id'
ORDER BY cycle_number DESC
LIMIT 5;
```

**如果有新记录（< 5 分钟）：**
→ 后端正常，问题在前端
→ 刷新页面或清除缓存

**如果没有新记录：**
→ 后端有问题，继续 Step 4

---

### **Step 4: 检查错误日志**

```bash
grep -i "error\|failed\|失败" logs/trader_*.log | tail -20
```

**常见错误及解决方案：**

| 错误信息 | 原因 | 解决方案 |
|---------|------|---------|
| `Trader已停止，跳过本次决策周期` | Trader 已停止 | 在页面上启动 |
| `风险控制：暂停交易中` | 触发风险控制 | 等待暂停结束 |
| `调用AI API失败` | AI API 问题 | 检查 API 密钥和网络 |
| `获取账户余额失败` | 交易所 API 问题 | 检查交易所 API 密钥 |
| `数据库连接失败` | 数据库问题 | 重启 PostgreSQL |
| `context deadline exceeded` | 超时 | 增加超时时间或检查网络 |

---

### **Step 5: 检查进程状态**

```bash
# 检查进程
ps aux | grep nofx

# 检查端口
netstat -tlnp | grep 8080

# 检查系统日志
journalctl -u nofx.service -n 50
```

**如果进程不存在：**
→ 重启服务
→ `systemctl start nofx.service`

---

### **Step 6: 检查前端**

1. 打开浏览器开发者工具（F12）
2. 查看 Console 标签页
3. 查看 Network 标签页

**检查：**
- 是否有 API 请求错误
- Token 是否过期
- 网络连接是否正常

**解决方案：**
- 刷新页面（Ctrl+F5）
- 重新登录
- 清除浏览器缓存

---

## 🎯 最常见的 3 个原因及解决方案

### **1. Trader 已停止 (70%)**

**检查：**
```sql
SELECT is_running FROM traders WHERE id = 'your_trader_id';
```

**解决：**
在页面上点击 "启动" 按钮

---

### **2. 程序崩溃 (20%)**

**检查：**
```bash
ps aux | grep nofx
```

**解决：**
```bash
systemctl restart nofx.service
```

---

### **3. 风险控制暂停 (10%)**

**检查：**
```bash
grep "风险控制" logs/trader_*.log | tail -5
```

**解决：**
等待暂停时间结束（会自动恢复）

---

## 🚀 快速恢复命令

如果不确定问题，可以按顺序执行以下命令：

```bash
# 1. 检查进程
ps aux | grep nofx

# 2. 检查日志
tail -50 logs/trader_*.log

# 3. 如果进程不存在，重启服务
systemctl restart nofx.service
# 或
pm2 restart nofx

# 4. 等待 30 秒后检查日志
sleep 30
tail -50 logs/trader_*.log

# 5. 如果仍然没有日志，检查数据库
psql -c "SELECT id, name, is_running FROM traders;"

# 6. 如果 is_running = false，通过 API 启动
curl -X POST http://localhost:8080/api/traders/your_trader_id/start \
  -H "Authorization: Bearer your_token"
```

---

## 📊 监控建议

为了避免将来再次出现这个问题，建议添加监控：

### **1. 定时检查脚本**

创建一个 cron 任务，每 10 分钟检查一次：

```bash
# /etc/cron.d/check-trader
*/10 * * * * /path/to/nofx/diagnose_trader.sh >> /var/log/trader-check.log 2>&1
```

### **2. Telegram 告警**

在代码中添加告警（如果 5 分钟内没有新决策）：

```go
// 在某个定时任务中
lastDecisionTime := getLastDecisionTime(traderID)
if time.Since(lastDecisionTime) > 5*time.Minute {
    message := fmt.Sprintf("⚠️ Trader %s 超过 5 分钟未产生新决策", traderName)
    logger.SendTelegramMessage(message)
}
```

### **3. 健康检查端点**

添加一个健康检查 API：

```go
// GET /api/health
func (s *Server) handleHealth(c *gin.Context) {
    traders := s.traderManager.ListTraders()
    
    for _, trader := range traders {
        status := trader.GetStatus()
        if isRunning, ok := status["is_running"].(bool); ok && isRunning {
            // 检查最后决策时间
            lastDecisionTime := getLastDecisionTime(trader.GetID())
            if time.Since(lastDecisionTime) > 10*time.Minute {
                c.JSON(http.StatusServiceUnavailable, gin.H{
                    "status": "unhealthy",
                    "reason": "No new decisions in 10 minutes",
                })
                return
            }
        }
    }
    
    c.JSON(http.StatusOK, gin.H{"status": "healthy"})
}
```

---

## 🔍 调试技巧

### **实时监控日志**

```bash
# 实时查看所有日志
tail -f logs/trader_*.log

# 只看决策周期
tail -f logs/trader_*.log | grep "AI决策周期"

# 只看错误
tail -f logs/trader_*.log | grep -i "error\|failed\|失败"
```

### **查看特定时间段的日志**

```bash
# 查看最近 1 小时的日志
find logs -name "trader_*.log" -mmin -60 -exec tail -100 {} \;

# 查看今天的所有决策周期
grep "AI决策周期" logs/trader_*.log | grep "$(date +%Y-%m-%d)"
```

### **数据库查询技巧**

```sql
-- 查看每小时的决策数量
SELECT 
    DATE_TRUNC('hour', timestamp) as hour,
    COUNT(*) as decision_count
FROM decision_logs
WHERE trader_id = 'your_trader_id'
  AND timestamp > NOW() - INTERVAL '24 hours'
GROUP BY hour
ORDER BY hour DESC;

-- 如果某个小时的 decision_count = 0，说明那个时间段没有决策
```

---

## 📞 获取帮助

如果以上方法都无法解决问题，请提供以下信息：

1. **Trader 状态**
   ```sql
   SELECT * FROM traders WHERE id = 'your_trader_id';
   ```

2. **最近的日志**
   ```bash
   tail -200 logs/trader_*.log
   ```

3. **最近的决策记录**
   ```sql
   SELECT * FROM decision_logs 
   WHERE trader_id = 'your_trader_id' 
   ORDER BY cycle_number DESC 
   LIMIT 10;
   ```

4. **系统信息**
   ```bash
   # 操作系统
   uname -a
   
   # 内存使用
   free -h
   
   # 磁盘使用
   df -h
   
   # 进程状态
   ps aux | grep nofx
   ```

---

## ✅ 总结

**最可能的原因（按概率）：**
1. **Trader 已停止** (70%) - 在页面上启动
2. **程序崩溃** (20%) - 重启服务
3. **风险控制暂停** (10%) - 等待恢复

**快速诊断：**
```bash
./diagnose_trader.sh
```

**快速恢复：**
1. 检查 Trader 状态
2. 重启服务
3. 在页面上启动 Trader
4. 查看日志确认恢复

**预防措施：**
1. 添加监控和告警
2. 配置自动重启
3. 定期检查日志
4. 设置健康检查
