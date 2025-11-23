# 🔧 前端登录修复说明

## 📋 问题描述

**症状**: 后端返回登录成功，但前端显示"未知错误"且未跳转

**原因**: 前端 `AuthContext.tsx` 的 `login` 函数只处理了需要OTP验证的情况，没有处理本地环境跳过OTP直接返回token的情况。

## ✅ 已修复

### 修改文件
- `web/src/contexts/AuthContext.tsx`

### 修改内容

**修复前**:
```typescript
if (response.ok) {
  if (data.requires_otp) {
    return {
      success: true,
      userID: data.user_id,
      requiresOTP: true,
      message: data.message,
    }
  }
} else {
  return { success: false, message: data.error }
}
// 走到这里返回"未知错误"
return { success: false, message: '未知错误' }
```

**修复后**:
```typescript
if (response.ok) {
  // 情况1: 需要OTP验证
  if (data.requires_otp) {
    return {
      success: true,
      userID: data.user_id,
      requiresOTP: true,
      message: data.message,
    }
  }
  
  // 情况2: 直接返回token（本地环境跳过OTP）
  if (data.token) {
    localStorage.setItem('auth_token', data.token)
    setToken(data.token)
    setUser({
      id: data.user_id,
      email: data.email,
    })
    return {
      success: true,
      message: data.message || '登录成功',
    }
  }
}
```

## 🚀 现在的登录流程

### 本地环境（SKIP_OTP=true）

1. 用户输入邮箱密码
2. 后端验证成功，直接返回token
3. 前端保存token和用户信息
4. ✅ 自动跳转到主页

### 生产环境（正常OTP流程）

1. 用户输入邮箱密码
2. 后端返回 `requires_otp: true`
3. 前端显示OTP输入框
4. 用户输入OTP验证码
5. 验证成功后返回token
6. ✅ 跳转到主页

## 📝 测试步骤

1. **重新编译前端**:
   ```bash
   cd web
   npm run build
   # 或开发模式
   npm run dev
   ```

2. **启动后端**:
   ```bash
   ./scripts/run_local.sh
   ```

3. **测试登录**:
   - 打开浏览器访问前端
   - 输入: `test@local.dev` / `test123`
   - ✅ 应该直接登录成功并跳转

## 💡 配置说明

### .env.local
```bash
# 跳过OTP验证
SKIP_OTP=true
```

### 后端响应格式

**跳过OTP时**:
```json
{
  "token": "eyJhbGci...",
  "user_id": "test-user-local",
  "email": "test@local.dev",
  "message": "登录成功（本地开发模式，已跳过OTP）"
}
```

**需要OTP时**:
```json
{
  "user_id": "xxx",
  "email": "xxx@example.com",
  "message": "请输入Google Authenticator验证码",
  "requires_otp": true
}
```

## 🔍 验证修复

### 检查点1: 后端日志
```
🔓 [本地开发] 跳过OTP验证，直接登录: test@local.dev
```

### 检查点2: 浏览器控制台
- Network标签: 查看 `/api/login` 返回 200 和 token
- Application标签: 查看 localStorage 中有 `auth_token`
- Console标签: 无错误信息

### 检查点3: 页面行为
- ✅ 登录成功提示
- ✅ 自动跳转到主页
- ✅ 显示用户信息

## 🎯 相关文件

- 后端登录: `api/server.go` (handleLogin)
- 前端登录: `web/src/contexts/AuthContext.tsx` (login)
- 环境配置: `.env.local` (SKIP_OTP)

---

**修复日期**: 2025-11-23
**修复内容**: 前端登录逻辑支持本地环境跳过OTP
