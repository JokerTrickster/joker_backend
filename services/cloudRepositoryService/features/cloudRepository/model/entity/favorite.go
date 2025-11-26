package entity

import "time"

// Favorite represents a user's favorited file
type Favorite struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	UserID      uint      `gorm:"not null;index:idx_user_favorited_at" json:"user_id"`
	FileID      uint      `gorm:"not null;uniqueIndex:uniq_user_file" json:"file_id"`
	FavoritedAt time.Time `gorm:"autoCreateTime" json:"favorited_at"`
}

// TableName specifies the table name for Favorite
func (Favorite) TableName() string {
	return "favorites"
}
