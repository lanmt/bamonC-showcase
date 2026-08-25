package service

import (
	"bamonC/model"
	"errors"
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type UserService struct {
	DB *gorm.DB
}

// 根据 username 获取用户
func (s *UserService) GetByUsername(username string) (*model.User, error) {
	var user model.User
	if err := s.DB.Where("username = ?", username).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

// 根据 id 获取用户
func (s *UserService) GetById(id uint) (*model.User, error) {
	var user model.User
	if err := s.DB.Where("id = ?", id).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

// 创建用户
func (s *UserService) CreateUser(username, password string) error {
	existing, _ := s.GetByUsername(username)
	if existing != nil {
		return errors.New("创建用户： username already exists")
	}
	
	hashed, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	
	user := &model.User{
		Username:      username,
		Password:      string(hashed),
		Role:          "user",
		Point:         "point-ebd81b4b-077b-4b38-9923-193acec3e13f",
		Auth:          "",
		AuthUpdatedAt: func() *time.Time { v := time.Now(); return &v }(),
		RelayEnabled:  false,
		Enabled:       false,
		UpdatedAt:     func() *time.Time { v := time.Now(); return &v }(),
	}
	
	return s.DB.Create(user).Error
}

// 修改用户
func (s *UserService) UpdateUser(username string, input *model.User) error {
	existing, _ := s.GetByUsername(username)
	if existing == nil {
		return errors.New("修改用户： user not exists")
	}

	// Password update
	if input.Password != "" {
		hashed, _ := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
		existing.Password = string(hashed)
	}

	existing.Point = input.Point
	existing.RelayUntil = input.RelayUntil
	existing.RelayEnabled = input.RelayEnabled
	existing.Enabled = input.Enabled
	existing.SelectBuddyID = input.SelectBuddyID

	if input.Auth != "" {
		existing.Auth = input.Auth
		existing.AuthUpdatedAt = func() *time.Time { v := time.Now(); return &v }()
	}

	existing.UpdatedAt = func() *time.Time { v := time.Now(); return &v }()
	
	return s.DB.Save(existing).Error
}

// 删除用户（settings 和 courts 会自动级联删除）
func (s *UserService) DeleteUser(username string) error {
	user, err := s.GetByUsername(username)
	if err != nil {
		return err
	}
	if user == nil {
		return errors.New("user not found")
	}

	return s.DB.Delete(user).Error
}

// 获取所有 enabled = true 的用户
func (s *UserService) GetAllEnabled() ([]model.User, error) {
	var userList []model.User
	if err := s.DB.Where("enabled = ?", true).Find(&userList).Error; err != nil {
		return nil, err
	}
	return userList, nil
}

// 获取所有的 Setting
