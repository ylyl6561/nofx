package middleware

import (
	"net/http"
	"nofx/config"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// APIKeyAuth API Key认证中间件
func APIKeyAuth(db *config.Database) gin.HandlerFunc {
	return func(c *gin.Context) {
		startTime := time.Now()

		// 从Header获取API Key
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "missing_api_key",
				"message": "Please provide API key in Authorization header",
				"example": "Authorization: Bearer aitrader_xxxxx",
			})
			c.Abort()
			return
		}

		// 解析Bearer Token
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "invalid_authorization_format",
				"message": "Use format: Authorization: Bearer aitrader_xxxxx",
			})
			c.Abort()
			return
		}

		apiKey := parts[1]

		// 验证API Key
		keyInfo, err := db.ValidateAPIKey(apiKey)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "invalid_api_key",
				"message": err.Error(),
			})
			c.Abort()
			return
		}

		// 将用户信息存入上下文
		c.Set("user_id", keyInfo.UserID)
		c.Set("api_key_id", keyInfo.ID)
		c.Set("rate_limit", keyInfo.RateLimit)
		c.Set("usage_count", keyInfo.UsageCount)

		// 继续处理请求
		c.Next()

		// 记录API使用
		responseTime := int(time.Since(startTime).Milliseconds())
		statusCode := c.Writer.Status()
		endpoint := c.Request.URL.Path
		method := c.Request.Method

		// 异步记录日志和更新使用次数
		go func() {
			_ = db.IncrementAPIUsage(keyInfo.ID)
			_ = db.LogAPIUsage(keyInfo.ID, keyInfo.UserID, endpoint, method, statusCode, responseTime)
		}()
	}
}

// OptionalAuth 可选认证中间件（支持JWT或API Key）
func OptionalAuth(db *config.Database, jwtSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			// 无认证，继续（某些端点可能允许匿名访问）
			c.Next()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 {
			c.Next()
			return
		}

		tokenType := parts[0]
		token := parts[1]

		if tokenType == "Bearer" && strings.HasPrefix(token, "aitrader_") {
			// API Key认证
			keyInfo, err := db.ValidateAPIKey(token)
			if err == nil {
				c.Set("user_id", keyInfo.UserID)
				c.Set("api_key_id", keyInfo.ID)
				c.Set("auth_type", "api_key")
			}
		} else if tokenType == "Bearer" {
			// JWT认证（现有逻辑）
			// 这里可以添加JWT验证逻辑
			c.Set("auth_type", "jwt")
		}

		c.Next()
	}
}
