package model

import "github.com/joho/godotenv"

func init() {
	// 按照词法顺序保证在db.go与redis.go的init之前执行
	_ = godotenv.Load(".env")
}
