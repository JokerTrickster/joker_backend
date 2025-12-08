package request

// CreateFolderRequestDTO for creating a new folder
type CreateFolderRequestDTO struct {
	FolderName     string `json:"folder_name" validate:"required,min=1,max=255"`
	ParentFolderID *uint  `json:"parent_folder_id"`
}

// UpdateFolderRequestDTO for updating folder name or parent
type UpdateFolderRequestDTO struct {
	FolderName     *string `json:"folder_name,omitempty" validate:"omitempty,min=1,max=255"`
	ParentFolderID *uint   `json:"parent_folder_id,omitempty"`
}

// MoveFilesToFolderRequestDTO for moving multiple files to a folder
type MoveFilesToFolderRequestDTO struct {
	FileIDs        []uint `json:"file_ids" validate:"required,min=1,dive,min=1"`
	TargetFolderID *uint  `json:"target_folder_id"` // null means move to root
}
