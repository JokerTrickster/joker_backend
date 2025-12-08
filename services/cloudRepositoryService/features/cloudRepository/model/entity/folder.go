package entity

import "time"

// Folder represents a user-defined folder for organizing files
type Folder struct {
	ID             uint        `gorm:"primaryKey" json:"id"`
	UserID         uint        `gorm:"not null;index" json:"user_id"`
	FolderName     string      `gorm:"size:255;not null" json:"folder_name"`
	ParentFolderID *uint       `gorm:"index" json:"parent_folder_id,omitempty"`
	Files          []CloudFile `gorm:"foreignKey:FolderID" json:"files,omitempty"`
	SubFolders     []Folder    `gorm:"foreignKey:ParentFolderID" json:"sub_folders,omitempty"`
	CreatedAt      time.Time   `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt      time.Time   `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt      *time.Time  `gorm:"index" json:"deleted_at,omitempty"`
}

// TableName specifies the table name for Folder
func (Folder) TableName() string {
	return "folders"
}
