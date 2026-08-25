package model

import (
	"context"
	"log"
	"os"

	"github.com/redis/go-redis/v9"
)

var (
	RDB *redis.Client
	Ctx = context.Background()
)

func init() {
	REDIS_IP := os.Getenv("REDIS_IP")
	REDIS_PORT := os.Getenv("REDIS_PORT")
	RDB = redis.NewClient(&redis.Options{
		Addr: REDIS_IP + ":" + REDIS_PORT, // Redis 地址
		DB:   1,                           // 默认 DB
	})

	// 测试连接
	_, err := RDB.Ping(Ctx).Result()
	if err != nil {
		log.Fatalf("❌ Redis连接失败: %v", err)
	} else {
		log.Println("✅ Redis连接成功")
	}
}
