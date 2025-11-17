# AI API URL 错误修复

## 错误信息
```
❌ 获取AI决策失败: 调用AI API失败: 发送请求失败: Post "/chat/completions": unsupported protocol scheme ""
```

## 根本原因

在 `trader/auto_trader.go` 中，Qwen 和 DeepSeek 的 AI 客户端初始化时调用了错误的方法：

### 错误代码（第 145 和 154 行）
```go
// Qwen - 错误
mcpClient.SetAPIKey(config.QwenKey, config.CustomAPIURL, config.CustomModelName)

// DeepSeek - 错误  
mcpClient.SetAPIKey(config.DeepSeekKey, config.CustomAPIURL, config.CustomModelName)
```

### 问题分析

`SetAPIKey` 方法的实现（`mcp/client.go` 第 77-91 行）：
```go
func (client *Client) SetAPIKey(apiKey, apiURL, customModel string) {
    client.Provider = ProviderCustom  // ← 强制设置为 Custom
    client.APIKey = apiKey
    client.BaseURL = apiURL           // ← 如果 apiURL 为空，BaseURL 就是空字符串
    client.Model = customModel
}
```

**当 `CustomAPIURL` 为空时（大多数情况）：**
1. `BaseURL` 被设置为空字符串 `""`
2. 请求 URL 变成 `"" + "/chat/completions"` = `"/chat/completions"`
3. HTTP 请求失败：`unsupported protocol scheme ""`

## 解决方案

### 修复后的代码
```go
// Qwen - 正确
mcpClient.SetQwenAPIKey(config.QwenKey, config.CustomAPIURL, config.CustomModelName)

// DeepSeek - 正确
mcpClient.SetDeepSeekAPIKey(config.DeepSeekKey, config.CustomAPIURL, config.CustomModelName)
```

### 为什么这样修复？

`SetQwenAPIKey` 和 `SetDeepSeekAPIKey` 方法会正确处理默认 URL：

```go
func (client *Client) SetDeepSeekAPIKey(apiKey string, customURL string, customModel string) {
    client.Provider = ProviderDeepSeek
    client.APIKey = apiKey
    if customURL != "" {
        client.BaseURL = customURL
    } else {
        client.BaseURL = "https://api.deepseek.com/v1"  // ← 使用默认 URL
    }
    // ...
}

func (client *Client) SetQwenAPIKey(apiKey string, customURL string, customModel string) {
    client.Provider = ProviderQwen
    client.APIKey = apiKey
    if customURL != "" {
        client.BaseURL = customURL
    } else {
        client.BaseURL = "https://dashscope.aliyuncs.com/compatible-mode/v1"  // ← 使用默认 URL
    }
    // ...
}
```

## 影响范围

- ✅ **DeepSeek 用户**：修复后可以正常使用默认 API
- ✅ **Qwen 用户**：修复后可以正常使用默认 API  
- ✅ **自定义 API 用户**：不受影响（使用 `SetAPIKey` 方法）

## 验证步骤

### 1. 重新编译
```bash
go build -o nofx main.go
```

### 2. 重启服务
```bash
./nofx
```

### 3. 检查日志
启动交易员后，应该看到类似日志：
```
🔧 [MCP] DeepSeek 使用默认 BaseURL: https://api.deepseek.com/v1
🔧 [MCP] DeepSeek 使用默认 Model: deepseek-chat
🔧 [MCP] DeepSeek API Key: sk-x...xxxx
📡 [MCP] AI 请求配置:
   Provider: deepseek
   BaseURL: https://api.deepseek.com/v1
   Model: deepseek-chat
```

### 4. 测试 AI 决策
交易员运行时应该能正常获取 AI 决策，不再报 `unsupported protocol scheme` 错误。

## 相关文件
- `/Users/yuliang/project/nofx/trader/auto_trader.go` - 已修复（第 145, 154 行）
- `/Users/yuliang/project/nofx/mcp/client.go` - AI 客户端实现

## 注意事项
1. 修复后需要重新编译和部署
2. 如果使用 Docker，需要重新构建镜像
3. 正在运行的交易员需要重启才能生效
