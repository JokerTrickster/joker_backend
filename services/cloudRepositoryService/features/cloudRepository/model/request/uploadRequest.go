package request

// UploadRequestDTO for requesting presigned upload URL
type UploadRequestDTO struct {
	FileName    string    `json:"file_name" validate:"required"`
	ContentType string    `json:"content_type" validate:"required"`
	FileType    string    `json:"file_type" validate:"required,oneof=image video"`
	FileSize    int64     `json:"file_size" validate:"required,min=1"`
	Tags        []string  `json:"tags" validate:"omitempty,dive,max=50"`         // Optional tags (max 50 chars each)
	Duration    *float64  `json:"duration" validate:"omitempty,min=0,max=86400"` // Optional video duration in seconds (max 24 hours)
}

// BatchUploadRequestDTO for requesting multiple presigned upload URLs (max 30 files)
type BatchUploadRequestDTO struct {
	Files []UploadRequestDTO `json:"files" validate:"required,min=1,max=30,dive"`
}
