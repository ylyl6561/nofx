package mcp

import "net/http"

// AIResponse AI API 响应（包含 token 使用信息）
type AIResponse struct {
	Content          string `json:"content"`            // AI 返回的内容
	PromptTokens     int    `json:"prompt_tokens"`      // 输入 token 数
	CompletionTokens int    `json:"completion_tokens"`  // 输出 token 数
	TotalTokens      int    `json:"total_tokens"`       // 总 token 数
}

// AIClient AI客户端接口
type AIClient interface {
	SetAPIKey(apiKey string, customURL string, customModel string)
	// CallWithMessages 使用 system + user prompt 调用AI API（旧版本，仅返回内容）
	CallWithMessages(systemPrompt, userPrompt string) (string, error)
	// CallWithMessagesAndTokens 使用 system + user prompt 调用AI API（新版本，返回内容和 token 信息）
	CallWithMessagesAndTokens(systemPrompt, userPrompt string) (*AIResponse, error)

	setAuthHeader(reqHeaders http.Header)
}
