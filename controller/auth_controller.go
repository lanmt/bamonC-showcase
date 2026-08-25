package controller

import (
	"bamonC/middleware"
	"bamonC/model"
	"net/http"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

type AuthController struct{}

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

func (ac *AuthController) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "参数错误"})
		return
	}

	var user model.User
	if err := model.DB.Where("username = ?", req.Username).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "用户名或密码错误"})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "用户名或密码错误"})
		return
	}

	token, err := middleware.GenerateToken(user.Username, user.Role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "生成Token失败"})
		return
	}

	c.SetCookie("jwt_token", token, 3600*24, "/", "", false, true)
	c.JSON(http.StatusOK, gin.H{"message": "登录成功", "role": user.Role, "username": user.Username})
}

func (ac *AuthController) Logout(c *gin.Context) {
	c.SetCookie("jwt_token", "", -1, "/", "", false, true)
	c.JSON(http.StatusOK, gin.H{"message": "登出成功"})
}

// CheckSession check if currently logged in
func (ac *AuthController) CheckSession(c *gin.Context) {
	username, _ := c.Get("username")
	role, _ := c.Get("role")
	c.JSON(http.StatusOK, gin.H{"username": username, "role": role})
}

type ResetAdminPasswordRequest struct {
	NewPassword string `json:"new_password" binding:"required"`
}

// ResetAdminPassword resets the admin password (only allowed by admin)
func (ac *AuthController) ResetAdminPassword(c *gin.Context) {
	var req ResetAdminPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "参数错误"})
		return
	}

	username, _ := c.Get("username")
	
	hashed, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "处理密码失败"})
		return
	}

	if err := model.DB.Model(&model.User{}).Where("username = ?", username).Update("password", string(hashed)).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "修改密码失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "密码修改成功"})
}
