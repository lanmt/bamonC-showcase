package model

type Court struct {
	ID          uint64  `gorm:"primaryKey;column:id"`
	UserID      uint64  `gorm:"not null;column:user_id"`
	VenueSiteID uint64  `gorm:"not null;column:venue_site_id"`
	CourtID     uint64  `gorm:"not null;column:court_id"`
	Time1ID     *uint64 `gorm:"column:time1_id"`
	Time2ID     *uint64 `gorm:"column:time2_id"`

	User User `gorm:"foreignKey:UserID;references:ID;constraint:OnDelete:CASCADE"`
}

// TableName sets the table name to bc.courts
func (Court) TableName() string {
	return "bc.courts"
}
