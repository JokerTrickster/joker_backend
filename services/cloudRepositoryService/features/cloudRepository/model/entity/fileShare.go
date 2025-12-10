package entity

import "time"

// FileShare represents a file sharing relationship between users
type FileShare struct {
	ID           uint            `gorm:"primaryKey" json:"id"`
	FileID       uint            `gorm:"not null;index" json:"file_id"`
	OwnerID      int32           `gorm:"not null;index" json:"owner_id"`
	SharedWithID int32           `gorm:"not null;index" json:"shared_with_id"`
	Permission   SharePermission `gorm:"size:10;default:'read';not null" json:"permission"`
	File         *CloudFile      `gorm:"foreignKey:FileID" json:"file,omitempty"`
	Owner        *User           `gorm:"foreignKey:OwnerID" json:"owner,omitempty"`
	SharedWith   *User           `gorm:"foreignKey:SharedWithID" json:"shared_with,omitempty"`
	CreatedAt    time.Time       `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt    time.Time       `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt    *time.Time      `gorm:"index" json:"deleted_at,omitempty"`
}

// TableName specifies the table name for FileShare
func (FileShare) TableName() string {
	return "file_shares"
}
