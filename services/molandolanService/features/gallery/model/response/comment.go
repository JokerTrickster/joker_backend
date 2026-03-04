package response

import "time"

type ResCommentItem struct {
	ID        string    `json:"id"`
	Author    ResAuthor `json:"author"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"createdAt"`
}

type ResCommentList struct {
	Items      []ResCommentItem `json:"items"`
	Pagination ResPagination    `json:"pagination"`
}
