package response

// InitiateMultipartUploadResponseDTO response for initiated multipart upload
type InitiateMultipartUploadResponseDTO struct {
	UploadID   string `json:"upload_id"`
	FileKey    string `json:"file_key"`
	PartSize   int    `json:"part_size"`
	TotalParts int    `json:"total_parts"`
}

// PresignedURLPart represents a presigned URL for a specific part
type PresignedURLPart struct {
	PartNumber int    `json:"part_number"`
	URL        string `json:"url"`
}

// GeneratePresignedURLsResponseDTO response with presigned URLs for parts
type GeneratePresignedURLsResponseDTO struct {
	URLs      []PresignedURLPart `json:"urls"`
	ExpiresIn int                `json:"expires_in"` // seconds
}

// CompleteMultipartUploadResponseDTO response for completed multipart upload
type CompleteMultipartUploadResponseDTO struct {
	FileID uint   `json:"file_id"`
	URL    string `json:"url"`
	Size   int64  `json:"size"`
}

// AbortMultipartUploadResponseDTO response for aborted multipart upload
type AbortMultipartUploadResponseDTO struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}
