package request

// UpdateFileTagsRequestDTO represents the request body for bulk updating file tags
type UpdateFileTagsRequestDTO struct {
	Tags []string `json:"tags" validate:"required,dive,max=50"` // Tags array (max 50 chars each)
}

// AddFileTagRequestDTO represents the request body for adding a single tag to a file
type AddFileTagRequestDTO struct {
	Tag string `json:"tag" validate:"required,max=50"` // Single tag to add (max 50 chars)
}
