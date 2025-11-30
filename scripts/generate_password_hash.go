//go:build ignore

package main

import (
	"fmt"
	"log"
	"os"

	"golang.org/x/crypto/bcrypt"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("使用方法: go run generate_password_hash.go <密码>")
		fmt.Println("示例: go run generate_password_hash.go test123")
		os.Exit(1)
	}

	password := os.Args[1]

	// 生成bcrypt哈希
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("密码: %s\n", password)
	fmt.Printf("哈希: %s\n", string(hash))
	fmt.Println()
	fmt.Println("SQL语句:")
	fmt.Printf("UPDATE users SET password_hash = '%s' WHERE email = 'test@local.dev';\n", string(hash))
}
