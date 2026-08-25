package task

import (
	"bamonC/model"
	"bamonC/request"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/tidwall/gjson"
)

// RelayTask 每 8 分钟执行一次，直到 relayUntil 结束
func RelayTask(user *model.User, court *model.Court, tradeNo string, relayUntil *time.Time) {

	relayTime := time.Date(
		time.Now().Year(),
		time.Now().Month(),
		time.Now().Day(),
		relayUntil.Hour(),
		relayUntil.Minute(),
		0, 0,
		time.Now().Location(),
	)

	ticker := time.NewTicker(8 * time.Minute)

	go func() {
		defer ticker.Stop()

		currentTradeNo := tradeNo // 每次执行会更新这个变量
		<-ticker.C
		for {
			// 若到达 relayUntil，停止任务
			if time.Now().After(relayTime) {
				log.Println("now time: ", time.Now())
				log.Println("realyTime: ", relayTime)
				log.Println(user.Username, "到达 relayUntil，停止 RelayTask")
				return
			}
			log.Println(user.Username, "执行RelayTask")
			newTradeNo, err := doRelay(user, court, currentTradeNo)
			if err != nil {
				log.Println(user.Username, "执行 relay 错误:", err)
				CaptchaLogService.SaveStepLog(user.Username, "RelayTask", "Error", fmt.Sprintf("执行接力流程错误: %v", err))
				break
			} else {
				currentTradeNo = newTradeNo
			}

			<-ticker.C
		}
	}()
}

// 示例：你自己的 relay 逻辑
func doRelay(user *model.User, court *model.Court, tradeNo string) (string, error) {
	log.Println(user.Username, "执行 relay, 使用 tradeNo:", tradeNo)
	e := ""
	token := ""
	for e == "" {
		captchaRes, err := request.GetCaptcha(user, user.Username)
		if captchaRes == "" || err != nil {
			log.Println(err)
			CaptchaLogService.SaveStepLog(user.Username, "Relay_GetCaptcha", "Error", fmt.Sprintf("GetCaptcha频繁或错误: %v", err))
			// 这里可能是请求频繁
			time.Sleep(10 * time.Second)
			continue
		}
		coords, err := request.GetCoords(user, captchaRes)
		if coords == "" || err != nil {
			log.Println(user.Username, "::GetCoords错误: ", err)
			CaptchaLogService.SaveStepLog(user.Username, "Relay_GetCoords", "Error", fmt.Sprintf("获取坐标失败: %v", err))
			// 这里是坐标计算失败
			time.Sleep(3 * time.Second)
			continue
		}
		//删除corrds中所有的回车和空格
		cleanedCoords := strings.ReplaceAll(strings.ReplaceAll(strings.ReplaceAll(coords, " ", ""), "\n", ""), "\t", "")

		secretKey := gjson.Get(captchaRes, "repData.secretKey").String()
		// submit需要captchaToken
		captchaToken := gjson.Get(captchaRes, "repData.token").String()
		// submit需要e

		e, err = request.CheckCaptcha(user, gjson.Get(cleanedCoords, "result").String(), secretKey, captchaToken)

		// ---- 保存验证码坐标Log并附加校验结果（异步，不阻塞主流程）----
		go func(baseImg, cleaned string, cerr error) {
			imgBase64 := gjson.Get(baseImg, "repData.originalImageBase64").String()
			wordList := gjson.Get(baseImg, "repData.wordList").String()
			if imgBase64 != "" && cleaned != "" {
				if logErr := CaptchaLogService.SaveCaptchaLog(user.Username, imgBase64, wordList, cleaned, cerr); logErr != nil {
					log.Println(user.Username, "::SaveCaptchaLog:", logErr)
				}
			}
		}(captchaRes, cleanedCoords, err)

		if e == "" || err != nil {
			log.Println(user.Username, "::GetCoords: ", err)
			time.Sleep(10 * time.Second)
			e = ""
			continue
		}
		//去除e中的回车和空格
		cleanedE := strings.ReplaceAll(strings.ReplaceAll(strings.ReplaceAll(e, " ", ""), "\n", ""), "\t", "")
		//把验证通过的e和token存到redis
		if cleanedE != "" {
			e = cleanedE
			token = captchaToken
			break
		}
		//等待两秒获取验证码
		time.Sleep(3 * time.Second)
	}
	err := request.CancelOrder(user, tradeNo)
	if err != nil {
		log.Println(user.Username, "取消订单失败，订单号", tradeNo)
		CaptchaLogService.SaveStepLog(user.Username, "Relay_Cancel", "Error", fmt.Sprintf("取消旧订单[%s]失败: %v", tradeNo, err))
		return "", err
	}
	CaptchaLogService.SaveStepLog(user.Username, "Relay_Cancel", "Success", fmt.Sprintf("成功取消旧订单: %s", tradeNo))
	time.Sleep(1 * time.Second)
	newTradeNo, err := request.SubmitReservation(user, court, e, token)
	if err != nil {
		log.Println(user.Username, "被截胡了，惨")
		CaptchaLogService.SaveStepLog(user.Username, "Relay_Submit", "Error", fmt.Sprintf("接力抢新重新预约失败（可能被截胡）: %v", err))
		return newTradeNo, err
	}
	log.Println(user.Username, "接力成功，订单号: ", newTradeNo)
	CaptchaLogService.SaveStepLog(user.Username, "Relay_Submit", "Success", fmt.Sprintf("接力抢新成功，新订单号: %s", newTradeNo))
	// 假设每次后端返回新的 tradeNo
	//newTradeNo := tradeNo + "_next"
	return newTradeNo, nil
}
