package routers

import (
	"bamonC/controller"
	"bamonC/middleware"
	"bamonC/model"
	"bamonC/service"
	"bamonC/task"

	"github.com/gin-gonic/gin"
)

// InitRouter 统一注册所有模块路由
func InitRouter(r *gin.Engine) {

	db := model.DB

	userService := &service.UserService{DB: db}
	courtService := &service.CourtService{DB: db}
	buddyService := &service.BuddyService{DB: db}
	captchaLogService := &service.CaptchaLogService{DB: db}
	systemConfigService := &service.SystemConfigService{DB: db}

	userController := &controller.UserController{Service: userService}
	courtController := &controller.CourtController{UserService: userService, CourtService: courtService}
	buddyController := &controller.BuddyController{UserService: userService, BuddyService: buddyService}
	captchaLogController := &controller.CaptchaLogController{LogService: captchaLogService}
	systemConfigController := &controller.SystemConfigController{ConfigService: systemConfigService}

	authController := &controller.AuthController{}

	// -----------------------------
	// 静态文件和模板
	// -----------------------------
	r.LoadHTMLGlob("templates/*.html") // HTML 模板
	r.Static("/static", "./static")    // JS/CSS/图片等静态文件

	// -----------------------------
	// 页面路由（前端可以依赖 /api/check_session 决定渲染）
	// -----------------------------
	r.GET("/", func(c *gin.Context) { c.HTML(200, "index.html", gin.H{}) })
	r.GET("/login.html", func(c *gin.Context) { c.HTML(200, "login.html", gin.H{}) })
	r.GET("/user_detail.html", func(c *gin.Context) { c.HTML(200, "user_detail.html", gin.H{}) })
	r.GET("/captcha_logs.html", func(c *gin.Context) { c.HTML(200, "captcha_logs.html", gin.H{}) })
	r.GET("/settings.html", func(c *gin.Context) { c.HTML(200, "settings.html", gin.H{}) })

	// -----------------------------
	// API 路由
	// -----------------------------
	api := r.Group("/api")
	
	// 公开的认证接口
	api.POST("/login", authController.Login)
	api.POST("/logout", authController.Logout)
	
	// 需要登录的接口
	auth := api.Group("/")
	auth.Use(middleware.AuthMiddleware())
	auth.GET("/check_session", authController.CheckSession)
	
	// 管理员专属接口
	admin := auth.Group("/")
	admin.Use(middleware.AdminMiddleware())
	admin.POST("/create_user/:username", userController.CreateUser)
	admin.DELETE("/del_user/:username", userController.DeleteUser)
	admin.GET("/system_config", systemConfigController.GetConfigs)
	admin.POST("/system_config", systemConfigController.UpdateConfig)
	admin.POST("/reset_admin_password", authController.ResetAdminPassword)
	
	// 只能操作自己（或者管理员操作任何人）的接口
	userMatched := auth.Group("/")
	userMatched.Use(middleware.UserMatchMiddleware())
	userMatched.GET("/get_user/:username", userController.GetUser)
	userMatched.POST("/update_user/:username", userController.UpdateUser)
	
	userMatched.GET("/get_courts/:username", courtController.GetCourts)
	userMatched.POST("/add_court/:username", courtController.AddCourt)
	userMatched.DELETE("/del_court/:username", courtController.DeleteCourt)
	
	userMatched.GET("/get_buddies/:username", buddyController.GetBuddies)
	userMatched.POST("/add_buddy/:username", buddyController.AddBuddy)
	userMatched.DELETE("/del_buddy/:username", buddyController.DeleteBuddy)
	
	userMatched.GET("/captcha_logs/:username", captchaLogController.GetLogs)
	userMatched.GET("/submit_test/:username", task.TestTask)
	userMatched.GET("/sync_buddy/:username", task.SyncBuddiesTask)

	// 管理员查看所有日志
	admin.GET("/captcha_logs", captchaLogController.GetAllLogs)
}
