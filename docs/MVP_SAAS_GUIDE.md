# 🎯 AI-Trader MVP SaaS 改造指南

## 📋 概述

本指南介绍如何将 AI-Trader 改造为支持 **双模式访问** 的 SaaS 平台：
- **模式 1**：Web 登录（保留现有用户体验）
- **模式 2**：API Key 调用（新增开发者接口）

### 设计原则
✅ **最小改动** - 不影响现有功能  
✅ **向后兼容** - 现有用户无感知  
✅ **易于扩展** - 为 Phase 2 做准备  

---

## 🗂️ 新增文件清单

### 后端文件
```
nofx/
├── config/
│   └── api_keys.go                    # API Key 管理（新增）
├── api/
│   ├── middleware/
│   │   └── auth.go                    # API Key 认证中间件（新增）
│   └── handlers/
│       ├── api_keys.go                # API Key 管理接口（新增）
│       └── trading_api.go             # 交易 API 接口（新增）
└── web/
    └── src/
        └── components/
            └── APIKeysPage.tsx        # API Keys 管理页面（新增）
```

### 修改的文件
```
nofx/
├── config/
│   └── database.go                    # 新增 2 个表
└── api/
    └── server.go                      # 新增路由
```

---

## 🗄️ 数据库变更

### 新增表

#### 1. `api_keys` 表
```sql
CREATE TABLE IF NOT EXISTS api_keys (
    id TEXT PRIMARY KEY,
    user_id TEXT NOT NULL,
    key_hash TEXT NOT NULL UNIQUE,
    key_prefix TEXT NOT NULL,
    name TEXT DEFAULT 'Default Key',
    enabled BOOLEAN DEFAULT 1,
    rate_limit INTEGER DEFAULT 1000,
    usage_count INTEGER DEFAULT 0,
    last_used_at DATETIME,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    expires_at DATETIME,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);
```

#### 2. `api_usage_logs` 表
```sql
CREATE TABLE IF NOT EXISTS api_usage_logs (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    api_key_id TEXT NOT NULL,
    user_id TEXT NOT NULL,
    endpoint TEXT NOT NULL,
    method TEXT NOT NULL,
    status_code INTEGER,
    response_time_ms INTEGER,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (api_key_id) REFERENCES api_keys(id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);
```

---

## 🔌 API 端点

### Web 认证路由（需要 JWT）

```
GET    /api/api-keys              # 列出 API Keys
POST   /api/api-keys              # 创建 API Key
DELETE /api/api-keys/:id          # 删除 API Key
POST   /api/api-keys/:id/revoke   # 撤销 API Key
GET    /api/api-keys/usage        # 使用统计
```

### RESTful API（需要 API Key）

```
GET    /api/v1/traders            # 列出交易员
GET    /api/v1/traders/:id        # 获取交易员详情
POST   /api/v1/traders/:id/start  # 启动交易员
POST   /api/v1/traders/:id/stop   # 停止交易员
GET    /api/v1/traders/:id/performance  # 获取绩效
GET    /api/v1/models             # 列出 AI 模型
GET    /api/v1/exchanges          # 列出交易所
```

---

## 🚀 部署步骤

### 1. 编译后端

```bash
cd nofx

# 构建
go build -o nofx .

# 或使用 Docker
docker compose build nofx
```

### 2. 数据库自动迁移

启动时会自动创建新表，无需手动操作。

### 3. 启动服务

```bash
# 本地启动
./nofx

# 或使用 Docker
docker compose up -d nofx
```

### 4. 部署到 Zeabur

```bash
# 1. 推送代码到 GitHub
git add .
git commit -m "feat: add API key support"
git push

# 2. 在 Zeabur 连接仓库并部署
# 3. 添加 Volume: /app (1GB)
# 4. 添加环境变量（见下方）
```

---

## ⚙️ 环境变量

```bash
# 必需
AI_MAX_TOKENS=4000
TZ=Asia/Shanghai
AI-Trader_ADMIN_PASSWORD=your_secure_password

# 可选
PORT=8080
```

---

## 📖 使用指南

### 方式 1：Web 界面使用

#### 步骤 1：登录
```
访问: https://your-domain.com
登录账号
```

#### 步骤 2：创建 API Key
```
1. 进入 "API Keys" 页面
2. 点击 "Create New API Key"
3. 输入名称和速率限制
4. 保存生成的 Key（只显示一次！）
```

#### 步骤 3：使用 API Key
```bash
curl -X GET https://your-domain.com/api/v1/traders \
  -H "Authorization: Bearer aitrader_your_api_key"
```

---

### 方式 2：纯 API 调用

#### 步骤 1：注册账号（如果需要）
```bash
curl -X POST https://your-domain.com/api/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "your_password"
  }'
```

#### 步骤 2：登录获取 JWT
```bash
curl -X POST https://your-domain.com/api/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "your_password"
  }'
```

#### 步骤 3：创建 API Key
```bash
curl -X POST https://your-domain.com/api/api-keys \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "My API Key",
    "rate_limit": 1000
  }'
```

#### 步骤 4：使用 API Key 调用交易接口
```bash
# 列出交易员
curl -X GET https://your-domain.com/api/v1/traders \
  -H "Authorization: Bearer aitrader_your_api_key"

# 启动交易员
curl -X POST https://your-domain.com/api/v1/traders/trader_001/start \
  -H "Authorization: Bearer aitrader_your_api_key" \
  -H "Content-Type: application/json" \
  -d '{"trader_id": "trader_001"}'

# 获取绩效
curl -X GET https://your-domain.com/api/v1/traders/trader_001/performance \
  -H "Authorization: Bearer aitrader_your_api_key"
```

---

## 🔒 安全特性

### API Key 安全
- ✅ 使用 SHA-256 哈希存储
- ✅ 只在创建时显示完整 Key
- ✅ 支持撤销和删除
- ✅ 支持过期时间

### 速率限制
- ✅ 每个 Key 独立限制
- ✅ 按月重置
- ✅ 超限自动拒绝

### 日志审计
- ✅ 记录所有 API 调用
- ✅ 包含响应时间
- ✅ 支持统计分析

---

## 📊 监控与统计

### 查看使用统计
```bash
curl -X GET https://your-domain.com/api/api-keys/usage \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

响应示例：
```json
{
  "total_calls": 1523,
  "avg_response_time": 245.6,
  "monthly_usage": 1523,
  "period_days": 30
}
```

---

## 🧪 测试

### 测试 API Key 认证
```bash
# 1. 无效的 Key
curl -X GET https://your-domain.com/api/v1/traders \
  -H "Authorization: Bearer invalid_key"
# 应返回 401 Unauthorized

# 2. 有效的 Key
curl -X GET https://your-domain.com/api/v1/traders \
  -H "Authorization: Bearer aitrader_valid_key"
# 应返回交易员列表

# 3. 超出速率限制
# 连续调用超过 rate_limit 次
# 应返回 429 Too Many Requests
```

---

## 🎨 前端集成

### 添加路由（如果使用 React Router）

```typescript
// App.tsx
import APIKeysPage from './components/APIKeysPage';

<Route path="/api-keys" element={<APIKeysPage />} />
```

### 添加导航菜单

```typescript
// Sidebar.tsx
<Link to="/api-keys">
  <Key size={20} />
  API Keys
</Link>
```

---

## 🔄 数据迁移

### 现有用户无需操作
- 现有功能完全保留
- 新表自动创建
- 无数据丢失风险

### 如果需要回滚
```bash
# 删除新增的表
sqlite3 config.db "DROP TABLE IF EXISTS api_keys;"
sqlite3 config.db "DROP TABLE IF EXISTS api_usage_logs;"

# 恢复旧版本代码
git checkout old_version
```

---

## 📈 Phase 2 扩展计划

### 计划功能
- [ ] 套餐管理（Free/Pro/Enterprise）
- [ ] Stripe 支付集成
- [ ] WebSocket 实时推送
- [ ] 更详细的使用统计
- [ ] 多语言 SDK（Python/JavaScript）
- [ ] API 文档（Swagger/OpenAPI）

### 数据库扩展
```sql
-- 订阅管理表
CREATE TABLE subscriptions (
    id TEXT PRIMARY KEY,
    user_id TEXT NOT NULL,
    plan TEXT NOT NULL,
    status TEXT NOT NULL,
    price REAL NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
```

---

## 🐛 常见问题

### Q1: API Key 无法创建？
**A:** 检查用户是否已登录，JWT Token 是否有效。

### Q2: API 调用返回 401？
**A:** 检查 Authorization Header 格式：`Bearer aitrader_xxxxx`

### Q3: 速率限制如何重置？
**A:** 目前按月自动重置，可在 `api_keys` 表手动重置 `usage_count`。

### Q4: 如何查看 API 调用日志？
**A:** 查询 `api_usage_logs` 表：
```sql
SELECT * FROM api_usage_logs 
WHERE user_id = 'your_user_id' 
ORDER BY created_at DESC 
LIMIT 100;
```

---

## 📞 支持

如有问题，请：
1. 查看日志：`docker compose logs -f nofx`
2. 检查数据库：`sqlite3 config.db`
3. 提交 Issue：GitHub Issues

---

## ✅ 检查清单

部署前检查：
- [ ] 代码已编译无错误
- [ ] 数据库表已创建
- [ ] 环境变量已配置
- [ ] Volume 已挂载
- [ ] API 端点可访问
- [ ] Web 界面正常
- [ ] API Key 可创建
- [ ] API 调用成功

---

## 🎉 完成！

现在你的 AI-Trader 已经支持：
- ✅ Web 登录配置交易
- ✅ API Key 调用接口
- ✅ 使用统计和监控
- ✅ 安全的认证机制

**开始验证你的商业模式吧！** 🚀
