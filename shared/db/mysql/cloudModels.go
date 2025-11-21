package mysql

import (
	"time"

	"gorm.io/datatypes"
)

// User extension (partial struct for GORM)
type User struct {
	ID           uint   `gorm:"primaryKey"`
	StorageUsed  int64  `gorm:"default:0"`
	StorageLimit int64  `gorm:"default:16106127360"` // 15GB
}

// File represents the files table
type File struct {
	ID           string         `gorm:"type:varchar(36);primaryKey"`
	UserID       uint           `gorm:"not null;index"`
	Name         string         `gorm:"size:255"`
	OriginalName string         `gorm:"size:255"`
	S3Key        string         `gorm:"size:512"`
	URL          string         `gorm:"size:1024"`
	MimeType     string         `gorm:"size:100"`
	Size         int64          `gorm:"not null"`
	CreatedAt    time.Time      `gorm:"autoCreateTime"`
	UpdatedAt    time.Time      `gorm:"autoUpdateTime"`
	IsDeleted    bool           `gorm:"default:false"`
	Metadata     datatypes.JSON `gorm:"type:json"`
	Tags         []Tag          `gorm:"many2many:file_tags;"`
}

// Tag represents the tags table
type Tag struct {
	ID        string    `gorm:"type:varchar(36);primaryKey"`
	Name      string    `gorm:"size:50;uniqueIndex"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
	Files     []File    `gorm:"many2many:file_tags;"`
}

// ActivityLog represents the activity_logs table
type ActivityLog struct {
	ID         string         `gorm:"type:varchar(36);primaryKey"`
	UserID     uint           `gorm:"not null;index"`
	ActionType string         `gorm:"size:50"`
	TargetID   string         `gorm:"size:36"`
	Metadata   datatypes.JSON `gorm:"type:json"`
	CreatedAt  time.Time      `gorm:"autoCreateTime"`
}

// TableName overrides
func (File) TableName() string {
	return "files"
}

func (Tag) TableName() string {
	return "tags"
}

func (ActivityLog) TableName() string {
	return "activity_logs"
}
