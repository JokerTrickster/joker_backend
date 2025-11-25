package entity

import "time"

// ActivityType represents the type of activity
type ActivityType string

const (
	ActivityTypeUpload   ActivityType = "upload"
	ActivityTypeDownload ActivityType = "download"
	ActivityTypeTagAdd   ActivityType = "tag_add"
	ActivityTypeTagDel   ActivityType = "tag_del"
)

// ActivityLog represents user activity logs
type ActivityLog struct {
	ID           uint         `gorm:"primaryKey" json:"id"`
	UserID       uint         `gorm:"not null;index:idx_user_activity" json:"user_id"`
	FileID       *uint        `gorm:"index" json:"file_id,omitempty"`
	ActivityType ActivityType `gorm:"size:20;not null;index:idx_user_activity" json:"activity_type"`
	TagName      string       `gorm:"size:100" json:"tag_name,omitempty"`
	CreatedAt    time.Time    `gorm:"autoCreateTime;index:idx_user_activity" json:"created_at"`
}

// TableName specifies the table name for ActivityLog
func (ActivityLog) TableName() string {
	return "activity_logs"
}