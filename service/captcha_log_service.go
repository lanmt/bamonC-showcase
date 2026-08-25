package service

import (
	"bamonC/model"
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	_ "image/jpeg"
	"image/png"
	"math"
	"os"
	"path/filepath"
	"time"

	"gorm.io/gorm"
)

type CaptchaLogService struct {
	DB *gorm.DB
}

// Coord 表示一个识别出的坐标点
type Coord struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

// SaveCaptchaLog 在原始验证码图片上画红点并存储最终校验结果 (合并GetCoords与CheckCaptcha阶段)
func (s *CaptchaLogService) SaveCaptchaLog(username, base64ImageStr, wordList, coordsJSON string, checkErr error) error {
	// 解析坐标
	type CoordsResult struct {
		Result []Coord `json:"result"`
	}
	var cr CoordsResult
	if err := json.Unmarshal([]byte(coordsJSON), &cr); err != nil {
		return fmt.Errorf("parse coordsJSON: %w", err)
	}

	// 解码图片
	imgBytes, err := base64.StdEncoding.DecodeString(base64ImageStr)
	if err != nil {
		return fmt.Errorf("decode base64 image: %w", err)
	}

	srcImg, _, err := image.Decode(bytes.NewReader(imgBytes))
	if err != nil {
		return fmt.Errorf("decode image: %w", err)
	}

	// 在图片上画红点
	bounds := srcImg.Bounds()
	dst := image.NewRGBA(bounds)
	draw.Draw(dst, bounds, srcImg, bounds.Min, draw.Src)

	red := color.RGBA{R: 255, G: 0, B: 0, A: 255}
	dotRadius := 6
	for i, c := range cr.Result {
		// 这里只画点
		drawRedDot(dst, int(math.Round(c.X)), int(math.Round(c.Y)), dotRadius, red)
		_ = i // 如果之后要写数字可以利用 i
	}

	// 确保目录存在
	logDir := filepath.Join("static", "captcha_logs", username)
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return fmt.Errorf("mkdirall: %w", err)
	}

	// 生成唯一文件名
	filename := fmt.Sprintf("%d.png", time.Now().UnixNano())
	fullPath := filepath.Join(logDir, filename)
	f, err := os.Create(fullPath)
	if err != nil {
		return fmt.Errorf("create image file: %w", err)
	}
	defer f.Close()

	if err := png.Encode(f, dst); err != nil {
		return fmt.Errorf("encode png: %w", err)
	}

	status := "Success"
	message := "验证成功"
	if checkErr != nil {
		status = "Error"
		message = checkErr.Error()
	}

	// 存入数据库（相对路径，方便前端访问 /static/captcha_logs/...）
	relativePath := filepath.ToSlash(filepath.Join("captcha_logs", username, filename))
	entry := &model.CaptchaLog{
		Username:   username,
		StepName:   "CheckCaptcha", // merged layer
		Status:     status,
		Message:    message,
		ImagePath:  relativePath,
		WordList:   wordList,
		CoordsJSON: coordsJSON,
	}
	return s.DB.Create(entry).Error
}

// SaveStepLog 记录通用步骤日志
func (s *CaptchaLogService) SaveStepLog(username, stepName, status, message string) {
	entry := &model.CaptchaLog{
		Username: username,
		StepName: stepName,
		Status:   status,
		Message:  message,
	}
	// 不阻塞，直接使用 goroutine 或者同步写都可以。为了稳定我们同步写，失败打log即可。
	if err := s.DB.Create(entry).Error; err != nil {
		fmt.Printf("SaveStepLog Error: %v\n", err)
	}
}

// SaveRawCaptchaImage 保存 GetCaptcha 阶段获取的未经处理的原始验证码图片
func (s *CaptchaLogService) SaveRawCaptchaImage(username, base64ImageStr, wordList string) error {
	imgBytes, err := base64.StdEncoding.DecodeString(base64ImageStr)
	if err != nil {
		return fmt.Errorf("decode base64 image: %w", err)
	}

	srcImg, _, err := image.Decode(bytes.NewReader(imgBytes))
	if err != nil {
		return fmt.Errorf("decode image: %w", err)
	}

	logDir := filepath.Join("static", "captcha_logs", username)
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return fmt.Errorf("mkdirall: %w", err)
	}

	filename := fmt.Sprintf("%d_raw.png", time.Now().UnixNano())
	fullPath := filepath.Join(logDir, filename)
	f, err := os.Create(fullPath)
	if err != nil {
		return fmt.Errorf("create image file: %w", err)
	}
	defer f.Close()

	if err := png.Encode(f, srcImg); err != nil {
		return fmt.Errorf("encode png: %w", err)
	}

	relativePath := filepath.ToSlash(filepath.Join("captcha_logs", username, filename))
	entry := &model.CaptchaLog{
		Username:   username,
		StepName:   "GetCaptcha",
		Status:     "Success",
		Message:    "成功获取验证码原始图片",
		ImagePath:  relativePath,
		WordList:   wordList,
	}
	return s.DB.Create(entry).Error
}

// GetFilteredLogs 分页且过滤查询
func (s *CaptchaLogService) GetFilteredLogs(username, dateStr, stepName string, page, pageSize int) ([]model.CaptchaLog, int64, error) {
	var logs []model.CaptchaLog
	var total int64

	query := s.DB.Model(&model.CaptchaLog{})
	if username != "" {
		query = query.Where("username = ?", username)
	}
	if stepName != "" {
		query = query.Where("step_name LIKE ?", "%"+stepName+"%")
	}
	if dateStr != "" {
		// PostgreSQL 支持 DATE(created_at)
		query = query.Where("CAST(created_at AS DATE) = ?", dateStr)
	}

	query.Count(&total)
	err := query.Order("created_at DESC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&logs).Error
	return logs, total, err
}

// drawRedDot 在图片上指定坐标处画一个实心圆点
func drawRedDot(img *image.RGBA, cx, cy, radius int, c color.RGBA) {
	bounds := img.Bounds()
	for y := cy - radius; y <= cy+radius; y++ {
		for x := cx - radius; x <= cx+radius; x++ {
			if x < bounds.Min.X || x >= bounds.Max.X || y < bounds.Min.Y || y >= bounds.Max.Y {
				continue
			}
			dx := x - cx
			dy := y - cy
			if dx*dx+dy*dy <= radius*radius {
				img.SetRGBA(x, y, c)
			}
		}
	}
}
