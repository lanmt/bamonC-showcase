package service

import (
	"bamonC/model"
	"errors"

	"gorm.io/gorm"
)

type BuddyService struct {
	DB *gorm.DB
}

// 根据 username 获取 buddies
func (s *BuddyService) GetBuddies(username string) ([]model.Buddy, error) {
	var user model.User
	var buddies []model.Buddy

	// 查找用户
	if err := s.DB.Where("username = ?", username).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}

	// 查找该用户的所有 Buddy（带上 User 信息）
	if err := s.DB.Where("user_id = ?", user.ID).
		Find(&buddies).Error; err != nil {
		return nil, err
	}

	return buddies, nil
}

// 根据 username 删除该用户的所有 buddies
func (s *BuddyService) DeleteBuddiesByUsername(username string) error {
	var user model.User

	// 查找用户
	if err := s.DB.Where("username = ?", username).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("user not found")
		}
		return err
	}

	// 删除该用户的所有 Buddy
	if err := s.DB.Where("user_id = ?", user.ID).Delete(&model.Buddy{}).Error; err != nil {
		return err
	}

	return nil
}

// 添加一条 buddy 记录
func (s *BuddyService) AddBuddy(buddy *model.Buddy) error {
	return s.DB.Create(buddy).Error
}

// 根据 ID 删除一条 buddy
func (s *BuddyService) DeleteBuddy(id uint64) error {
	return s.DB.Delete(&model.Buddy{}, id).Error
}
