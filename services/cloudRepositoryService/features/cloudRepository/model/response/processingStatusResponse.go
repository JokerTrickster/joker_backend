package response

import "time"

// ProcessingStatusResponse returns the current processing status of a file
type ProcessingStatusResponse struct {
	FileID                  uint      `json:"file_id"`
	Status                  string    `json:"status"`
	Progress                int       `json:"progress"`
	Stage                   *string   `json:"stage"`
	EstimatedCompletionTime *string   `json:"estimated_completion_time,omitempty"`
	ThumbnailURL            *string   `json:"thumbnail_url,omitempty"`
	Duration                *float64  `json:"duration,omitempty"`
	Metadata                *Metadata `json:"metadata,omitempty"`
	Error                   *ErrorInfo `json:"error,omitempty"`
	CreatedAt               time.Time `json:"created_at"`
	UpdatedAt               time.Time `json:"updated_at"`
	CompletedAt             *time.Time `json:"completed_at,omitempty"`
	FailedAt                *time.Time `json:"failed_at,omitempty"`
}

// Metadata contains file metadata information
type Metadata struct {
	Width  *int    `json:"width,omitempty"`
	Height *int    `json:"height,omitempty"`
	Codec  *string `json:"codec,omitempty"`
}

// ErrorInfo contains error information
type ErrorInfo struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// BatchProcessingStatusRequest for requesting multiple file statuses
type BatchProcessingStatusRequest struct {
	FileIDs []uint `json:"file_ids" validate:"required,min=1,max=100"`
}

// BatchProcessingStatusResponse returns multiple file processing statuses
type BatchProcessingStatusResponse struct {
	Results []ProcessingStatusResponse `json:"results"`
}
