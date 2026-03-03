package entity

import (
	"time"

	"gorm.io/gorm"
)

type News struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	Title     string         `gorm:"size:200;not null" json:"title"`
	Summary   string         `gorm:"size:500" json:"summary"`
	Content   string         `gorm:"type:text;not null" json:"content"`
	Thumbnail string         `gorm:"size:512" json:"thumbnail"`
	Category  string         `gorm:"size:50;not null" json:"category"`
	Date      string         `gorm:"type:date;not null" json:"date"`
	CreatedAt time.Time      `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime" json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

func (News) TableName() string {
	return "news"
}
