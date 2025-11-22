package handlers

import (
	"net/http"
	"nofx/config"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// 统一的API响应格式辅助函数
func successResponse(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    data,
		"code":    http.StatusOK,
	})
}

func errorResponse(c *gin.Context, code int, message string) {
	c.JSON(code, gin.H{
		"success": false,
		"error":   message,
		"code":    code,
	})
}

// UserConfigHandlers 用户配置处理器
type UserConfigHandlers struct {
	database *config.Database
}

// NewUserConfigHandlers 创建用户配置处理器
func NewUserConfigHandlers(database *config.Database) *UserConfigHandlers {
	return &UserConfigHandlers{
		database: database,
	}
}

// GetUserTradingConfig 获取用户交易配置
func (h *UserConfigHandlers) GetUserTradingConfig(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		errorResponse(c, http.StatusUnauthorized, "未授权")
		return
	}
	
	config, err := h.database.GetUserTradingConfig(userID)
	if err != nil {
		errorResponse(c, http.StatusInternalServerError, "获取用户配置失败: "+err.Error())
		return
	}
	
	// 转换为响应格式
	response := map[string]interface{}{
		"user_id":              config.UserID,
		"btc_eth_leverage":     config.BTCETHLeverage,
		"altcoin_leverage":     config.AltcoinLeverage,
		"max_daily_loss":       config.MaxDailyLoss,
		"max_drawdown":         config.MaxDrawdown,
		"stop_trading_minutes": config.StopTradingMinutes,
		"use_default_coins":    config.UseDefaultCoins,
		"custom_coins":         config.CustomCoins,
		"created_at":           config.CreatedAt,
		"updated_at":           config.UpdatedAt,
	}
	
	successResponse(c, response)
}

// SaveUserTradingConfig 保存用户交易配置
func (h *UserConfigHandlers) SaveUserTradingConfig(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		errorResponse(c, http.StatusUnauthorized, "未授权")
		return
	}
	
	var req struct {
		BTCETHLeverage     *int      `json:"btc_eth_leverage"`
		AltcoinLeverage    *int      `json:"altcoin_leverage"`
		MaxDailyLoss       *float64  `json:"max_daily_loss"`
		MaxDrawdown        *float64  `json:"max_drawdown"`
		StopTradingMinutes *int      `json:"stop_trading_minutes"`
		UseDefaultCoins    bool      `json:"use_default_coins"`
		CustomCoins        []string  `json:"custom_coins"`
	}
	
	if err := c.ShouldBindJSON(&req); err != nil {
		errorResponse(c, http.StatusBadRequest, "请求参数错误: "+err.Error())
		return
	}
	
	// 构建配置对象
	config := &config.UserTradingConfig{
		UserID:              userID,
		BTCETHLeverage:      req.BTCETHLeverage,
		AltcoinLeverage:     req.AltcoinLeverage,
		MaxDailyLoss:        req.MaxDailyLoss,
		MaxDrawdown:         req.MaxDrawdown,
		StopTradingMinutes:  req.StopTradingMinutes,
		UseDefaultCoins:     req.UseDefaultCoins,
		CustomCoins:         req.CustomCoins,
	}
	
	if err := h.database.SaveUserTradingConfig(config); err != nil {
		errorResponse(c, http.StatusInternalServerError, "保存用户配置失败: "+err.Error())
		return
	}
	
	successResponse(c, gin.H{"message": "保存成功"})
}

// DeleteUserTradingConfig 删除用户交易配置（恢复使用系统默认值）
func (h *UserConfigHandlers) DeleteUserTradingConfig(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		errorResponse(c, http.StatusUnauthorized, "未授权")
		return
	}
	
	if err := h.database.DeleteUserTradingConfig(userID); err != nil {
		errorResponse(c, http.StatusInternalServerError, "删除用户配置失败: "+err.Error())
		return
	}
	
	successResponse(c, gin.H{"message": "已恢复使用系统默认配置"})
}

// GetDecisionLogs 获取决策日志
func (h *UserConfigHandlers) GetDecisionLogs(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		errorResponse(c, http.StatusUnauthorized, "未授权")
		return
	}
	
	traderID := c.Param("id")
	if traderID == "" {
		errorResponse(c, http.StatusBadRequest, "缺少trader_id参数")
		return
	}
	
	// 解析查询参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	fromStr := c.Query("from")
	toStr := c.Query("to")
	
	var from, to time.Time
	if fromStr != "" {
		from, _ = time.Parse(time.RFC3339, fromStr)
	}
	if toStr != "" {
		to, _ = time.Parse(time.RFC3339, toStr)
	}
	
	// 查询决策日志
	logs, total, err := h.database.GetDecisionLogs(userID, traderID, from, to, page, limit)
	if err != nil {
		errorResponse(c, http.StatusInternalServerError, "查询决策日志失败: "+err.Error())
		return
	}
	
	successResponse(c, gin.H{
		"total": total,
		"page":  page,
		"limit": limit,
		"logs":  logs,
	})
}

// GetDecisionLogStatistics 获取决策日志统计
func (h *UserConfigHandlers) GetDecisionLogStatistics(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		errorResponse(c, http.StatusUnauthorized, "未授权")
		return
	}
	
	traderID := c.Param("id")
	if traderID == "" {
		errorResponse(c, http.StatusBadRequest, "缺少trader_id参数")
		return
	}
	
	// 解析查询参数
	fromStr := c.Query("from")
	toStr := c.Query("to")
	
	var from, to time.Time
	if fromStr != "" {
		from, _ = time.Parse(time.RFC3339, fromStr)
	}
	if toStr != "" {
		to, _ = time.Parse(time.RFC3339, toStr)
	}
	
	// 查询统计信息
	stats, err := h.database.GetDecisionLogStatistics(userID, traderID, from, to)
	if err != nil {
		errorResponse(c, http.StatusInternalServerError, "查询统计信息失败: "+err.Error())
		return
	}
	
	successResponse(c, stats)
}

// GetEquityHistory 获取权益历史
func (h *UserConfigHandlers) GetEquityHistory(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		errorResponse(c, http.StatusUnauthorized, "未授权")
		return
	}
	
	traderID := c.Param("id")
	if traderID == "" {
		errorResponse(c, http.StatusBadRequest, "缺少trader_id参数")
		return
	}
	
	// 解析查询参数
	fromStr := c.Query("from")
	toStr := c.Query("to")
	interval := c.DefaultQuery("interval", "1h")
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "1000"))
	
	var from, to time.Time
	if fromStr != "" {
		from, _ = time.Parse(time.RFC3339, fromStr)
	}
	if toStr != "" {
		to, _ = time.Parse(time.RFC3339, toStr)
	}
	
	// 查询权益历史
	records, err := h.database.GetEquityHistory(userID, traderID, from, to, interval, limit)
	if err != nil {
		errorResponse(c, http.StatusInternalServerError, "查询权益历史失败: "+err.Error())
		return
	}
	
	successResponse(c, gin.H{
		"trader_id": traderID,
		"data":      records,
	})
}

// GetEquityHistoryBatch 批量获取权益历史
func (h *UserConfigHandlers) GetEquityHistoryBatch(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		errorResponse(c, http.StatusUnauthorized, "未授权")
		return
	}
	
	var req struct {
		TraderIDs []string `json:"trader_ids"`
		From      string   `json:"from"`
		To        string   `json:"to"`
		Limit     int      `json:"limit"`
	}
	
	if err := c.ShouldBindJSON(&req); err != nil {
		errorResponse(c, http.StatusBadRequest, "请求参数错误: "+err.Error())
		return
	}
	
	var from, to time.Time
	if req.From != "" {
		from, _ = time.Parse(time.RFC3339, req.From)
	}
	if req.To != "" {
		to, _ = time.Parse(time.RFC3339, req.To)
	}
	if req.Limit == 0 {
		req.Limit = 1000
	}
	
	// 批量查询
	result, err := h.database.GetEquityHistoryBatch(userID, req.TraderIDs, from, to, req.Limit)
	if err != nil {
		errorResponse(c, http.StatusInternalServerError, "批量查询权益历史失败: "+err.Error())
		return
	}
	
	successResponse(c, result)
}
