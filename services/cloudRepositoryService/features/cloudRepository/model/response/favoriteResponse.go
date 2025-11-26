package response

import "time"

// FavoriteResponseDTO for add/remove favorite operations
type FavoriteResponseDTO struct {
	Success     bool      `json:"success"`
	FavoritedAt time.Time `json:"favoritedAt"`
}

// PaginationMeta represents pagination metadata
type PaginationMeta struct {
	Total      int64 `json:"total"`
	Page       int   `json:"page"`
	Size       int   `json:"size"`
	TotalPages int   `json:"total_pages"`
}

// ListFavoritesResponseDTO for listing favorite files
type ListFavoritesResponseDTO struct {
	Data       []FileInfoDTO  `json:"data"`
	Pagination PaginationMeta `json:"pagination"`
}
