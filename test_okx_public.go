package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

func main() {
	fmt.Println("🧪 测试 OKX 公共 API（无需认证）")
	fmt.Println("=" + string(make([]byte, 60)) + "=")
	
	// 测试公共 API - 获取系统时间
	fmt.Println("\n⏰ 测试 1: 获取 OKX 服务器时间")
	testPublicTime()
	
	// 测试公共 API - 获取交易对信息
	fmt.Println("\n📊 测试 2: 获取 BTC-USDT 交易对信息")
	testPublicInstruments()
}

func testPublicTime() {
	url := "https://www.okx.com/api/v5/public/time"
	
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		fmt.Printf("❌ 请求失败: %v\n", err)
		return
	}
	defer resp.Body.Close()
	
	data, _ := io.ReadAll(resp.Body)
	
	var response map[string]interface{}
	if err := json.Unmarshal(data, &response); err != nil {
		fmt.Printf("❌ 解析失败: %v\n", err)
		return
	}
	
	code, _ := response["code"].(string)
	if code == "0" {
		if dataArray, ok := response["data"].([]interface{}); ok && len(dataArray) > 0 {
			if timeData, ok := dataArray[0].(map[string]interface{}); ok {
				serverTime := timeData["ts"].(string)
				fmt.Printf("✅ OKX 服务器时间: %s\n", serverTime)
				
				// 转换为可读时间
				ts, _ := time.Parse("1504", serverTime[:len(serverTime)-3])
				fmt.Printf("   可读时间: %s\n", ts.Format("2006-01-02 15:04:05"))
			}
		}
	} else {
		fmt.Printf("❌ 失败: %v\n", response)
	}
}

func testPublicInstruments() {
	url := "https://www.okx.com/api/v5/public/instruments?instType=SWAP&instId=BTC-USDT-SWAP"
	
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		fmt.Printf("❌ 请求失败: %v\n", err)
		return
	}
	defer resp.Body.Close()
	
	data, _ := io.ReadAll(resp.Body)
	
	var response map[string]interface{}
	if err := json.Unmarshal(data, &response); err != nil {
		fmt.Printf("❌ 解析失败: %v\n", err)
		return
	}
	
	code, _ := response["code"].(string)
	if code == "0" {
		fmt.Printf("✅ 成功获取交易对信息\n")
		if dataArray, ok := response["data"].([]interface{}); ok && len(dataArray) > 0 {
			prettyData, _ := json.MarshalIndent(dataArray[0], "   ", "  ")
			fmt.Printf("   数据: %s\n", string(prettyData))
		}
	} else {
		fmt.Printf("❌ 失败: %v\n", response)
	}
}
