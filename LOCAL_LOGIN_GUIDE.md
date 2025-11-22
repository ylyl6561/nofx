# 🔐 本地登录快速指南

## 📊 测试账号

```
邮箱: test@local.dev
密码: test123
```

## 🚀 三种登录模式

### 模式1: 跳过OTP（推荐）⭐

**配置** (`.env.local`):
```bash
SKIP_OTP=true
```

**登录流程**:
1. 输入邮箱: `test@local.dev`
2. 输入密码: `test123`
3. ✅ 直接登录成功！

**适用场景**: 测试登录功能，但不想设置Google Authenticator

---

### 模式2: 跳过所有认证（最快）⚡

**配置** (`.env.local`):
```bash
SKIP_AUTH=true
```

**登录流程**:
- 无需登录
- 所有API自动使用测试用户

**适用场景**: 快速开发，直接测试API

---

### 模式3: 完整流程（包含OTP）🔒

**配置** (`.env.local`):
```bash
# 不设置 SKIP_OTP 和 SKIP_AUTH
# 或明确设置为 false
```

**登录流程**:
1. 输入邮箱: `test@local.dev`
2. 输入密码: `test123`
3. 输入OTP验证码（从Google Authenticator获取）
4. 登录成功

**OTP设置**:
- Secret: `JBSWY3DPEHPK3PXP`
- 在线工具: https://totp.danhersam.com/

**适用场景**: 测试完整的登录安全流程

---

## 📝 完整的 .env.local 配置

### 推荐配置（跳过OTP）

```bash
# 环境标识
APP_ENV=local

# 数据库
DATABASE_URL=postgresql://yuliang@localhost:5432/nofx_test?sslmode=disable

# 跳过OTP验证（推荐）
SKIP_OTP=true

# 其他配置
API_SERVER_PORT=8080
LOG_LEVEL=debug
```

## 🔄 快速切换

```bash
# 跳过OTP
echo "SKIP_OTP=true" >> .env.local

# 跳过所有认证
echo "SKIP_AUTH=true" >> .env.local

# 恢复正常流程
# 删除或注释掉 SKIP_OTP 和 SKIP_AUTH
```

## 💡 使用建议

| 场景 | 推荐配置 |
|------|---------|
| 快速开发API | `SKIP_AUTH=true` |
| 测试登录功能 | `SKIP_OTP=true` |
| 测试完整安全流程 | 不设置任何SKIP |
| 前端开发 | `SKIP_OTP=true` |

## 🔍 验证配置

启动应用后，查看日志：

```bash
# 如果看到这行，说明跳过OTP已启用
🔓 [本地开发] 跳过OTP验证，直接登录: test@local.dev

# 如果看到这行，说明跳过认证已启用
🔓 [本地开发] 跳过认证，使用测试用户: test@local.dev (test-user-local)
```

## 🎯 现在开始

1. **编辑 `.env.local`**，添加：
   ```bash
   SKIP_OTP=true
   ```

2. **重启应用**：
   ```bash
   ./scripts/run_local.sh
   ```

3. **登录前端**：
   - 邮箱: `test@local.dev`
   - 密码: `test123`
   - ✅ 直接登录成功！

---

💡 **提示**: 开发时使用 `SKIP_OTP=true`，提交代码前记得检查配置！
