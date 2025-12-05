package entity

import "time"

// MultipartUploadStatus represents the status of multipart upload
type MultipartUploadStatus string

const (
	MultipartUploadStatusInitiated MultipartUploadStatus = "initiated"
	MultipartUploadStatusUploading MultipartUploadStatus = "uploading"
	MultipartUploadStatusCompleted MultipartUploadStatus = "completed"
	MultipartUploadStatusAborted   MultipartUploadStatus = "aborted"
)

// MultipartUpload represents a multipart upload session
type MultipartUpload struct {
	ID          uint                  `gorm:"primaryKey" json:"id"`
	UserID      uint                  `gorm:"not null;index" json:"user_id"`
	UploadID    string                `gorm:"size:255;not null;index" json:"upload_id"`
	FileKey     string                `gorm:"size:500;not null" json:"file_key"`
	FileName    string                `gorm:"size:255;not null" json:"file_name"`
	FileSize    int64                 `gorm:"not null" json:"file_size"`
	ContentType string                `gorm:"size:100" json:"content_type"`
	PartSize    int                   `gorm:"not null" json:"part_size"`
	TotalParts  int                   `gorm:"not null" json:"total_parts"`
	Status      MultipartUploadStatus `gorm:"size:20;default:'initiated';index" json:"status"`
	CreatedAt   time.Time             `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time             `gorm:"autoUpdateTime" json:"updated_at"`
}

// TableName specifies the table name for MultipartUpload
func (MultipartUpload) TableName() string {
	return "multipart_uploads"
}
