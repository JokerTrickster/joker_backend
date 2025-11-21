package model

// DTO (Data Transfer Objects)

// UploadRequestDTO for requesting presigned upload URL
type UploadRequestDTO struct {
	FileName    string `json:"file_name" validate:"required"`
	ContentType string `json:"content_type" validate:"required"`
	FileType    string `json:"file_type" validate:"required,oneof=image video"`
	FileSize    int64  `json:"file_size" validate:"required,min=1"`
}

// UploadResponseDTO returns presigned upload URL
type UploadResponseDTO struct {
	FileID       uint   `json:"file_id"`
	UploadURL    string `json:"upload_url"`
	S3Key        string `json:"s3_key"`
	ExpiresIn    int    `json:"expires_in"` // seconds
}

// BatchUploadRequestDTO for requesting multiple presigned upload URLs (max 30 files)
type BatchUploadRequestDTO struct {
	Files []UploadRequestDTO `json:"files" validate:"required,min=1,max=30,dive"`
}

// BatchUploadResponseDTO returns multiple presigned upload URLs
type BatchUploadResponseDTO struct {
	Results   []UploadResponseDTO `json:"results"`
	TotalCount int                `json:"total_count"`
	SuccessCount int              `json:"success_count"`
	FailedCount  int              `json:"failed_count"`
}

// DownloadResponseDTO returns presigned download URL
type DownloadResponseDTO struct {
	DownloadURL string `json:"download_url"`
	FileName    string `json:"file_name"`
	ExpiresIn   int    `json:"expires_in"` // seconds
}

// ListFilesRequestDTO for filtering and pagination
type ListFilesRequestDTO struct {
	FileType  string   `query:"file_type" validate:"omitempty,oneof=image video"`
	Keyword   string   `query:"keyword"` // Search in filename or tags
	Tags      []string `query:"tags"`    // Filter by specific tags
	Sort      string   `query:"sort" validate:"omitempty,oneof=latest oldest name size"`
	StartDate string   `query:"start_date"` // YYYY-MM-DD
	EndDate   string   `query:"end_date"`   // YYYY-MM-DD
	Page      int      `query:"page"`
	PageSize  int      `query:"page_size"`
}

// TagDTO represents a tag in response
type TagDTO struct {
	ID   uint   `json:"id"`
	Name string `json:"name"`
}

// FileInfoDTO represents file metadata
type FileInfoDTO struct {
	ID          uint     `json:"id"`
	FileName    string   `json:"file_name"`
	FileType    string   `json:"file_type"`
	ContentType string   `json:"content_type"`
	FileSize    int64    `json:"file_size"`
	Tags        []TagDTO `json:"tags"`
	CreatedAt   string   `json:"created_at"`
	UpdatedAt   string   `json:"updated_at"`
}

// ListFilesResponseDTO for listing files
type ListFilesResponseDTO struct {
	Files      []FileInfoDTO `json:"files"`
	TotalCount int64         `json:"total_count"`
	Page       int           `json:"page"`
	PageSize   int           `json:"page_size"`
}
