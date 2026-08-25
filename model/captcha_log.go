package model

import "time"

// CaptchaLog 记录每次验证码识别的坐标及标注图片路径，现扩展为通用步骤日志
type CaptchaLog struct {
	ID         uint64     `gorm:"primaryKey;autoIncrement;column:id"`
	Username   string     `gorm:"not null;column:username;index"`
	StepName   string     `gorm:"column:step_name;index"` // 比如 GetCaptcha, SubmitReservation 等
	Status     string     `gorm:"column:status"`          // Success, Error
	Message    string     `gorm:"column:message"`         // 执行详情或错误信息
	ImagePath  string     `gorm:"column:image_path"`      // 相对于static目录的路径
	WordList   string     `gorm:"column:word_list"`       // 验证码要求的汉字列表
	CoordsJSON string     `gorm:"column:coords_json"`     // 原始坐标JSON字符串
	CreatedAt  *time.Time `gorm:"column:created_at;autoCreateTime;index"`
}

func (CaptchaLog) TableName() string {
	return "bc.captcha_logs"
}
