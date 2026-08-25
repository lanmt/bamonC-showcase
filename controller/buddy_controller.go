package controller

import (
	"bamonC/model"
	"net/http"
	"strconv"

	"bamonC/service"

	"github.com/gin-gonic/gin"
)

type BuddyController struct {
	UserService  *service.UserService
	BuddyService *service.BuddyService
}

// GET /get_buddies/:username
func (c *BuddyController) GetBuddies(ctx *gin.Context) {
	username := ctx.Param("username")

	Buddy, err := c.BuddyService.GetBuddies(username)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if Buddy == nil {
		ctx.JSON(http.StatusOK, gin.H{"data": nil})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "success", "buddies": Buddy})
}

/*
	UserID    uint64 `gorm:"not null;column:user_id"`
	BuddyID   uint64 `gorm:"not null;column:buddy_id"`
	BuddyName string `gorm:"column:buddy_name"`
*/
// POST /add_buddy/:username
// 前端传入 JSON，例如：
// { "UserID": 1, "BuddyID": 2, "BuddyName": "小李"}
func (c *BuddyController) AddBuddy(ctx *gin.Context) {
	username := ctx.Param("username")

	var input model.Buddy
	if err := ctx.ShouldBindJSON(&input); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid JSON"})
		return
	}

	// 校验用户是否存在
	existing, err := c.UserService.GetByUsername(username)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if existing == nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "user not found"})
	}
	// 插入数据库
	if err := c.BuddyService.AddBuddy(&input); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// GORM 会自动把生成的 ID 填充到 input.ID 里
	ctx.JSON(http.StatusOK, gin.H{
		"message": "buddy added successfully",
		"buddy":   input,
	})
}

// DELETE /del_buddy/:username?id=123
func (c *BuddyController) DeleteBuddy(ctx *gin.Context) {
	idStr := ctx.Query("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	if err := c.BuddyService.DeleteBuddy(id); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "buddy deleted successfully"})
}
