package response

import "time"

// UserInfoDTO represents basic user information
type UserInfoDTO struct {
	ID    int32  `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

// ShareUserDTO represents a user who has access to a shared resource
type ShareUserDTO struct {
	ID       int32     `json:"id"`
	Name     string    `json:"name"`
	Email    string    `json:"email"`
	SharedAt time.Time `json:"shared_at"`
}

// ShareFolderResponseDTO for folder share operation response
type ShareFolderResponseDTO struct {
	Message     string         `json:"message"`
	SharedUsers []ShareUserDTO `json:"shared_users"`
}

// ShareFileResponseDTO for file share operation response
type ShareFileResponseDTO struct {
	Message     string         `json:"message"`
	SharedUsers []ShareUserDTO `json:"shared_users"`
}

// FolderShareListResponseDTO for listing users who have access to a folder
type FolderShareListResponseDTO struct {
	FolderID    uint           `json:"folder_id"`
	FolderName  string         `json:"folder_name"`
	SharedUsers []ShareUserDTO `json:"shared_users"`
}

// FileShareListResponseDTO for listing users who have access to a file
type FileShareListResponseDTO struct {
	FileID      uint           `json:"file_id"`
	FileName    string         `json:"file_name"`
	SharedUsers []ShareUserDTO `json:"shared_users"`
}

// SharedFolderDTO represents a folder that has been shared with the current user
type SharedFolderDTO struct {
	ID         uint        `json:"id"`
	FolderName string      `json:"folder_name"`
	Owner      UserInfoDTO `json:"owner"`
	FileCount  int         `json:"file_count"`
	SharedAt   time.Time   `json:"shared_at"`
	CreatedAt  time.Time   `json:"created_at"`
}

// SharedFileDTO represents a file that has been shared with the current user
type SharedFileDTO struct {
	ID           uint        `json:"id"`
	FileName     string      `json:"file_name"`
	FileType     string      `json:"file_type"`
	ContentType  string      `json:"content_type"`
	FileSize     int64       `json:"file_size"`
	Owner        UserInfoDTO `json:"owner"`
	DownloadURL  string      `json:"download_url"`
	ThumbnailURL *string     `json:"thumbnail_url,omitempty"`
	SharedAt     time.Time   `json:"shared_at"`
	CreatedAt    time.Time   `json:"created_at"`
}

// SharedWithMeFoldersResponseDTO for listing folders shared with the current user
type SharedWithMeFoldersResponseDTO struct {
	Folders []SharedFolderDTO `json:"folders"`
}

// SharedWithMeFilesResponseDTO for listing files shared with the current user
type SharedWithMeFilesResponseDTO struct {
	Files []SharedFileDTO `json:"files"`
}

// RevokeShareResponseDTO for share revocation response
type RevokeShareResponseDTO struct {
	Message string `json:"message"`
}
