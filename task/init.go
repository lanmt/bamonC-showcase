package task

import (
	"bamonC/model"
	"bamonC/service"
)

var UserService *service.UserService
var CourtService *service.CourtService
var BuddyService *service.BuddyService
var RedisService *service.RedisService
var CaptchaLogService *service.CaptchaLogService
var SystemConfigService *service.SystemConfigService

func init() {
	db := model.DB
	rdb := model.RDB
	ctx := model.Ctx

	UserService = &service.UserService{DB: db}
	CourtService = &service.CourtService{DB: db}
	BuddyService = &service.BuddyService{DB: db}
	RedisService = &service.RedisService{
		RDB: rdb,
		Ctx: ctx,
	}
	CaptchaLogService = &service.CaptchaLogService{DB: db}
	SystemConfigService = &service.SystemConfigService{DB: db}
}
