package entity

import (
	"time"

	"gorm.io/gorm"
)

type GalleryPost struct {
	ID           uint           `gorm:"primaryKey" json:"id"`
	AuthorID     uint           `gorm:"not null" json:"authorId"`
	MediaType    string         `gorm:"size:10;not null" json:"mediaType"`
	MediaURL     string         `gorm:"size:512;not null;column:media_url" json:"mediaUrl"`
	ThumbnailURL string         `gorm:"size:512;not null;column:thumbnail_url" json:"thumbnailUrl"`
	Caption      string         `gorm:"size:500" json:"caption"`
	LikeCount    int            `gorm:"not null;default:0" json:"likeCount"`
	CommentCount int            `gorm:"not null;default:0" json:"commentCount"`
	CreatedAt    time.Time      `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt    time.Time      `gorm:"autoUpdateTime" json:"updatedAt"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
}

func (GalleryPost) TableName() string {
	return "gallery_posts"
}
