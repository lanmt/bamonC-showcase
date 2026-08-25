package task

import (
	"bamonC/model"
	"bamonC/request"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/tidwall/gjson"
)

func CaptchaTask() {
	log.Println("开始场地级预约抢票合并任务(原 CaptchaTask + SubmitTask)")
	users, err := UserService.GetAllEnabled()

	if err != nil {
		log.Println("CaptchaTask error: ", err)
	}

	var wg sync.WaitGroup
	for i := range users {
		user := &users[i]
		wg.Add(1)
		go func(user *model.User) {
			defer wg.Done()
			username := user.Username

			courts, err := CourtService.GetCourts(username)
			if err != nil {
				log.Println(username, "::GetCourts: ", err)
				return
			}

			// 以场地为单位并行
			var courtWg sync.WaitGroup
			for _, court := range courts {
				courtWg.Add(1)
				go func(c model.Court) {
					defer courtWg.Done()

					for {
						cutoffStr, _ := SystemConfigService.GetConfig("captcha_cutoff_time", "23:00:00")
						t, parseErr := time.Parse("15:04:05", cutoffStr)
						if parseErr != nil {
							// 支持只传 23:00 的格式
							t, parseErr = time.Parse("15:04", cutoffStr)
							if parseErr != nil {
								t, _ = time.Parse("15:04:05", "23:00:00")
							}
						}
						now := time.Now()
						cutoffTime := time.Date(
							now.Year(), now.Month(), now.Day(),
							t.Hour(), t.Minute(), t.Second(), 0, now.Location(),
						)

						// 如果当前时间超出配置的截止下限时间，则停止当前场地的获取队列
						if now.After(cutoffTime) {
							log.Printf("%s 场地 [%v] 超过了设定的验证码停止执行时间: %s\n", username, c.CourtID, cutoffStr)
							break
						}

						captchaRes, err := request.GetCaptcha(user, username)
						if captchaRes == "" || err != nil {
							log.Println(username, "GetCaptcha频繁，附带错误: ", err)
							CaptchaLogService.SaveStepLog(username, "GetCaptcha_Task", "Error", fmt.Sprintf("获取验证码频繁/错误: %v", err))
							// 出现错误等待两秒
							time.Sleep(2 * time.Second)
							continue
						}

						go func(baseImg string) {
							imgBase64 := gjson.Get(baseImg, "repData.originalImageBase64").String()
							wordList := gjson.Get(baseImg, "repData.wordList").String()
							if imgBase64 != "" {
								if logErr := CaptchaLogService.SaveRawCaptchaImage(username, imgBase64, wordList); logErr != nil {
									log.Println(username, "::SaveRawCaptchaImage:", logErr)
								}
							}
						}(captchaRes)

						coords, err := request.GetCoords(user, captchaRes)
						if coords == "" || err != nil {
							log.Println(username, "::GetCoords: ", err)
							CaptchaLogService.SaveStepLog(username, "GetCoords_Task", "Error", fmt.Sprintf("GetCoords坐标计算失败: %v", err))
							// 识别出错立即重新获取
							continue
						}
						
						cleanedCoords := strings.ReplaceAll(strings.ReplaceAll(strings.ReplaceAll(coords, " ", ""), "\n", ""), "\t", "")

						secretKey := gjson.Get(captchaRes, "repData.secretKey").String()
						captchaToken := gjson.Get(captchaRes, "repData.token").String()

						e, err := request.CheckCaptcha(user, gjson.Get(cleanedCoords, "result").String(), secretKey, captchaToken)

						// ---- 保存验证码坐标Log并附加校验结果（异步，不阻塞主流程）----
						go func(baseImg, cleaned string, cerr error) {
							imgBase64 := gjson.Get(baseImg, "repData.originalImageBase64").String()
							wordList := gjson.Get(baseImg, "repData.wordList").String()
							if imgBase64 != "" && cleaned != "" {
								if logErr := CaptchaLogService.SaveCaptchaLog(username, imgBase64, wordList, cleaned, cerr); logErr != nil {
									log.Println(username, "::SaveCaptchaLog:", logErr)
								}
							}
						}(captchaRes, cleanedCoords, err)

						if e == "" || err != nil {
							log.Println(username, "::CheckCaptcha: ", err)
							// 识别报错后立即重新获取
							continue
						}

						cleanedE := strings.ReplaceAll(strings.ReplaceAll(strings.ReplaceAll(e, " ", ""), "\n", ""), "\t", "")
						if cleanedE != "" {
							// 验证通过，马上发送 submit
							tradeNo, submitErr := request.SubmitReservation(user, &c, cleanedE, captchaToken)
							if submitErr != nil {
								log.Println(username, "::SubmitReservation: ", submitErr)
								CaptchaLogService.SaveStepLog(username, "SubmitReservation_Task", "Error", fmt.Sprintf("预约场地 %v 失败: %v", c.CourtID, submitErr))
								// 发送成功（返回失败结果依然算发送完成），不再重复尝试获取
								break
							}
							
							log.Println(username, "::预约成功")
							CaptchaLogService.SaveStepLog(username, "SubmitReservation_Task", "Success", fmt.Sprintf("预约场地 %v 成功, 订单号: %s", c.CourtID, tradeNo))
							
							RedisService.SetUserOrder(username, tradeNo)
							
							if user.RelayEnabled {
								log.Println(username, "开始接力任务")
								CaptchaLogService.SaveStepLog(username, "RelayTask", "Info", "触发接力任务")
								RelayTask(user, &c, tradeNo, user.RelayUntil)
							}
							break // 已经成功发送并预定成功，跳出循环
						}
					}
				}(court)
			}
			courtWg.Wait() // 等待该用户下所有场地都处理完成

		}(user)
	}
	wg.Wait()
}
