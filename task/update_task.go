package task

import (
	"log"
)

func UpdateTimeIdTask() {
	log.Println("每日更新时间")
	err := CourtService.Add15ToAllTimeIDs()
	if err != nil {
		log.Println("更新时间错误: ", err)
	}
}
