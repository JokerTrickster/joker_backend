package entity

import "time"

// FolderShare represents a folder sharing relationship between users
type FolderShare struct {
	ID           uint       `gorm:"primaryKey" json:"id"`
	FolderID     uint       `gorm:"not null;index" json:"folder_id"`
	OwnerID      uint       `gorm:"not null;index" json:"owner_id"`
	SharedWithID uint       `gorm:"not null;index" json:"shared_with_id"`
	Folder       *Folder    `gorm:"foreignKey:FolderID" json:"folder,omitempty"`
	Owner        *User      `gorm:"foreignKey:OwnerID" json:"owner,omitempty"`
	SharedWith   *User      `gorm:"foreignKey:SharedWithID" json:"shared_with,omitempty"`
	CreatedAt    time.Time  `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt    time.Time  `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt    *time.Time `gorm:"index" json:"deleted_at,omitempty"`
}

// TableName specifies the table name for FolderShare
func (FolderShare) TableName() string {
	return "folder_shares"
}
