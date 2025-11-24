package request

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
