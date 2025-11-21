package model

import "time"

// FileType represents the type of file (image or video)
type FileType string

const (
	FileTypeImage FileType = "image"
	FileTypeVideo FileType = "video"
)

// CloudFile represents a file stored in cloud storage
type CloudFile struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	UserID      uint      `gorm:"not null;index" json:"user_id"`
	FileName    string    `gorm:"size:255;not null" json:"file_name"`
	S3Key       string    `gorm:"size:512;not null;uniqueIndex" json:"s3_key"`
	FileType    FileType  `gorm:"size:20;not null;index" json:"file_type"`
	ContentType string    `gorm:"size:100;not null" json:"content_type"`
	FileSize    int64     `gorm:"not null" json:"file_size"`
	Tags        []Tag     `gorm:"many2many:file_tags;" json:"tags,omitempty"`
	CreatedAt   time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt   *time.Time `gorm:"index" json:"deleted_at,omitempty"`
}

// TableName specifies the table name for CloudFile
func (CloudFile) TableName() string {
	return "cloud_files"
}
