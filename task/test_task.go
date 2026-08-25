package task

import (
	"bamonC/model"
	"bamonC/request"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/tidwall/gjson"
)

func TestTask(ctx *gin.Context) {
	username := ctx.Param("username")

	// ---- 频率限制：每用户每分钟只能触发一次 ----
	allowed, err := RedisService.CheckAndSetTestRateLimit(username)
	if err != nil {
		log.Println(username, "::CheckAndSetTestRateLimit error:", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"message": "服务器内部错误，请稍后重试"})
		return
	}
	if !allowed {
		log.Println(username, "::测试触发频繁，拒绝本次请求")
		ctx.JSON(http.StatusTooManyRequests, gin.H{"message": "请稍等，每分钟只能触发一次测试"})
		return
	}

	user, _ := UserService.GetByUsername(username)
	courts, err := CourtService.GetCourts(username)
	if err != nil {
		log.Println(username, "::GetCourts: ", err)
		return
	}

	start := time.Now()
	log.Println(username, ":测试任务开启")
	
	var courtWg sync.WaitGroup
	cnt := 0
	for _, court := range courts {
		courtWg.Add(1)
		go func(c model.Court) {
			defer courtWg.Done()
			for {
				elapsed := time.Since(start)
				if elapsed >= 3*time.Minute || cnt>=10 {
					log.Println(username, "::3 minutes passed for court", c.CourtID, ", exiting loop.")
					break
				}
				
				captchaRes, err := request.GetCaptcha(user, username)
				if captchaRes == "" || err != nil {
					log.Println(username, "GetCaptcha频繁，附带错误: ", err)
					CaptchaLogService.SaveStepLog(username, "GetCaptcha", "Error", fmt.Sprintf("获取验证码频繁/错误: %v", err))
					// 出错等待两秒
					time.Sleep(2 * time.Second)
					cnt++;
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
					CaptchaLogService.SaveStepLog(username, "GetCoords", "Error", fmt.Sprintf("GetCoords坐标计算失败: %v", err))
					// 识别出错立即重新获取
					continue
				}
				
				cleanedCoords := strings.ReplaceAll(strings.ReplaceAll(strings.ReplaceAll(coords, " ", ""), "\n", ""), "\t", "")
				secretKey := gjson.Get(captchaRes, "repData.secretKey").String()
				captchaToken := gjson.Get(captchaRes, "repData.token").String()

				e, err := request.CheckCaptcha(user, gjson.Get(cleanedCoords, "result").String(), secretKey, captchaToken)
				log.Println(username, "::CheckCaptcha返回: ", e, "错误: ", err)

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
					// 识别验证出错，立即重新获取验证码
					continue
				}
				
				cleanedE := strings.ReplaceAll(strings.ReplaceAll(strings.ReplaceAll(e, " ", ""), "\n", ""), "\t", "")
				if cleanedE != "" {
					tradeNo, submitErr := request.SubmitReservation(user, &c, cleanedE, captchaToken)
					if submitErr != nil {
						log.Println(username, "::SubmitReservation: ", submitErr)
						CaptchaLogService.SaveStepLog(username, "SubmitReservation", "Error", fmt.Sprintf("预约场地 %v 失败: %v", c.CourtID, submitErr))
						break // 成功发送了submit请求，不管是否预定成功，退出循环
					}
					
					log.Println(username, "::预约成功")
					CaptchaLogService.SaveStepLog(username, "SubmitReservation", "Success", fmt.Sprintf("预约场地 %v 成功, 订单号: %s", c.CourtID, tradeNo))
					
					RedisService.SetUserOrder(username, tradeNo)
					
					if user.RelayEnabled {
						log.Println(username, "测试接力任务")
						CaptchaLogService.SaveStepLog(username, "RelayTask", "Info", "触发接力任务")
						RelayTask(user, &c, tradeNo, user.RelayUntil)
					}
					break // 成功发送了请求
				}
			}
		}(court)
	}
	courtWg.Wait()
	
	fmt.Println("测试结束")
	ctx.JSON(http.StatusOK, gin.H{"message": "测试运行结束"})
}
