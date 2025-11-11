package main

import (
	"context"
	"fmt"

	"github.com/adshao/go-binance/v2/futures"
)

func main() {
	apiKey := "qzfYzFjtj7qy39TuEeAWIzCF1q22mjEfOFKu9E1oZNkArlX2zrQpWhTt92BVMbNt"
	secretKey := "kG5oZio84XFxWPlA9u9R6mP1xSybXZJnq21NAGaOeZBqNdtUAWcHdYniNJs2PGeS"
	
	fmt.Println("🧪 测试 go-binance 库的 BaseURL 设置")
	fmt.Println("=" + string(make([]byte, 60)) + "=")
	
	// 创建客户端
	client := futures.NewClient(apiKey, secretKey)
	
	// 设置测试网 URL
	client.BaseURL = "https://testnet.binancefuture.com"
	
	fmt.Printf("✅ 设置 BaseURL: %s\n\n", client.BaseURL)
	
	// 测试获取账户信息
	fmt.Println("📊 测试获取账户信息...")
	account, err := client.NewGetAccountService().Do(context.Background())
	if err != nil {
		fmt.Printf("❌ 失败: %v\n", err)
		return
	}
	
	fmt.Printf("✅ 成功!\n")
	fmt.Printf("   总资产: %v USDT\n", account.TotalWalletBalance)
	fmt.Printf("   可用余额: %v USDT\n", account.AvailableBalance)
	fmt.Printf("   资产数量: %d\n", len(account.Assets))
	
	// 显示 USDT 余额
	for _, asset := range account.Assets {
		if asset.Asset == "USDT" {
			fmt.Printf("   USDT 余额: %s\n", asset.WalletBalance)
		}
	}
}
