package response

import "time"

type ResNewsItem struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	Summary   string    `json:"summary"`
	Thumbnail string    `json:"thumbnail"`
	Category  string    `json:"category"`
	Date      string    `json:"date"`
	CreatedAt time.Time `json:"createdAt"`
}

type ResNewsDetail struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	Summary   string    `json:"summary"`
	Content   string    `json:"content"`
	Thumbnail string    `json:"thumbnail"`
	Category  string    `json:"category"`
	Date      string    `json:"date"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type ResPagination struct {
	Page       int   `json:"page"`
	Limit      int   `json:"limit"`
	Total      int64 `json:"total"`
	TotalPages int   `json:"totalPages"`
}

type ResNewsList struct {
	Items      []ResNewsItem `json:"items"`
	Pagination ResPagination `json:"pagination"`
}
