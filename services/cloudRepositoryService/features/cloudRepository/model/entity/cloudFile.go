package entity

import "time"

// FileType represents the type of file (image or video)
type FileType string

const (
	FileTypeImage FileType = "image"
	FileTypeVideo FileType = "video"
)

// ProcessingStatus represents the status of video processing
type ProcessingStatus string

const (
	ProcessingStatusPending    ProcessingStatus = "pending"
	ProcessingStatusProcessing ProcessingStatus = "processing"
	ProcessingStatusCompleted  ProcessingStatus = "completed"
	ProcessingStatusFailed     ProcessingStatus = "failed"
)

// ProcessingStage represents the current stage of processing
type ProcessingStage string

const (
	ProcessingStageNone               ProcessingStage = ""
	ProcessingStageValidating         ProcessingStage = "validating_file"
	ProcessingStageExtractingMetadata ProcessingStage = "extracting_metadata"
	ProcessingStageGeneratingThumbnail ProcessingStage = "generating_thumbnail"
	ProcessingStageUploadingThumbnail ProcessingStage = "uploading_thumbnail"
	ProcessingStageFinalizing         ProcessingStage = "finalizing"
	ProcessingStageDone               ProcessingStage = "done"
)

// CloudFile represents a file stored in cloud storage
type CloudFile struct {
	ID                    uint             `gorm:"primaryKey" json:"id"`
	UserID                uint             `gorm:"not null;index" json:"user_id"`
	FolderID              *uint            `gorm:"index" json:"folder_id,omitempty"`
	FileName              string           `gorm:"size:255;not null" json:"file_name"`
	S3Key                 string           `gorm:"size:512;not null;uniqueIndex" json:"s3_key"`
	ThumbnailKey          string           `gorm:"size:512" json:"thumbnail_key,omitempty"`
	FileType              FileType         `gorm:"size:20;not null;index" json:"file_type"`
	ContentType           string           `gorm:"size:100;not null" json:"content_type"`
	FileSize              int64            `gorm:"not null" json:"file_size"`
	Duration              *float64         `gorm:"type:decimal(10,2)" json:"duration,omitempty"` // Video duration in seconds
	ProcessingStatus      ProcessingStatus `gorm:"size:20;default:'pending';index" json:"processing_status"`
	ProcessingProgress    int              `gorm:"default:0" json:"processing_progress"`         // 0-100
	ProcessingStage       ProcessingStage  `gorm:"size:50" json:"processing_stage,omitempty"`
	ProcessingError       string           `gorm:"size:512" json:"processing_error,omitempty"`
	ProcessingStartedAt   *time.Time       `json:"processing_started_at,omitempty"`
	ProcessingCompletedAt *time.Time       `json:"processing_completed_at,omitempty"`
	Tags                  []Tag            `gorm:"many2many:file_tags;" json:"tags,omitempty"`
	CreatedAt             time.Time        `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt             time.Time        `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt             *time.Time       `gorm:"index" json:"deleted_at,omitempty"`
}

// TableName specifies the table name for CloudFile
func (CloudFile) TableName() string {
	return "cloud_files"
}
