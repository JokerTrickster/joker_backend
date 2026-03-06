package entity

import (
	"time"

	"gorm.io/gorm"
)

type MorandoranUser struct {
	ID           uint           `gorm:"primaryKey" json:"id"`
	Nickname     string         `gorm:"size:50;not null" json:"nickname"`
	Email        string         `gorm:"size:255;not null;uniqueIndex" json:"email"`
	Password     *string        `gorm:"size:255" json:"-"`
	Role         string         `gorm:"size:10;not null;default:user" json:"role"`
	Provider     string         `gorm:"size:20;not null;default:email" json:"provider"`
	ProfileImage *string        `gorm:"size:512;column:profile_image" json:"profileImage"`
	CreatedAt    time.Time      `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt    time.Time      `gorm:"autoUpdateTime" json:"updatedAt"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName returns "morandoran_users" (legacy DB table name, intentionally kept to avoid migration)
func (MorandoranUser) TableName() string {
	return "morandoran_users"
}
