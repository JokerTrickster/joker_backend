package response

import "time"

type ResProductItem struct {
	ID            string    `json:"id"`
	Name          string    `json:"name"`
	Price         uint      `json:"price"`
	OriginalPrice *uint     `json:"originalPrice"`
	Description   string    `json:"description"`
	Image         string    `json:"image"`
	Category      string    `json:"category"`
	Badge         *string   `json:"badge"`
	InStock       bool      `json:"inStock"`
	CreatedAt     time.Time `json:"createdAt"`
}

type ResProductDetail struct {
	ID            string    `json:"id"`
	Name          string    `json:"name"`
	Price         uint      `json:"price"`
	OriginalPrice *uint     `json:"originalPrice"`
	Description   string    `json:"description"`
	Image         string    `json:"image"`
	Category      string    `json:"category"`
	Badge         *string   `json:"badge"`
	InStock       bool      `json:"inStock"`
	CreatedAt     time.Time `json:"createdAt"`
	UpdatedAt     time.Time `json:"updatedAt"`
}

type ResPagination struct {
	Page       int   `json:"page"`
	Limit      int   `json:"limit"`
	Total      int64 `json:"total"`
	TotalPages int   `json:"totalPages"`
}

type ResProductList struct {
	Items      []ResProductItem `json:"items"`
	Pagination ResPagination    `json:"pagination"`
}
