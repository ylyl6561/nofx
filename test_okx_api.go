package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"time"
)

func main() {
	// 从数据库读取的配置
	apiKey := "1b2e1ad2-1c0c-478b-a1bf-0ceb215df962"
	secretKey := "2C89C15169F51E9F9664C7981F83126B"
	passphrase := "YLyl6561@"
	
	// 测试获取账户余额
	timestamp := time.Now().UTC().Format("2006-01-02T15:04:05.000Z")
	method := "GET"
	requestPath := "/api/v5/account/balance"
	body := ""
	
	// 生成签名
	message := timestamp + method + requestPath + body
	h := hmac.New(sha256.New, []byte(secretKey))
	h.Write([]byte(message))
	sign := base64.StdEncoding.EncodeToString(h.Sum(nil))
	
	// 创建请求
	url := "https://www.okx.com" + requestPath
	req, _ := http.NewRequest(method, url, nil)
	
	// 设置必需的请求头
	req.Header.Set("OK-ACCESS-KEY", apiKey)
	req.Header.Set("OK-ACCESS-SIGN", sign)
	req.Header.Set("OK-ACCESS-TIMESTAMP", timestamp)
	req.Header.Set("OK-ACCESS-PASSPHRASE", passphrase)
	req.Header.Set("Content-Type", "application/json")
	// 模拟盘必须设置这个头部
	req.Header.Set("x-simulated-trading", "1")
	
	fmt.Printf("🔍 测试 OKX API\n")
	fmt.Printf("   URL: %s\n", url)
	fmt.Printf("   API Key: %s\n", apiKey)
	fmt.Printf("   Passphrase: %s\n", passphrase)
	fmt.Printf("   Timestamp: %s\n", timestamp)
	fmt.Printf("   Sign: %s\n\n", sign)
	
	// 发送请求
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("❌ 请求失败: %v\n", err)
		return
	}
	defer resp.Body.Close()
	
	data, _ := io.ReadAll(resp.Body)
	fmt.Printf("📥 响应状态: %d\n", resp.StatusCode)
	fmt.Printf("📥 响应内容: %s\n", string(data))
}
