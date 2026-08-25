package controller

import (
	"bamonC/model"
	"bamonC/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

type UserController struct {
	Service *service.UserService
}

// GET /get_user/:username  获取用户信息
func (c *UserController) GetUser(ctx *gin.Context) {
	username := ctx.Param("username")

	user, err := c.Service.GetByUsername(username)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if user == nil {
		ctx.JSON(http.StatusOK, gin.H{"data": nil})
		return
	}
	responseUser := *user
	responseUser.Auth = ""
	ctx.JSON(http.StatusOK, gin.H{"message": "success", "data": responseUser})
}

// POST /create_user/:username  创建用户
type CreateUserReq struct {
	Password string `json:"password" binding:"required"`
}

func (c *UserController) CreateUser(ctx *gin.Context) {
	username := ctx.Param("username")

	var req CreateUserReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "参数错误，需要password"})
		return
	}

	err := c.Service.CreateUser(username, req.Password)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "user created", "user": username})
}

// POST /update_user/:username  修改用户设置
func (c *UserController) UpdateUser(ctx *gin.Context) {
	username := ctx.Param("username")
	var input model.User
	if err := ctx.ShouldBindJSON(&input); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := c.Service.UpdateUser(username, &input); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "success"})
}

// DELETE /delelte_user/:username  删除用户
func (c *UserController) DeleteUser(ctx *gin.Context) {
	username := ctx.Param("username")

	if err := c.Service.DeleteUser(username); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "user deleted"})
}
