package handlers

import (
	"net/http"
	"ai-trader/config"

	"github.com/gin-gonic/gin"
)

// APIKeyHandlers API Key相关处理器
type APIKeyHandlers struct {
	db *config.Database
}

// NewAPIKeyHandlers 创建API Key处理器
func NewAPIKeyHandlers(db *config.Database) *APIKeyHandlers {
	return &APIKeyHandlers{db: db}
}

// CreateAPIKeyRequest 创建API Key请求
type CreateAPIKeyRequest struct {
	Name      string `json:"name" binding:"required"`
	RateLimit int    `json:"rate_limit,omitempty"`
}

// CreateAPIKey 创建新的API Key
func (h *APIKeyHandlers) CreateAPIKey(c *gin.Context) {
	var req CreateAPIKeyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_request",
			"message": err.Error(),
		})
		return
	}

	// 获取用户ID
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "unauthorized",
		})
		return
	}

	// 设置默认速率限制
	if req.RateLimit == 0 {
		req.RateLimit = 1000 // 默认每月1000次
	}

	// 创建API Key
	apiKey, keyInfo, err := h.db.CreateAPIKey(userID.(string), req.Name, req.RateLimit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "creation_failed",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"api_key": apiKey, // 只在创建时返回完整key
		"info":    keyInfo,
		"warning": "Please save this API key. You won't be able to see it again!",
	})
}

// ListAPIKeys 列出用户的所有API Keys
func (h *APIKeyHandlers) ListAPIKeys(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "unauthorized",
		})
		return
	}

	keys, err := h.db.GetAPIKeys(userID.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "query_failed",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"keys":  keys,
		"count": len(keys),
	})
}

// RevokeAPIKey 撤销API Key
func (h *APIKeyHandlers) RevokeAPIKey(c *gin.Context) {
	keyID := c.Param("id")
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "unauthorized",
		})
		return
	}

	err := h.db.RevokeAPIKey(keyID, userID.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "revoke_failed",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "API key revoked successfully",
	})
}

// DeleteAPIKey 删除API Key
func (h *APIKeyHandlers) DeleteAPIKey(c *gin.Context) {
	keyID := c.Param("id")
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "unauthorized",
		})
		return
	}

	err := h.db.DeleteAPIKey(keyID, userID.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "delete_failed",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "API key deleted successfully",
	})
}

// GetUsageStats 获取API使用统计
func (h *APIKeyHandlers) GetUsageStats(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "unauthorized",
		})
		return
	}

	// 默认查询最近30天
	days := 30

	stats, err := h.db.GetAPIUsageStats(userID.(string), days)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "query_failed",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, stats)
}
