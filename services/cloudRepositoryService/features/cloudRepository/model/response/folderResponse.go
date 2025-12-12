package response

import "time"

// FolderResponseDTO represents a folder with file count
type FolderResponseDTO struct {
	ID             uint      `json:"id"`
	FolderName     string    `json:"folder_name"`
	ParentFolderID *uint     `json:"parent_folder_id,omitempty"`
	FileCount      int       `json:"file_count"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// FolderTreeResponseDTO represents a folder with sub-folders
type FolderTreeResponseDTO struct {
	ID             uint                    `json:"id"`
	FolderName     string                  `json:"folder_name"`
	ParentFolderID *uint                   `json:"parent_folder_id,omitempty"`
	FileCount      int                     `json:"file_count"`
	SubFolders     []FolderTreeResponseDTO `json:"sub_folders,omitempty"`
	CreatedAt      time.Time               `json:"created_at"`
	UpdatedAt      time.Time               `json:"updated_at"`
}

// MoveFilesResponseDTO for move files operation
type MoveFilesResponseDTO struct {
	MovedCount int    `json:"moved_count"`
	Message    string `json:"message"`
}
