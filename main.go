package main

import (
	"bamonC/model"
	"bamonC/routers"
	"bamonC/task"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

func main() {

	r := gin.Default()

	// bound router
	routers.InitRouter(r)
	// start corn task
	task.CreateCornTask()

	// Initialize default admin if not exists
	var count int64
	model.DB.Model(&model.User{}).Where("role = ?", "admin").Count(&count)
	if count == 0 {
		defaultPassword := os.Getenv("INITIAL_ADMIN_PASSWORD")
		if defaultPassword == "" {
			defaultPassword = "change-me-before-deployment"
		}
		hashed, _ := bcrypt.GenerateFromPassword([]byte(defaultPassword), bcrypt.DefaultCost)
		model.DB.Create(&model.User{
			Username: "admin",
			Password: string(hashed),
			Role:     "admin",
		})
		log.Println("Default admin user created; configure INITIAL_ADMIN_PASSWORD before deployment")
	}

	port := os.Getenv("BC_PORT")
	log.Println("服务启动，监听端口 ", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("服务启动失败: %v", err)
	}
}
