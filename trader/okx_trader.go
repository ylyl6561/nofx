package trader

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

// OKXTrader OKX交易所交易器
type OKXTrader struct {
	apiKey     string
	secretKey  string
	passphrase string
	baseURL    string
	client     *http.Client
	isTestnet  bool // 是否使用测试网

	// 余额缓存
	cachedBalance     map[string]interface{}
	balanceCacheTime  time.Time
	balanceCacheMutex sync.RWMutex

	// 持仓缓存
	cachedPositions     []map[string]interface{}
	positionsCacheTime  time.Time
	positionsCacheMutex sync.RWMutex

	// 缓存有效期（15秒）
	cacheDuration time.Duration
}

// NewOKXTrader 创建OKX交易器
func NewOKXTrader(apiKey, secretKey, passphrase string, testnet bool) *OKXTrader {
	baseURL := "https://www.okx.com"
	if testnet {
		// OKX 模拟盘使用相同的域名，但需要使用模拟盘的 API Key
		// 模拟盘和实盘通过 API Key 区分，URL 相同
		baseURL = "https://www.okx.com"
		log.Printf("🧪 OKX 交易器使用模拟盘模式")
	}
	
	return &OKXTrader{
		apiKey:        apiKey,
		secretKey:     secretKey,
		passphrase:    passphrase,
		baseURL:       baseURL,
		isTestnet:     testnet,
		client:        &http.Client{Timeout: 30 * time.Second},
		cacheDuration: 15 * time.Second,
	}
}

// sign 生成OKX API签名
func (t *OKXTrader) sign(timestamp, method, requestPath, body string) string {
	message := timestamp + method + requestPath + body
	h := hmac.New(sha256.New, []byte(t.secretKey))
	h.Write([]byte(message))
	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}

// request 发送HTTP请求到OKX
func (t *OKXTrader) request(method, endpoint, body string) ([]byte, error) {
	timestamp := time.Now().UTC().Format("2006-01-02T15:04:05.000Z")
	requestPath := endpoint
	
	sign := t.sign(timestamp, method, requestPath, body)

	url := t.baseURL + endpoint
	var req *http.Request
	var err error

	if body != "" {
		req, err = http.NewRequest(method, url, strings.NewReader(body))
	} else {
		req, err = http.NewRequest(method, url, nil)
	}
	
	if err != nil {
		return nil, err
	}

	// 设置请求头
	req.Header.Set("OK-ACCESS-KEY", t.apiKey)
	req.Header.Set("OK-ACCESS-SIGN", sign)
	req.Header.Set("OK-ACCESS-TIMESTAMP", timestamp)
	req.Header.Set("OK-ACCESS-PASSPHRASE", t.passphrase)
	req.Header.Set("Content-Type", "application/json")
	
	// 调试日志
	log.Printf("🔍 OKX API 请求: %s %s", method, endpoint)
	log.Printf("   API Key: %s... (长度: %d)", t.apiKey[:10], len(t.apiKey))
	log.Printf("   Passphrase: %s (长度: %d)", t.passphrase, len(t.passphrase))
	log.Printf("   Testnet: %v", t.isTestnet)

	resp, err := t.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return data, nil
}

// GetBalance 获取账户余额（带缓存）
func (t *OKXTrader) GetBalance() (map[string]interface{}, error) {
	// 检查缓存
	t.balanceCacheMutex.RLock()
	if t.cachedBalance != nil && time.Since(t.balanceCacheTime) < t.cacheDuration {
		cacheAge := time.Since(t.balanceCacheTime)
		t.balanceCacheMutex.RUnlock()
		log.Printf("✓ 使用缓存的账户余额（缓存时间: %.1f秒前）", cacheAge.Seconds())
		return t.cachedBalance, nil
	}
	t.balanceCacheMutex.RUnlock()

	// 调用API获取余额
	data, err := t.request("GET", "/api/v5/account/balance", "")
	if err != nil {
		return nil, fmt.Errorf("获取余额失败: %w", err)
	}

	var response struct {
		Code string `json:"code"`
		Msg  string `json:"msg"`
		Data []struct {
			TotalEq string `json:"totalEq"`
			Details []struct {
				Ccy       string `json:"ccy"`
				AvailBal  string `json:"availBal"`
				FrozenBal string `json:"frozenBal"`
				Eq        string `json:"eq"`
			} `json:"details"`
		} `json:"data"`
	}

	if err := json.Unmarshal(data, &response); err != nil {
		return nil, fmt.Errorf("解析余额数据失败: %w", err)
	}

	if response.Code != "0" {
		return nil, fmt.Errorf("OKX API错误: %s", response.Msg)
	}

	// 构建余额信息
	result := make(map[string]interface{})
	if len(response.Data) > 0 {
		totalEq, _ := strconv.ParseFloat(response.Data[0].TotalEq, 64)
		result["total_equity"] = totalEq

		// 查找USDT余额
		for _, detail := range response.Data[0].Details {
			if detail.Ccy == "USDT" {
				availBal, _ := strconv.ParseFloat(detail.AvailBal, 64)
				frozenBal, _ := strconv.ParseFloat(detail.FrozenBal, 64)
				eq, _ := strconv.ParseFloat(detail.Eq, 64)
				
				result["available_balance"] = availBal
				result["frozen_balance"] = frozenBal
				result["equity"] = eq
				break
			}
		}
	}

	// 更新缓存
	t.balanceCacheMutex.Lock()
	t.cachedBalance = result
	t.balanceCacheTime = time.Now()
	t.balanceCacheMutex.Unlock()

	log.Printf("✓ 已获取OKX账户余额: 总权益=%.2f USDT", result["total_equity"])
	return result, nil
}

// GetPositions 获取持仓信息（带缓存）
func (t *OKXTrader) GetPositions() ([]map[string]interface{}, error) {
	// 检查缓存
	t.positionsCacheMutex.RLock()
	if t.cachedPositions != nil && time.Since(t.positionsCacheTime) < t.cacheDuration {
		cacheAge := time.Since(t.positionsCacheTime)
		t.positionsCacheMutex.RUnlock()
		log.Printf("✓ 使用缓存的持仓信息（缓存时间: %.1f秒前）", cacheAge.Seconds())
		return t.cachedPositions, nil
	}
	t.positionsCacheMutex.RUnlock()

	// 调用API获取持仓
	data, err := t.request("GET", "/api/v5/account/positions", "")
	if err != nil {
		return nil, fmt.Errorf("获取持仓失败: %w", err)
	}

	var response struct {
		Code string `json:"code"`
		Msg  string `json:"msg"`
		Data []struct {
			InstId    string `json:"instId"`
			PosSide   string `json:"posSide"`
			Pos       string `json:"pos"`
			AvgPx     string `json:"avgPx"`
			MarkPx    string `json:"markPx"`
			Upl       string `json:"upl"`
			UplRatio  string `json:"uplRatio"`
			Lever     string `json:"lever"`
			Notional  string `json:"notionalUsd"`
			Margin    string `json:"margin"`
		} `json:"data"`
	}

	if err := json.Unmarshal(data, &response); err != nil {
		return nil, fmt.Errorf("解析持仓数据失败: %w", err)
	}

	if response.Code != "0" {
		return nil, fmt.Errorf("OKX API错误: %s", response.Msg)
	}

	// 转换为标准格式
	var positions []map[string]interface{}
	for _, pos := range response.Data {
		posSize, _ := strconv.ParseFloat(pos.Pos, 64)
		if posSize == 0 {
			continue // 跳过空仓位
		}

		avgPx, _ := strconv.ParseFloat(pos.AvgPx, 64)
		markPx, _ := strconv.ParseFloat(pos.MarkPx, 64)
		upl, _ := strconv.ParseFloat(pos.Upl, 64)
		uplRatio, _ := strconv.ParseFloat(pos.UplRatio, 64)
		lever, _ := strconv.ParseFloat(pos.Lever, 64)
		notional, _ := strconv.ParseFloat(pos.Notional, 64)

		position := map[string]interface{}{
			"symbol":          convertOKXSymbol(pos.InstId),
			"position_side":   strings.ToUpper(pos.PosSide),
			"position_amount": posSize,
			"entry_price":     avgPx,
			"mark_price":      markPx,
			"unrealized_pnl":  upl,
			"pnl_ratio":       uplRatio * 100,
			"leverage":        lever,
			"notional":        notional,
		}
		positions = append(positions, position)
	}

	// 更新缓存
	t.positionsCacheMutex.Lock()
	t.cachedPositions = positions
	t.positionsCacheTime = time.Now()
	t.positionsCacheMutex.Unlock()

	log.Printf("✓ 已获取OKX持仓信息: %d个持仓", len(positions))
	return positions, nil
}

// GetPrice 获取最新价格
func (t *OKXTrader) GetPrice(symbol string) (float64, error) {
	instId := convertToOKXSymbol(symbol)
	endpoint := fmt.Sprintf("/api/v5/market/ticker?instId=%s", instId)
	
	data, err := t.request("GET", endpoint, "")
	if err != nil {
		return 0, fmt.Errorf("获取价格失败: %w", err)
	}

	var response struct {
		Code string `json:"code"`
		Msg  string `json:"msg"`
		Data []struct {
			Last string `json:"last"`
		} `json:"data"`
	}

	if err := json.Unmarshal(data, &response); err != nil {
		return 0, fmt.Errorf("解析价格数据失败: %w", err)
	}

	if response.Code != "0" || len(response.Data) == 0 {
		return 0, fmt.Errorf("OKX API错误: %s", response.Msg)
	}

	price, err := strconv.ParseFloat(response.Data[0].Last, 64)
	if err != nil {
		return 0, fmt.Errorf("解析价格失败: %w", err)
	}

	return price, nil
}

// OpenPosition 开仓
func (t *OKXTrader) OpenPosition(symbol, side string, quantity, leverage float64) error {
	instId := convertToOKXSymbol(symbol)
	
	// 设置杠杆
	if err := t.setLeverage(instId, leverage); err != nil {
		log.Printf("⚠️ 设置杠杆失败: %v", err)
	}

	// 构建订单参数
	posSide := "long"
	if side == "SELL" {
		posSide = "short"
	}

	orderData := map[string]interface{}{
		"instId":  instId,
		"tdMode":  "cross", // 全仓模式
		"side":    side,
		"posSide": posSide,
		"ordType": "market",
		"sz":      fmt.Sprintf("%.0f", quantity),
	}

	body, _ := json.Marshal([]interface{}{orderData})
	data, err := t.request("POST", "/api/v5/trade/order", string(body))
	if err != nil {
		return fmt.Errorf("开仓失败: %w", err)
	}

	var response struct {
		Code string `json:"code"`
		Msg  string `json:"msg"`
		Data []struct {
			OrdId string `json:"ordId"`
			SCode string `json:"sCode"`
			SMsg  string `json:"sMsg"`
		} `json:"data"`
	}

	if err := json.Unmarshal(data, &response); err != nil {
		return fmt.Errorf("解析开仓响应失败: %w", err)
	}

	if response.Code != "0" || (len(response.Data) > 0 && response.Data[0].SCode != "0") {
		return fmt.Errorf("OKX开仓失败: %s", response.Msg)
	}

	log.Printf("✓ OKX开仓成功: %s %s %.0f张", symbol, side, quantity)
	return nil
}

// ClosePosition 平仓
func (t *OKXTrader) ClosePosition(symbol, positionSide string, quantity float64) error {
	instId := convertToOKXSymbol(symbol)
	
	// 确定平仓方向
	side := "sell"
	if positionSide == "SHORT" {
		side = "buy"
	}

	orderData := map[string]interface{}{
		"instId":  instId,
		"tdMode":  "cross",
		"side":    side,
		"posSide": strings.ToLower(positionSide),
		"ordType": "market",
		"sz":      fmt.Sprintf("%.0f", quantity),
	}

	body, _ := json.Marshal([]interface{}{orderData})
	data, err := t.request("POST", "/api/v5/trade/order", string(body))
	if err != nil {
		return fmt.Errorf("平仓失败: %w", err)
	}

	var response struct {
		Code string `json:"code"`
		Msg  string `json:"msg"`
	}

	if err := json.Unmarshal(data, &response); err != nil {
		return fmt.Errorf("解析平仓响应失败: %w", err)
	}

	if response.Code != "0" {
		return fmt.Errorf("OKX平仓失败: %s", response.Msg)
	}

	log.Printf("✓ OKX平仓成功: %s %s %.0f张", symbol, positionSide, quantity)
	return nil
}

// SetStopLoss 设置止损单
func (t *OKXTrader) SetStopLoss(symbol string, positionSide string, quantity, stopPrice float64) error {
	instId := convertToOKXSymbol(symbol)
	
	// 确定止损方向
	side := "sell"
	if positionSide == "SHORT" {
		side = "buy"
	}

	orderData := map[string]interface{}{
		"instId":     instId,
		"tdMode":     "cross",
		"side":       side,
		"posSide":    strings.ToLower(positionSide),
		"ordType":    "conditional",
		"sz":         fmt.Sprintf("%.0f", quantity),
		"slTriggerPx": fmt.Sprintf("%.2f", stopPrice),
		"slOrdPx":    "-1", // 市价止损
	}

	body, _ := json.Marshal([]interface{}{orderData})
	data, err := t.request("POST", "/api/v5/trade/order-algo", string(body))
	if err != nil {
		return fmt.Errorf("设置止损失败: %w", err)
	}

	var response struct {
		Code string `json:"code"`
		Msg  string `json:"msg"`
	}

	if err := json.Unmarshal(data, &response); err != nil {
		return fmt.Errorf("解析止损响应失败: %w", err)
	}

	if response.Code != "0" {
		return fmt.Errorf("OKX设置止损失败: %s", response.Msg)
	}

	log.Printf("✓ OKX止损设置成功: %s %s 止损价=%.2f", symbol, positionSide, stopPrice)
	return nil
}

// CancelStopOrders 取消止盈止损单
func (t *OKXTrader) CancelStopOrders(symbol string) error {
	instId := convertToOKXSymbol(symbol)
	
	// 获取所有算法单
	endpoint := fmt.Sprintf("/api/v5/trade/orders-algo-pending?instType=SWAP&instId=%s", instId)
	data, err := t.request("GET", endpoint, "")
	if err != nil {
		return fmt.Errorf("获取算法单失败: %w", err)
	}

	var response struct {
		Code string `json:"code"`
		Data []struct {
			AlgoId string `json:"algoId"`
		} `json:"data"`
	}

	if err := json.Unmarshal(data, &response); err != nil {
		return err
	}

	// 取消所有算法单
	for _, order := range response.Data {
		cancelData := []map[string]interface{}{
			{
				"instId": instId,
				"algoId": order.AlgoId,
			},
		}
		body, _ := json.Marshal(cancelData)
		t.request("POST", "/api/v5/trade/cancel-algos", string(body))
	}

	log.Printf("✓ 已取消 %s 的所有止盈止损单", symbol)
	return nil
}

// setLeverage 设置杠杆倍数
func (t *OKXTrader) setLeverage(instId string, leverage float64) error {
	leverageData := []map[string]interface{}{
		{
			"instId":  instId,
			"lever":   fmt.Sprintf("%.0f", leverage),
			"mgnMode": "cross",
		},
	}

	body, _ := json.Marshal(leverageData)
	data, err := t.request("POST", "/api/v5/account/set-leverage", string(body))
	if err != nil {
		return err
	}

	var response struct {
		Code string `json:"code"`
		Msg  string `json:"msg"`
	}

	if err := json.Unmarshal(data, &response); err != nil {
		return err
	}

	if response.Code != "0" {
		return fmt.Errorf("设置杠杆失败: %s", response.Msg)
	}

	return nil
}

// CancelAllOrders 取消所有订单
func (t *OKXTrader) CancelAllOrders(symbol string) error {
	// 取消普通订单
	instId := convertToOKXSymbol(symbol)
	
	// 获取所有未完成订单
	endpoint := fmt.Sprintf("/api/v5/trade/orders-pending?instType=SWAP&instId=%s", instId)
	data, err := t.request("GET", endpoint, "")
	if err != nil {
		return fmt.Errorf("获取订单失败: %w", err)
	}

	var response struct {
		Code string `json:"code"`
		Data []struct {
			OrdId string `json:"ordId"`
		} `json:"data"`
	}

	if err := json.Unmarshal(data, &response); err != nil {
		return err
	}

	// 取消所有订单
	for _, order := range response.Data {
		cancelData := []map[string]interface{}{
			{
				"instId": instId,
				"ordId":  order.OrdId,
			},
		}
		body, _ := json.Marshal(cancelData)
		t.request("POST", "/api/v5/trade/cancel-order", string(body))
	}

	// 同时取消止盈止损单
	t.CancelStopOrders(symbol)

	log.Printf("✓ 已取消 %s 的所有订单", symbol)
	return nil
}

// CancelStopLossOrders 取消止损单（OKX无法区分止损和止盈，取消所有算法单）
func (t *OKXTrader) CancelStopLossOrders(symbol string) error {
	return t.CancelStopOrders(symbol)
}

// CancelTakeProfitOrders 取消止盈单（OKX无法区分止损和止盈，取消所有算法单）
func (t *OKXTrader) CancelTakeProfitOrders(symbol string) error {
	return t.CancelStopOrders(symbol)
}

// OpenLong 开多仓
func (t *OKXTrader) OpenLong(symbol string, quantity float64, leverage int) (map[string]interface{}, error) {
	if err := t.OpenPosition(symbol, "buy", quantity, float64(leverage)); err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"symbol":   symbol,
		"side":     "LONG",
		"quantity": quantity,
		"leverage": leverage,
	}, nil
}

// OpenShort 开空仓
func (t *OKXTrader) OpenShort(symbol string, quantity float64, leverage int) (map[string]interface{}, error) {
	if err := t.OpenPosition(symbol, "sell", quantity, float64(leverage)); err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"symbol":   symbol,
		"side":     "SHORT",
		"quantity": quantity,
		"leverage": leverage,
	}, nil
}

// CloseLong 平多仓
func (t *OKXTrader) CloseLong(symbol string, quantity float64) (map[string]interface{}, error) {
	if err := t.ClosePosition(symbol, "LONG", quantity); err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"symbol":   symbol,
		"side":     "CLOSE_LONG",
		"quantity": quantity,
	}, nil
}

// CloseShort 平空仓
func (t *OKXTrader) CloseShort(symbol string, quantity float64) (map[string]interface{}, error) {
	if err := t.ClosePosition(symbol, "SHORT", quantity); err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"symbol":   symbol,
		"side":     "CLOSE_SHORT",
		"quantity": quantity,
	}, nil
}

// SetLeverage 设置杠杆
func (t *OKXTrader) SetLeverage(symbol string, leverage int) error {
	instId := convertToOKXSymbol(symbol)
	return t.setLeverage(instId, float64(leverage))
}

// SetMarginMode 设置仓位模式（OKX默认全仓）
func (t *OKXTrader) SetMarginMode(symbol string, isCrossMargin bool) error {
	// OKX 的保证金模式设置较复杂，这里简化处理
	// 默认使用全仓模式
	if !isCrossMargin {
		log.Printf("⚠️ OKX暂不支持逐仓模式，将使用全仓模式")
	}
	return nil
}

// GetMarketPrice 获取市场价格
func (t *OKXTrader) GetMarketPrice(symbol string) (float64, error) {
	return t.GetPrice(symbol)
}

// SetTakeProfit 设置止盈单
func (t *OKXTrader) SetTakeProfit(symbol string, positionSide string, quantity, takeProfitPrice float64) error {
	instId := convertToOKXSymbol(symbol)
	
	// 确定止盈方向
	side := "sell"
	if positionSide == "SHORT" {
		side = "buy"
	}

	orderData := map[string]interface{}{
		"instId":      instId,
		"tdMode":      "cross",
		"side":        side,
		"posSide":     strings.ToLower(positionSide),
		"ordType":     "conditional",
		"sz":          fmt.Sprintf("%.0f", quantity),
		"tpTriggerPx": fmt.Sprintf("%.2f", takeProfitPrice),
		"tpOrdPx":     "-1", // 市价止盈
	}

	body, _ := json.Marshal([]interface{}{orderData})
	data, err := t.request("POST", "/api/v5/trade/order-algo", string(body))
	if err != nil {
		return fmt.Errorf("设置止盈失败: %w", err)
	}

	var response struct {
		Code string `json:"code"`
		Msg  string `json:"msg"`
	}

	if err := json.Unmarshal(data, &response); err != nil {
		return fmt.Errorf("解析止盈响应失败: %w", err)
	}

	if response.Code != "0" {
		return fmt.Errorf("OKX设置止盈失败: %s", response.Msg)
	}

	log.Printf("✓ OKX止盈设置成功: %s %s 止盈价=%.2f", symbol, positionSide, takeProfitPrice)
	return nil
}

// FormatQuantity 格式化数量
func (t *OKXTrader) FormatQuantity(symbol string, quantity float64) (string, error) {
	// OKX 合约数量通常为整数（张数）
	return fmt.Sprintf("%.0f", quantity), nil
}

// 辅助函数：转换币种符号
func convertToOKXSymbol(symbol string) string {
	// BTCUSDT -> BTC-USDT-SWAP
	if strings.HasSuffix(symbol, "USDT") {
		base := strings.TrimSuffix(symbol, "USDT")
		return base + "-USDT-SWAP"
	}
	return symbol
}

func convertOKXSymbol(instId string) string {
	// BTC-USDT-SWAP -> BTCUSDT
	parts := strings.Split(instId, "-")
	if len(parts) >= 2 {
		return parts[0] + parts[1]
	}
	return instId
}
