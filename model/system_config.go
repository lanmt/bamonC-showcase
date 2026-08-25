package model

// SystemConfig 系统动态配置表
type SystemConfig struct {
	Key   string `gorm:"primaryKey;column:config_key"`
	Value string `gorm:"column:config_value"`
}

func (SystemConfig) TableName() string {
	return "bc.system_configs"
}
