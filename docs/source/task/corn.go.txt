package task

import (
	"log"

	"github.com/robfig/cron/v3"
)

var globalCron *cron.Cron

func CreateCornTask() {
	RestartCronTasks()
}

// RestartCronTasks 加载配置并重新调度定时任务
func RestartCronTasks() {
	if globalCron != nil {
		log.Println("检测到配置更新，停止旧的定时任务调度")
		globalCron.Stop()
	}

	globalCron = cron.New(cron.WithSeconds())

	// 尝试从DB或者内存获取配置，如果获取不到则用默认
	captchaCron, _ := SystemConfigService.GetConfig("captcha_task_cron", "00 58 11 * * *")
	updateCron, _ := SystemConfigService.GetConfig("update_time_id_cron", "00 00 5 * * *")

	log.Printf("重启定时任务... ReserveTask: [%s] | Update: [%s]\n", captchaCron, updateCron)

	globalCron.AddFunc(captchaCron, CaptchaTask)
	globalCron.AddFunc(updateCron, UpdateTimeIdTask)
	
	globalCron.Start()
}
