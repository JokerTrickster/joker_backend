package response

import "time"

// UpdateFileTagsResponseDTO represents the response after updating file tags
type UpdateFileTagsResponseDTO struct {
	FileID    uint       `json:"file_id"`
	Tags      []string   `json:"tags"`
	UpdatedAt time.Time  `json:"updated_at"`
}

// AddFileTagResponseDTO represents the response after adding a tag to a file
type AddFileTagResponseDTO struct {
	FileID    uint       `json:"file_id"`
	Tag       string     `json:"tag"`
	UpdatedAt time.Time  `json:"updated_at"`
}
