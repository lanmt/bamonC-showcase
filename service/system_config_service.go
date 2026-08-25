package service

import (
	"bamonC/model"

	"gorm.io/gorm"
)

type SystemConfigService struct {
	DB *gorm.DB
}

// GetConfig 获取配置项，如果不存在则返回 defaultValue，并写入数据库
func (s *SystemConfigService) GetConfig(key, defaultValue string) (string, error) {
	var conf model.SystemConfig
	err := s.DB.Where("config_key = ?", key).First(&conf).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			// 如果不存在，写入默认值
			conf = model.SystemConfig{
				Key:   key,
				Value: defaultValue,
			}
			if createErr := s.DB.Create(&conf).Error; createErr != nil {
				return defaultValue, createErr
			}
			return defaultValue, nil
		}
		return defaultValue, err
	}
	return conf.Value, nil
}

// SetConfig 设置配置项
func (s *SystemConfigService) SetConfig(key, value string) error {
	var conf model.SystemConfig
	err := s.DB.Where("config_key = ?", key).First(&conf).Error
	if err == gorm.ErrRecordNotFound {
		conf = model.SystemConfig{Key: key, Value: value}
		return s.DB.Create(&conf).Error
	} else if err != nil {
		return err
	}

	conf.Value = value
	return s.DB.Save(&conf).Error
}

// GetAllConfigs 获取所有配置
func (s *SystemConfigService) GetAllConfigs() (map[string]string, error) {
	var configs []model.SystemConfig
	if err := s.DB.Find(&configs).Error; err != nil {
		return nil, err
	}
	result := make(map[string]string)
	for _, c := range configs {
		result[c.Key] = c.Value
	}
	return result, nil
}
