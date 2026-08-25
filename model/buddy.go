package model

type Buddy struct {
	ID        uint64 `gorm:"primaryKey;column:id"`
	UserID    uint64 `gorm:"not null;column:user_id"`
	BuddyID   uint64 `gorm:"not null;column:buddy_id"`
	BuddyName string `gorm:"column:buddy_name"`

	User User `gorm:"foreignKey:UserID;references:ID;constraint:OnDelete:CASCADE"`
}

// TableName sets the table name to bc.buddies
func (Buddy) TableName() string {
	return "bc.buddies"
}
