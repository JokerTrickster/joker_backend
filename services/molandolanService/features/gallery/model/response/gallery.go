package response

import "time"

type ResAuthor struct {
	ID       string `json:"id"`
	Nickname string `json:"nickname"`
}

type ResGalleryListItem struct {
	ID           string    `json:"id"`
	MediaType    string    `json:"mediaType"`
	ThumbnailURL string    `json:"thumbnailUrl"`
	LikeCount    int       `json:"likeCount"`
	CommentCount int       `json:"commentCount"`
	CreatedAt    time.Time `json:"createdAt"`
}

type ResPagination struct {
	Page       int   `json:"page"`
	Limit      int   `json:"limit"`
	Total      int64 `json:"total"`
	TotalPages int   `json:"totalPages"`
}

type ResGalleryList struct {
	Items      []ResGalleryListItem `json:"items"`
	Pagination ResPagination        `json:"pagination"`
}

type ResGalleryDetail struct {
	ID           string    `json:"id"`
	Author       ResAuthor `json:"author"`
	MediaType    string    `json:"mediaType"`
	MediaURL     string    `json:"mediaUrl"`
	ThumbnailURL string    `json:"thumbnailUrl"`
	Caption      string    `json:"caption"`
	LikeCount    int       `json:"likeCount"`
	CommentCount int       `json:"commentCount"`
	IsLiked      bool      `json:"isLiked"`
	CreatedAt    time.Time `json:"createdAt"`
}

type ResGalleryCreate struct {
	ID           string    `json:"id"`
	Author       ResAuthor `json:"author"`
	MediaType    string    `json:"mediaType"`
	MediaURL     string    `json:"mediaUrl"`
	ThumbnailURL string    `json:"thumbnailUrl"`
	Caption      string    `json:"caption"`
	LikeCount    int       `json:"likeCount"`
	CommentCount int       `json:"commentCount"`
	CreatedAt    time.Time `json:"createdAt"`
}

type ResLikeToggle struct {
	IsLiked   bool `json:"isLiked"`
	LikeCount int  `json:"likeCount"`
}
