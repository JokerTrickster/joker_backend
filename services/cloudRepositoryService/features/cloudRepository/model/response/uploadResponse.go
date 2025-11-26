package response

// UploadResponseDTO returns presigned upload URL
type UploadResponseDTO struct {
	FileID           uint   `json:"file_id"`
	UploadURL        string `json:"upload_url"`
	S3Key            string `json:"s3_key"`
	ThumbnailURL     string `json:"thumbnail_upload_url,omitempty"`
	ThumbnailKey     string `json:"thumbnail_key,omitempty"`
	ExpiresIn        int    `json:"expires_in"` // seconds
}

// BatchUploadResponseDTO returns multiple presigned upload URLs
type BatchUploadResponseDTO struct {
	Results      []UploadResponseDTO `json:"results"`
	TotalCount   int                 `json:"total_count"`
	SuccessCount int                 `json:"success_count"`
	FailedCount  int                 `json:"failed_count"`
}
