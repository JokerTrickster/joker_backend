package entity

import "time"

type GalleryLike struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	UserID    uint      `gorm:"not null;uniqueIndex:idx_gallery_likes_user_gallery" json:"userId"`
	GalleryID uint      `gorm:"not null;uniqueIndex:idx_gallery_likes_user_gallery" json:"galleryId"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"createdAt"`
}

func (GalleryLike) TableName() string {
	return "gallery_likes"
}
