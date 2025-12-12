package request

// InitiateMultipartUploadRequestDTO for initiating multipart upload
type InitiateMultipartUploadRequestDTO struct {
	FileName    string `json:"file_name" validate:"required"`
	FileSize    int64  `json:"file_size" validate:"required,min=5242880"` // Minimum 5MB for multipart
	ContentType string `json:"content_type" validate:"required"`
	FileType    string `json:"file_type" validate:"required,oneof=image video"` // Still track file type
}

// GeneratePresignedURLsRequestDTO for generating presigned URLs for parts
type GeneratePresignedURLsRequestDTO struct {
	UploadID    string `json:"upload_id" validate:"required"`
	FileKey     string `json:"file_key" validate:"required"`
	PartNumbers []int  `json:"part_numbers" validate:"required,min=1,max=10000,dive,min=1,max=10000"`
}

// CompletedPart represents a completed upload part
type CompletedPart struct {
	PartNumber int    `json:"part_number" validate:"required,min=1"`
	ETag       string `json:"etag" validate:"required"`
}

// CompleteMultipartUploadRequestDTO for completing multipart upload
type CompleteMultipartUploadRequestDTO struct {
	UploadID string          `json:"upload_id" validate:"required"`
	FileKey  string          `json:"file_key" validate:"required"`
	Parts    []CompletedPart `json:"parts" validate:"required,min=1,max=10000,dive"`
}

// AbortMultipartUploadRequestDTO for aborting multipart upload
type AbortMultipartUploadRequestDTO struct {
	UploadID string `json:"upload_id" validate:"required"`
	FileKey  string `json:"file_key" validate:"required"`
}
