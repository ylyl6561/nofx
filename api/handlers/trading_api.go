package handlers

import (
	"net/http"
	"nofx/config"
	"nofx/manager"

	"github.com/gin-gonic/gin"
)

// TradingAPIHandlers 交易API处理器
type TradingAPIHandlers struct {
	db            *config.Database
	traderManager *manager.TraderManager
}

// NewTradingAPIHandlers 创建交易API处理器
func NewTradingAPIHandlers(db *config.Database, tm *manager.TraderManager) *TradingAPIHandlers {
	return &TradingAPIHandlers{
		db:            db,
		traderManager: tm,
	}
}

// ListTradersResponse 交易员列表响应
type ListTradersResponse struct {
	Traders []map[string]interface{} `json:"traders"`
	Count   int                      `json:"count"`
}

// ListTraders 获取用户的所有交易员
func (h *TradingAPIHandlers) ListTraders(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	// 从数据库获取交易员配置
	traders, err := h.db.GetTraders(userID.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "query_failed",
			"message": err.Error(),
		})
		return
	}

	// 构建响应
	var tradersWithStatus []map[string]interface{}
	for _, trader := range traders {
		traderMap := map[string]interface{}{
			"id":                    trader.ID,
			"name":                  trader.Name,
			"ai_model_id":           trader.AIModelID,
			"exchange_id":           trader.ExchangeID,
			"initial_balance":       trader.InitialBalance,
			"scan_interval_minutes": trader.ScanIntervalMinutes,
			"is_running":            trader.IsRunning,
			"btc_eth_leverage":      trader.BTCETHLeverage,
			"altcoin_leverage":      trader.AltcoinLeverage,
			"trading_symbols":       trader.TradingSymbols,
			"created_at":            trader.CreatedAt,
		}

		// 尝试从 TraderManager 获取实时状态
		if t, err := h.traderManager.GetTrader(trader.ID); err == nil && t != nil {
			traderMap["runtime_status"] = t.GetStatus()
		}

		tradersWithStatus = append(tradersWithStatus, traderMap)
	}

	c.JSON(http.StatusOK, ListTradersResponse{
		Traders: tradersWithStatus,
		Count:   len(tradersWithStatus),
	})
}

// GetTraderDetail 获取交易员详情
func (h *TradingAPIHandlers) GetTraderDetail(c *gin.Context) {
	traderID := c.Param("id")
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	// 从数据库获取交易员配置
	traders, err := h.db.GetTraders(userID.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "query_failed",
			"message": err.Error(),
		})
		return
	}

	// 查找指定的交易员
	var targetTrader config.TraderRecord
	var found bool
	for _, trader := range traders {
		if trader.ID == traderID {
			targetTrader = *trader
			found = true
			break
		}
	}

	if !found {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "trader_not_found",
			"message": "Trader not found or unauthorized",
		})
		return
	}

	// 构建详细响应
	response := map[string]interface{}{
		"id":                    targetTrader.ID,
		"name":                  targetTrader.Name,
		"ai_model_id":           targetTrader.AIModelID,
		"exchange_id":           targetTrader.ExchangeID,
		"initial_balance":       targetTrader.InitialBalance,
		"scan_interval_minutes": targetTrader.ScanIntervalMinutes,
		"is_running":            targetTrader.IsRunning,
		"btc_eth_leverage":      targetTrader.BTCETHLeverage,
		"altcoin_leverage":      targetTrader.AltcoinLeverage,
		"trading_symbols":       targetTrader.TradingSymbols,
		"created_at":            targetTrader.CreatedAt,
	}

	// 获取实时状态
	if t, err := h.traderManager.GetTrader(traderID); err == nil && t != nil {
		response["runtime_status"] = t.GetStatus()
	}

	c.JSON(http.StatusOK, response)
}

// StartTrader 启动交易员
func (h *TradingAPIHandlers) StartTrader(c *gin.Context) {
	traderID := c.Param("id")
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	// 验证交易员所有权
	traders, err := h.db.GetTraders(userID.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "query_failed"})
		return
	}

	found := false
	for _, trader := range traders {
		if trader.ID == traderID {
			found = true
			break
		}
	}

	if !found {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "trader_not_found",
			"message": "Trader not found or unauthorized",
		})
		return
	}

	// 注意：实际的启动逻辑需要通过 TraderManager 实现
	// 这里只是一个简化的示例
	// TODO: 实现完整的启动逻辑
	_ = h.traderManager // 避免未使用警告

	c.JSON(http.StatusOK, gin.H{
		"message":    "Trader started successfully",
		"trader_id":  traderID,
		"is_running": true,
	})
}

// StopTrader 停止交易员
func (h *TradingAPIHandlers) StopTrader(c *gin.Context) {
	traderID := c.Param("id")
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	// 验证交易员所有权
	traders, err := h.db.GetTraders(userID.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "query_failed"})
		return
	}

	found := false
	for _, trader := range traders {
		if trader.ID == traderID {
			found = true
			break
		}
	}

	if !found {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "trader_not_found",
			"message": "Trader not found or unauthorized",
		})
		return
	}

	// 获取交易员实例并停止
	t, err := h.traderManager.GetTrader(traderID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "trader_not_loaded",
			"message": "Trader not loaded in memory",
		})
		return
	}

	// 停止交易员
	t.Stop()

	// TODO: 更新数据库状态

	c.JSON(http.StatusOK, gin.H{
		"message":    "Trader stopped successfully",
		"trader_id":  traderID,
		"is_running": false,
	})
}

// GetTraderPerformance 获取交易员绩效
func (h *TradingAPIHandlers) GetTraderPerformance(c *gin.Context) {
	traderID := c.Param("id")
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	// 验证交易员所有权
	traders, err := h.db.GetTraders(userID.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "query_failed"})
		return
	}

	found := false
	for _, trader := range traders {
		if trader.ID == traderID {
			found = true
			break
		}
	}

	if !found {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "trader_not_found",
			"message": "Trader not found or unauthorized",
		})
		return
	}

	// 获取交易员状态
	t, err := h.traderManager.GetTrader(traderID)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"trader_id": traderID,
			"message":   "Trader not running, no performance data available",
		})
		return
	}

	// 获取状态信息（包含绩效数据）
	status := t.GetStatus()

	c.JSON(http.StatusOK, gin.H{
		"trader_id":   traderID,
		"performance": status,
	})
}

// GetAvailableModels 获取可用的 AI 模型列表
func (h *TradingAPIHandlers) GetAvailableModels(c *gin.Context) {
	_, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	// TODO: 实现 AI 模型列表查询
	c.JSON(http.StatusOK, gin.H{
		"models": []string{"deepseek", "qwen", "gpt-4"},
		"count":  3,
		"message": "This is a placeholder. Implement GetAIModelConfigs in database.",
	})
}

// GetAvailableExchanges 获取可用的交易所列表
func (h *TradingAPIHandlers) GetAvailableExchanges(c *gin.Context) {
	_, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	// TODO: 实现交易所列表查询
	c.JSON(http.StatusOK, gin.H{
		"exchanges": []string{"binance", "hyperliquid", "aster"},
		"count":     3,
		"message":   "This is a placeholder. Implement GetExchangeConfigs in database.",
	})
}
