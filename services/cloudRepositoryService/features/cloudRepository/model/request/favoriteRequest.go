package request

// AddFavoriteRequestDTO for adding a file to favorites
type AddFavoriteRequestDTO struct {
	FileID uint `json:"fileId" validate:"required,min=1"`
}

// ListFavoritesRequestDTO for listing favorites with pagination and filtering
type ListFavoritesRequestDTO struct {
	Page  int    `query:"page" validate:"omitempty,min=1"`
	Size  int    `query:"size" validate:"omitempty,min=1,max=100"`
	Sort  string `query:"sort" validate:"omitempty,oneof=uploadDate fileName"`
	Order string `query:"order" validate:"omitempty,oneof=asc desc"`
	Q     string `query:"q" validate:"omitempty,max=255"`     // Filename search
	Ext   string `query:"ext" validate:"omitempty,max=10"`    // File extension filter
	Tag   string `query:"tag" validate:"omitempty,max=50"`    // Tag filter
}
