# 🚀 AI-Trader MVP SaaS 快速启动指南

## 📦 完整改造总结

> 本项目基于开源项目改造，保留 Git 历史以便追溯技术演进。

### 已完成的工作

#### 1. 数据库改造 ✅
- 新增 `api_keys` 表（API Key 管理）
- 新增 `api_usage_logs` 表（使用日志）
- 自动迁移，无需手动操作

#### 2. 后端功能 ✅
- API Key 生成与验证（SHA-256 哈希）
- API Key 认证中间件
- RESTful API 端点（`/api/v1/*`）
- 使用统计和日志记录
- 速率限制（每月重置）

#### 3. API 端点 ✅
**Web 认证路由**（需要 JWT）：
- `GET /api/api-keys` - 列出 API Keys
- `POST /api/api-keys` - 创建 API Key
- `DELETE /api/api-keys/:id` - 删除 API Key
- `GET /api/api-keys/usage` - 使用统计

**RESTful API**（需要 API Key）：
- `GET /api/v1/traders` - 列出交易员
- `GET /api/v1/traders/:id` - 获取详情
- `POST /api/v1/traders/:id/start` - 启动
- `POST /api/v1/traders/:id/stop` - 停止
- `GET /api/v1/traders/:id/performance` - 绩效
- `GET /api/v1/models` - AI 模型列表
- `GET /api/v1/exchanges` - 交易所列表

#### 4. 前端页面 ✅
- API Keys 管理页面（`APIKeysPage.tsx`）
- 使用统计展示
- API Key 创建/删除
- 快速开始文档

---

## 🎯 立即开始

### 方式 1：本地测试

```bash
# 1. 编译后端
cd /Users/yuliang/project/nofx
go build -o nofx .

# 2. 启动服务
./nofx

# 3. 访问
open http://localhost:8080
```

### 方式 2：Docker 部署

```bash
# 1. 构建
docker compose build nofx

# 2. 启动
docker compose up -d nofx

# 3. 查看日志
docker compose logs -f nofx
```

### 方式 3：Zeabur 部署

```bash
# 1. 推送代码
git add .
git commit -m "feat: add API key support for MVP SaaS"
git push

# 2. 在 Zeabur 部署
# - 连接 GitHub 仓库
# - 选择 Dockerfile: docker/Dockerfile.backend
# - 添加 Volume: /app (1GB)
# - 添加环境变量（见下方）
# - 点击 Deploy

# 3. 环境变量
AI_MAX_TOKENS=4000
TZ=Asia/Shanghai
NOFX_ADMIN_PASSWORD=your_password
```

---

## 🧪 测试 API

### 1. 创建 API Key（Web 界面）

```
1. 访问 http://localhost:8080
2. 登录账号
3. 进入 "API Keys" 页面（需要添加到导航）
4. 点击 "Create New API Key"
5. 保存生成的 Key
```

### 2. 测试 API 调用

```bash
# 设置 API Key
export API_KEY="aitrader_your_generated_key"

# 测试 1：列出交易员
curl -X GET http://localhost:8080/api/v1/traders \
  -H "Authorization: Bearer $API_KEY"

# 测试 2：获取 AI 模型
curl -X GET http://localhost:8080/api/v1/models \
  -H "Authorization: Bearer $API_KEY"

# 测试 3：获取交易所
curl -X GET http://localhost:8080/api/v1/exchanges \
  -H "Authorization: Bearer $API_KEY"

# 测试 4：启动交易员
curl -X POST http://localhost:8080/api/v1/traders/trader_001/start \
  -H "Authorization: Bearer $API_KEY" \
  -H "Content-Type: application/json" \
  -d '{"trader_id": "trader_001"}'
```

---

## 📝 前端集成（可选）

如果你想在 Web 界面添加 API Keys 管理：

### 1. 添加路由

```typescript
// web/src/App.tsx
import APIKeysPage from './components/APIKeysPage';

// 在路由中添加
<Route path="/api-keys" element={<APIKeysPage />} />
```

### 2. 添加导航

```typescript
// web/src/components/Sidebar.tsx
import { Key } from 'lucide-react';

<Link to="/api-keys">
  <Key size={20} />
  <span>API Keys</span>
</Link>
```

### 3. 安装前端依赖（如果还没有）

```bash
cd web
npm install
npm run dev
```

---

## 🔐 API Key 格式

### 生成的 Key 格式
```
aitrader_xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
```

### 使用方式
```bash
Authorization: Bearer aitrader_xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
```

---

## 📊 使用示例

### Python 客户端

```python
import requests

class NOFXClient:
    def __init__(self, api_key, base_url="http://localhost:8080"):
        self.api_key = api_key
        self.base_url = base_url
        self.headers = {
            "Authorization": f"Bearer {api_key}",
            "Content-Type": "application/json"
        }
    
    def list_traders(self):
        response = requests.get(
            f"{self.base_url}/api/v1/traders",
            headers=self.headers
        )
        return response.json()
    
    def start_trader(self, trader_id):
        response = requests.post(
            f"{self.base_url}/api/v1/traders/{trader_id}/start",
            headers=self.headers,
            json={"trader_id": trader_id}
        )
        return response.json()
    
    def get_performance(self, trader_id):
        response = requests.get(
            f"{self.base_url}/api/v1/traders/{trader_id}/performance",
            headers=self.headers
        )
        return response.json()

# 使用
client = NOFXClient("aitrader_your_api_key")
traders = client.list_traders()
print(f"Found {traders['count']} traders")

# 启动第一个交易员
if traders['count'] > 0:
    trader_id = traders['traders'][0]['id']
    result = client.start_trader(trader_id)
    print(f"Started trader: {result}")
```

### JavaScript 客户端

```javascript
class NOFXClient {
  constructor(apiKey, baseURL = 'http://localhost:8080') {
    this.apiKey = apiKey;
    this.baseURL = baseURL;
  }

  async listTraders() {
    const response = await fetch(`${this.baseURL}/api/v1/traders`, {
      headers: {
        'Authorization': `Bearer ${this.apiKey}`,
      },
    });
    return response.json();
  }

  async startTrader(traderId) {
    const response = await fetch(
      `${this.baseURL}/api/v1/traders/${traderId}/start`,
      {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${this.apiKey}`,
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({ trader_id: traderId }),
      }
    );
    return response.json();
  }

  async getPerformance(traderId) {
    const response = await fetch(
      `${this.baseURL}/api/v1/traders/${traderId}/performance`,
      {
        headers: {
          'Authorization': `Bearer ${this.apiKey}`,
        },
      }
    );
    return response.json();
  }
}

// 使用
const client = new NOFXClient('aitrader_your_api_key');
const traders = await client.listTraders();
console.log(`Found ${traders.count} traders`);
```

---

## 🎯 验证商业模式

### 测试场景

#### 场景 1：个人开发者
```
1. 注册账号
2. 配置 AI 模型和交易所
3. 创建交易员
4. 获取 API Key
5. 通过 API 启动/停止交易
6. 监控绩效
```

#### 场景 2：量化团队
```
1. 多个成员共享账号
2. 创建多个 API Keys（不同权限）
3. 集成到自己的系统
4. 批量管理交易员
5. 统一监控和分析
```

#### 场景 3：SaaS 客户
```
1. 购买 API 额度
2. 获取 API Key
3. 调用交易信号 API
4. 集成到自己的交易系统
5. 按使用量付费
```

---

## 📈 下一步计划

### Phase 1.5（当前 MVP 优化）
- [ ] 添加 API 文档页面
- [ ] 优化错误提示
- [ ] 添加使用示例
- [ ] 性能监控

### Phase 2（商业化）
- [ ] 套餐管理（Free/Pro/Enterprise）
- [ ] Stripe 支付集成
- [ ] 更详细的统计
- [ ] WebSocket 实时推送
- [ ] 多语言 SDK

### Phase 3（企业功能）
- [ ] 团队协作
- [ ] 权限管理
- [ ] 私有部署
- [ ] SLA 保证

---

## 🐛 故障排查

### 问题 1：API Key 无法创建
```bash
# 检查用户是否登录
curl -X GET http://localhost:8080/api/api-keys \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"

# 查看日志
docker compose logs -f nofx | grep "api-keys"
```

### 问题 2：API 调用返回 401
```bash
# 检查 API Key 格式
echo $API_KEY
# 应该是: aitrader_xxxxx

# 检查 Header 格式
curl -v http://localhost:8080/api/v1/traders \
  -H "Authorization: Bearer $API_KEY"
# 应该看到: Authorization: Bearer aitrader_xxxxx
```

### 问题 3：速率限制
```bash
# 查看当前使用量
curl -X GET http://localhost:8080/api/api-keys/usage \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"

# 重置使用量（临时）
sqlite3 config.db "UPDATE api_keys SET usage_count = 0 WHERE id = 'key_id';"
```

---

## ✅ 功能检查清单

### 后端
- [x] API Key 生成
- [x] API Key 验证
- [x] 速率限制
- [x] 使用日志
- [x] 统计接口
- [x] 交易员 API
- [x] 模型/交易所 API

### 前端
- [x] API Keys 管理页面
- [x] 创建 API Key
- [x] 删除 API Key
- [x] 使用统计展示
- [x] 快速开始文档

### 测试
- [ ] 创建 API Key
- [ ] 调用 API 成功
- [ ] 速率限制生效
- [ ] 日志记录正常
- [ ] 统计数据准确

---

## 🎉 恭喜！

你已经完成了 NOFX 的 MVP SaaS 改造！

**现在可以：**
- ✅ 通过 Web 界面配置交易
- ✅ 通过 API Key 调用接口
- ✅ 监控使用情况
- ✅ 验证商业模式

**开始测试你的想法吧！** 🚀

---

## 📞 需要帮助？

如果遇到问题：
1. 查看 `docs/MVP_SAAS_GUIDE.md` 详细文档
2. 检查日志：`docker compose logs -f nofx`
3. 查看数据库：`sqlite3 config.db`
4. 提交 Issue 或联系支持

**祝你成功！** 💪
