package controller

import (
	"bamonC/service"
	"bamonC/task"
	"net/http"

	"github.com/gin-gonic/gin"
)

type SystemConfigController struct {
	ConfigService *service.SystemConfigService
}

// GetConfigs 返回前端所有配置
func (c *SystemConfigController) GetConfigs(ctx *gin.Context) {
	configs, err := c.ConfigService.GetAllConfigs()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	
	// 如果没有任何配置，确保向前端补充些缺省键以便渲染
	if _, ok := configs["captcha_task_cron"]; !ok {
		// 为了给前端展示，可以默认填充
		configs["captcha_task_cron"], _ = c.ConfigService.GetConfig("captcha_task_cron", "00 58 11 * * *")
		configs["update_time_id_cron"], _ = c.ConfigService.GetConfig("update_time_id_cron", "00 00 5 * * *")
		configs["captcha_cutoff_time"], _ = c.ConfigService.GetConfig("captcha_cutoff_time", "23:00:00")
	}

	ctx.JSON(http.StatusOK, gin.H{"configs": configs})
}

// UpdateConfig 接受前端的修改请求
func (c *SystemConfigController) UpdateConfig(ctx *gin.Context) {
	var body map[string]string
	if err := ctx.ShouldBindJSON(&body); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": "参数格式错误"})
		return
	}

	for k, v := range body {
		if k == "" || v == "" {
			continue
		}
		if err := c.ConfigService.SetConfig(k, v); err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
			return
		}
	}

	// 同步让定时任务重新加载配置
	task.RestartCronTasks()

	ctx.JSON(http.StatusOK, gin.H{"message": "配置更新成功"})
}
