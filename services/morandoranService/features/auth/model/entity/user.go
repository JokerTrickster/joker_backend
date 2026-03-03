package entity

import (
	"time"

	"gorm.io/gorm"
)

type MorandoranUser struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	Nickname  string         `gorm:"size:50;not null" json:"nickname"`
	Email     string         `gorm:"size:255;not null;uniqueIndex" json:"email"`
	Password  string         `gorm:"size:255;not null" json:"-"`
	Role      string         `gorm:"size:10;not null;default:user" json:"role"`
	CreatedAt time.Time      `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime" json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

func (MorandoranUser) TableName() string {
	return "morandoran_users"
}
