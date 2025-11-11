package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

func main() {
	// 从数据库读取的配置
	apiKey := "qzfYzFjtj7qy39TuEeAWIzCF1q22mjEfOFKu9E1oZNkArlX2zrQpWhTt92BVMbNt"
	secretKey := "kG5oZio84XFxWPlA9u9R6mP1xSybXZJnq21NAGaOeZBqNdtUAWcHdYniNJs2PGeS"
	isTestnet := true
	
	fmt.Println("🧪 Binance API 测试工具")
	fmt.Println("=" + string(make([]byte, 60)) + "=")
	fmt.Printf("API Key: %s...\n", apiKey[:20])
	fmt.Printf("Secret Key: %s...\n", secretKey[:20])
	fmt.Printf("环境: %s\n\n", map[bool]string{true: "测试网", false: "实盘"}[isTestnet])
	
	// 测试 1: 服务器时间（无需签名）
	fmt.Println("📊 测试 1: 获取服务器时间（公共接口）")
	testServerTime(isTestnet)
	
	// 测试 2: 账户信息（需要签名）
	fmt.Println("\n📊 测试 2: 获取账户信息（私有接口）")
	testAccountInfo(apiKey, secretKey, isTestnet)
}

func testServerTime(isTestnet bool) {
	baseURL := "https://fapi.binance.com"
	if isTestnet {
		baseURL = "https://testnet.binancefuture.com"
	}
	
	url := baseURL + "/fapi/v1/time"
	resp, err := http.Get(url)
	if err != nil {
		fmt.Printf("❌ 请求失败: %v\n", err)
		return
	}
	defer resp.Body.Close()
	
	data, _ := io.ReadAll(resp.Body)
	
	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		fmt.Printf("❌ 解析失败: %v\n", err)
		return
	}
	
	if serverTime, ok := result["serverTime"].(float64); ok {
		t := time.Unix(int64(serverTime)/1000, 0)
		fmt.Printf("✅ 服务器时间: %s\n", t.Format("2006-01-02 15:04:05"))
		fmt.Printf("   时间戳: %.0f\n", serverTime)
	} else {
		fmt.Printf("❌ 响应: %s\n", string(data))
	}
}

func testAccountInfo(apiKey, secretKey string, isTestnet bool) {
	baseURL := "https://fapi.binance.com"
	if isTestnet {
		baseURL = "https://testnet.binancefuture.com"
	}
	
	// 生成时间戳
	timestamp := time.Now().UnixMilli()
	
	// 构建查询字符串
	queryString := fmt.Sprintf("timestamp=%d", timestamp)
	
	// 生成签名
	h := hmac.New(sha256.New, []byte(secretKey))
	h.Write([]byte(queryString))
	signature := hex.EncodeToString(h.Sum(nil))
	
	// 完整 URL
	url := fmt.Sprintf("%s/fapi/v2/account?%s&signature=%s", baseURL, queryString, signature)
	
	// 创建请求
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("X-MBX-APIKEY", apiKey)
	
	fmt.Printf("   请求 URL: %s\n", baseURL+"/fapi/v2/account")
	fmt.Printf("   时间戳: %d\n", timestamp)
	fmt.Printf("   签名: %s\n", signature[:20]+"...")
	
	// 发送请求
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("❌ 请求失败: %v\n", err)
		return
	}
	defer resp.Body.Close()
	
	data, _ := io.ReadAll(resp.Body)
	
	fmt.Printf("   HTTP 状态: %d\n", resp.StatusCode)
	
	if resp.StatusCode == 200 {
		var result map[string]interface{}
		if err := json.Unmarshal(data, &result); err == nil {
			fmt.Printf("✅ 成功!\n")
			if assets, ok := result["assets"].([]interface{}); ok {
				fmt.Printf("   账户资产数量: %d\n", len(assets))
				// 显示 USDT 余额
				for _, asset := range assets {
					if a, ok := asset.(map[string]interface{}); ok {
						if a["asset"] == "USDT" {
							fmt.Printf("   USDT 余额: %v\n", a["walletBalance"])
						}
					}
				}
			}
		}
	} else {
		var errResult map[string]interface{}
		if err := json.Unmarshal(data, &errResult); err == nil {
			code := errResult["code"]
			msg := errResult["msg"]
			fmt.Printf("❌ 失败: %v - %v\n", code, msg)
			
			// 错误代码说明
			if code == float64(-2015) {
				fmt.Println("\n💡 错误 -2015 可能的原因:")
				fmt.Println("   1. API Key 无效或已被删除")
				fmt.Println("   2. API Key 是实盘的，但在访问测试网（或相反）")
				fmt.Println("   3. IP 白名单限制（如果设置了）")
				fmt.Println("   4. API Key 权限不足（需要 Read 权限）")
			}
		} else {
			fmt.Printf("❌ 响应: %s\n", string(data))
		}
	}
}
