package entity

import "time"

// Tag represents a user-defined tag for files
type Tag struct {
	ID        uint        `gorm:"primaryKey" json:"id"`
	UserID    uint        `gorm:"not null;index" json:"user_id"`
	Name      string      `gorm:"size:50;not null" json:"name"`
	Files     []CloudFile `gorm:"many2many:file_tags;" json:"files,omitempty"`
	CreatedAt time.Time   `gorm:"autoCreateTime" json:"created_at"`
}

// TableName specifies the table name for Tag
func (Tag) TableName() string {
	return "tags"
}
