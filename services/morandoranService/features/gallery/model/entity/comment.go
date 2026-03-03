package entity

import (
	"time"

	"gorm.io/gorm"
)

type GalleryComment struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	GalleryID uint           `gorm:"not null" json:"galleryId"`
	AuthorID  uint           `gorm:"not null" json:"authorId"`
	Content   string         `gorm:"size:300;not null" json:"content"`
	CreatedAt time.Time      `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime" json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

func (GalleryComment) TableName() string {
	return "gallery_comments"
}
