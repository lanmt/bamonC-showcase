package model

import "time"

type User struct {
	ID            uint64     `gorm:"primaryKey;column:id"`
	Username      string     `gorm:"unique;not null;column:username"`
	Password      string     `gorm:"column:password;not null"`
	Role          string     `gorm:"column:role;default:'user'"`
	Point         string     `gorm:"column:point"`
	Auth          string     `gorm:"column:auth"`
	AuthUpdatedAt *time.Time `gorm:"column:auth_updated_at"`
	RelayUntil    *time.Time `gorm:"column:relay_until"`
	RelayEnabled  bool       `gorm:"column:relay_enabled;default:false"`
	Enabled       bool       `gorm:"column:enabled;default:true"`
	UpdatedAt     *time.Time `gorm:"column:updated_at;autoUpdateTime"`
	SelectBuddyID *uint64    `gorm:"column:select_buddy_id"`

	Buddies []Buddy `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`
	Courts  []Court `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`
}

// TableName sets the table name to bc.users
func (User) TableName() string {
	return "bc.users"
}
