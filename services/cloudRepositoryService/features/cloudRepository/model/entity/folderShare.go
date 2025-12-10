package entity

import "time"

// SharePermission represents the type of permission granted in a share
type SharePermission string

const (
	SharePermissionRead  SharePermission = "read"  // Can view and download
	SharePermissionWrite SharePermission = "write" // Can view, download, upload, and modify
)

// FolderShare represents a folder sharing relationship between users
type FolderShare struct {
	ID           uint            `gorm:"primaryKey" json:"id"`
	FolderID     uint            `gorm:"not null;index" json:"folder_id"`
	OwnerID      int32           `gorm:"not null;index" json:"owner_id"`
	SharedWithID int32           `gorm:"not null;index" json:"shared_with_id"`
	Permission   SharePermission `gorm:"size:10;default:'read';not null" json:"permission"`
	Folder       *Folder         `gorm:"foreignKey:FolderID" json:"folder,omitempty"`
	Owner        *User           `gorm:"foreignKey:OwnerID" json:"owner,omitempty"`
	SharedWith   *User           `gorm:"foreignKey:SharedWithID" json:"shared_with,omitempty"`
	CreatedAt    time.Time       `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt    time.Time       `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt    *time.Time      `gorm:"index" json:"deleted_at,omitempty"`
}

// TableName specifies the table name for FolderShare
func (FolderShare) TableName() string {
	return "folder_shares"
}
