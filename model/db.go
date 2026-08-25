package model

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"time"
	_ "time/tzdata"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func init() {
	// 读取环境变量，如果不存在则使用默认值
	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	dbname := os.Getenv("DB_NAME")

	portInt, err := strconv.Atoi(port)
	if err != nil {
		log.Printf("invalid DB_PORT, using default 5432: %v", err)
		portInt = 5432
	}

	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%d sslmode=disable TimeZone=Asia/Shanghai",
		host, user, password, dbname, portInt,
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("failed to connect database:", err)
	}

	// 创建 schema bc，如果不存在
	if err := db.Exec("CREATE SCHEMA IF NOT EXISTS bc").Error; err != nil {
		log.Fatal("failed to create schema:", err)
	}

	// 自动迁移表（顺序：User -> Buddy -> Court）
	// 迁移数据表
	if err := db.AutoMigrate(
		&User{},
		&Court{},
		&Buddy{},
		&CaptchaLog{},
		&SystemConfig{},
	); err != nil {
		log.Fatal("failed to migrate tables:", err)
	}

	DB = db
	log.Println("Database initialized successfully!")
}

// 统一时间格式，可在需要时使用
func now() time.Time {
	return time.Now().Local()
}
