package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

func main() {
	// 从数据库读取的配置
	// 请替换为你的实际 API 凭据
	apiKey := "1b2e1ad2-1c0c-478b-a1bf-0ceb215df962"
	secretKey := "2C89C15169F51E9F9664C7981F83126B"
	passphrase := "YLyl6561@"
	isDemo := true // 模拟盘
	
	fmt.Println("=" + string(make([]byte, 70)) + "=")
	fmt.Println("🧪 OKX API 测试工具")
	fmt.Println("=" + string(make([]byte, 70)) + "=")
	fmt.Printf("\n📝 配置信息:\n")
	fmt.Printf("   API Key: %s\n", apiKey)
	fmt.Printf("   Secret Key: %s\n", secretKey)
	fmt.Printf("   Passphrase: %s\n", passphrase)
	fmt.Printf("   模式: %s\n", map[bool]string{true: "模拟盘", false: "实盘"}[isDemo])
	
	// 测试模拟盘和实盘
	fmt.Println("\n" + string(make([]byte, 70)))
	fmt.Println("🧪 测试 1: 使用模拟盘模式")
	fmt.Println(string(make([]byte, 70)))
	testAccountBalance(apiKey, secretKey, passphrase, true)
	
	fmt.Println("\n" + string(make([]byte, 70)))
	fmt.Println("🧪 测试 2: 使用实盘模式")
	fmt.Println(string(make([]byte, 70)))
	testAccountBalance(apiKey, secretKey, passphrase, false)
	
	// 测试 2: 获取账户配置
	fmt.Println("\n⚙️  测试 2: 获取账户配置")
	testAccountConfig(apiKey, secretKey, passphrase, isDemo)
	
	// 测试 3: 获取持仓信息
	fmt.Println("\n📈 测试 3: 获取持仓信息")
	testPositions(apiKey, secretKey, passphrase, isDemo)
}

func testAccountBalance(apiKey, secretKey, passphrase string, isDemo bool) {
	makeRequest("GET", "/api/v5/account/balance", "", apiKey, secretKey, passphrase, isDemo)
}

func testAccountConfig(apiKey, secretKey, passphrase string, isDemo bool) {
	makeRequest("GET", "/api/v5/account/config", "", apiKey, secretKey, passphrase, isDemo)
}

func testPositions(apiKey, secretKey, passphrase string, isDemo bool) {
	makeRequest("GET", "/api/v5/account/positions", "", apiKey, secretKey, passphrase, isDemo)
}

func makeRequest(method, endpoint, body, apiKey, secretKey, passphrase string, isDemo bool) {
	// 生成时间戳 (ISO 8601 格式，带毫秒)
	timestamp := time.Now().UTC().Format("2006-01-02T15:04:05.000Z")
	
	// 生成签名
	// 根据 OKX 文档: timestamp + method + requestPath + body
	// method 必须大写 (GET, POST 等)
	// requestPath 是完整路径 (例如: /api/v5/account/balance)
	// body 如果为空则省略
	preHashString := timestamp + method + endpoint + body
	
	// 使用 HMAC SHA256 签名
	h := hmac.New(sha256.New, []byte(secretKey))
	h.Write([]byte(preHashString))
	
	// Base64 编码
	sign := base64.StdEncoding.EncodeToString(h.Sum(nil))
	
	// 调试信息
	fmt.Printf("   🔐 签名详情:\n")
	fmt.Printf("      Pre-hash: %s\n", preHashString)
	fmt.Printf("      Secret Key: %s\n", secretKey)
	fmt.Printf("      Signature: %s\n", sign)
	
	// 创建请求
	url := "https://www.okx.com" + endpoint
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		fmt.Printf("❌ 创建请求失败: %v\n", err)
		return
	}
	
	// 设置请求头
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("OK-ACCESS-KEY", apiKey)
	req.Header.Set("OK-ACCESS-SIGN", sign)
	req.Header.Set("OK-ACCESS-TIMESTAMP", timestamp)
	req.Header.Set("OK-ACCESS-PASSPHRASE", passphrase)
	
	// 模拟盘必须添加此头部
	if isDemo {
		req.Header.Set("x-simulated-trading", "1")
	}
	
	// 打印请求信息
	fmt.Printf("   请求: %s %s\n", method, endpoint)
	fmt.Printf("   时间戳: %s\n", timestamp)
	if isDemo {
		fmt.Printf("   模式: 模拟盘 (x-simulated-trading: 1)\n")
	} else {
		fmt.Printf("   模式: 实盘\n")
	}
	
	// 发送请求
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("❌ 请求失败: %v\n", err)
		return
	}
	defer resp.Body.Close()
	
	// 读取响应
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("❌ 读取响应失败: %v\n", err)
		return
	}
	
	// 解析响应
	var response map[string]interface{}
	if err := json.Unmarshal(data, &response); err != nil {
		fmt.Printf("❌ 解析响应失败: %v\n", err)
		fmt.Printf("   原始响应: %s\n", string(data))
		return
	}
	
	// 检查响应状态
	code, _ := response["code"].(string)
	msg, _ := response["msg"].(string)
	
	fmt.Printf("   HTTP 状态: %d\n", resp.StatusCode)
	fmt.Printf("   响应代码: %s\n", code)
	
	if code == "0" {
		fmt.Printf("✅ 成功!\n")
		if dataArray, ok := response["data"].([]interface{}); ok && len(dataArray) > 0 {
			prettyData, _ := json.MarshalIndent(dataArray[0], "   ", "  ")
			fmt.Printf("   数据: %s\n", string(prettyData))
		}
	} else {
		fmt.Printf("❌ 失败: %s (错误代码: %s)\n", msg, code)
		
		// 提供错误代码说明
		explainErrorCode(code)
	}
}

func explainErrorCode(code string) {
	errorCodes := map[string]string{
		"50119": "API Key 不存在 - 请检查 API Key 是否正确，或在 OKX 后台确认该 Key 是否已被删除",
		"50101": "API Key 与当前环境不匹配 - 实盘 Key 不能用于模拟盘，反之亦然",
		"50113": "无效的签名 - 请检查 Secret Key 是否正确",
		"50111": "无效的 Passphrase - 请检查 Passphrase 是否正确",
		"50102": "时间戳错误 - 服务器时间与本地时间相差超过 30 秒",
		"50103": "请求头缺失 - 缺少必需的请求头",
		"50104": "无效的 OK-ACCESS-KEY",
		"50105": "无效的 OK-ACCESS-TIMESTAMP",
		"50106": "无效的 OK-ACCESS-SIGN",
	}
	
	if explanation, ok := errorCodes[code]; ok {
		fmt.Printf("   💡 说明: %s\n", explanation)
	}
}
