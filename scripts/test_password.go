package main

import (
	"fmt"
	"log"

	"golang.org/x/crypto/bcrypt"
)

func main() {
	password := "test123"
	hash := "$2a$10$aT1VB5d99Do30MzhBK.b6uUIZu/N/d8WUTnZbsXodykwuseQr9fsC"

	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	if err != nil {
		log.Printf("❌ 密码验证失败: %v", err)
		fmt.Println("密码不匹配")
	} else {
		log.Printf("✅ 密码验证成功")
		fmt.Println("密码匹配")
	}
}
