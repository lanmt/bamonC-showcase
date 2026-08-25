package service

import (
	"bamonC/model"
	"errors"

	"gorm.io/gorm"
)

type CourtService struct {
	DB *gorm.DB
}

// 根据用户名获取该用户的所有 Court
func (s *CourtService) GetCourts(username string) ([]model.Court, error) {
	var user model.User
	var courts []model.Court

	// 查找用户
	if err := s.DB.Where("username = ?", username).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}

	// 查找该用户的所有 Court（带上 User 信息）
	if err := s.DB.Where("user_id = ?", user.ID).
		Find(&courts).Error; err != nil {
		return nil, err
	}

	return courts, nil
}

// 添加一条 Court 记录
func (s *CourtService) AddCourt(court *model.Court) error {
	return s.DB.Create(court).Error
}

// 根据 ID 删除一条 Court
func (s *CourtService) DeleteCourt(id uint64) error {
	return s.DB.Delete(&model.Court{}, id).Error
}

func (s *CourtService) Add15ToAllTimeIDs() error {
	err := s.DB.Exec(`
		UPDATE bc.courts
		SET time1_id = CASE 
				WHEN time1_id + 15 > 8792 THEN 8688 + ((time1_id + 15) - 8792 - 1)
				ELSE time1_id + 15
			END,
		    time2_id = CASE 
				WHEN time2_id IS NOT NULL THEN
					CASE 
						WHEN time2_id + 15 > 8792 THEN 8688 + ((time2_id + 15) - 8792 - 1)
						ELSE time2_id + 15
					END
				ELSE NULL
			END
		WHERE venue_site_id = 38
	`).Error
	if err != nil {
		return err
	}

	// venue_site_id = 39
	err = s.DB.Exec(`
		UPDATE bc.courts
		SET time1_id = CASE 
				WHEN time1_id + 15 > 9172 THEN 9068 + ((time1_id + 15) - 9172 - 1)
				ELSE time1_id + 15
			END,
		    time2_id = CASE 
				WHEN time2_id IS NOT NULL THEN
					CASE 
						WHEN time2_id + 15 > 9172 THEN 9068 + ((time2_id + 15) - 9172 - 1)
						ELSE time2_id + 15
					END
				ELSE NULL
			END
		WHERE venue_site_id = 39
	`).Error
	if err != nil {
		return err
	}
	return nil
}
