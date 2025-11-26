package response

// TagDTO represents a tag in response
type TagDTO struct {
	ID   uint   `json:"id"`
	Name string `json:"name"`
}

// FileInfoDTO represents file metadata
type FileInfoDTO struct {
	ID           uint     `json:"id"`
	FileName     string   `json:"file_name"`
	FileType     string   `json:"file_type"`
	ContentType  string   `json:"content_type"`
	FileSize     int64    `json:"file_size"`
	Duration     *float64 `json:"duration,omitempty"` // Video duration in seconds
	Tags         []TagDTO `json:"tags"`
	DownloadURL  string   `json:"download_url"`
	ThumbnailURL string   `json:"thumbnail_url,omitempty"`
	CreatedAt    string   `json:"created_at"`
	UpdatedAt    string   `json:"updated_at"`
}

// ListFilesResponseDTO for listing files
type ListFilesResponseDTO struct {
	Files      []FileInfoDTO `json:"files"`
	TotalCount int64         `json:"total_count"`
	Page       int           `json:"page"`
	PageSize   int           `json:"page_size"`
}
