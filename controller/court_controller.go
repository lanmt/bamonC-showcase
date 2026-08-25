package controller

import (
	"bamonC/model"
	"bamonC/service"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type CourtController struct {
	UserService  *service.UserService
	CourtService *service.CourtService
}

// GET /get_courts/:username
func (c *CourtController) GetCourts(ctx *gin.Context) {
	username := ctx.Param("username")

	courts, err := c.CourtService.GetCourts(username)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "success", "courts": courts})
}

// POST /add_court/:username
// 前端传入 JSON，例如：
// { "UserID": 1, "VenueSiteID": 2, "CourtID": 3, "Time1ID": 4, "Time2ID": 5 }
func (c *CourtController) AddCourt(ctx *gin.Context) {
	username := ctx.Param("username")

	var input model.Court
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
	if err := c.CourtService.AddCourt(&input); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// GORM 会自动把生成的 ID 填充到 input.ID 里
	ctx.JSON(http.StatusOK, gin.H{
		"message": "court added successfully",
		"court":   input,
	})
}

// DELETE /del_court/:username?id=123
func (c *CourtController) DeleteCourt(ctx *gin.Context) {
	idStr := ctx.Query("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	if err := c.CourtService.DeleteCourt(id); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "court deleted successfully"})
}
