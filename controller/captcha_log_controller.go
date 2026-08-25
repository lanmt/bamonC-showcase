package controller

import (
	"bamonC/service"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type CaptchaLogController struct {
	LogService *service.CaptchaLogService
}

// GetLogs 返回某用户的captcha日志（分页）
func (c *CaptchaLogController) GetLogs(ctx *gin.Context) {
	username := ctx.Param("username")
	dateStr := ctx.Query("date")
	stepName := ctx.Query("step")
	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(ctx.DefaultQuery("page_size", "20"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	logs, total, err := c.LogService.GetFilteredLogs(username, dateStr, stepName, page, pageSize)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{
		"total":     total,
		"page":      page,
		"page_size": pageSize,
		"logs":      logs,
	})
}

// GetAllLogs 返回所有用户的captcha日志（分页）
func (c *CaptchaLogController) GetAllLogs(ctx *gin.Context) {
	dateStr := ctx.Query("date")
	stepName := ctx.Query("step")
	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(ctx.DefaultQuery("page_size", "20"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	logs, total, err := c.LogService.GetFilteredLogs("", dateStr, stepName, page, pageSize)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{
		"total":     total,
		"page":      page,
		"page_size": pageSize,
		"logs":      logs,
	})
}
