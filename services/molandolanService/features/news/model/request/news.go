package request

type ReqCreateNews struct {
	Title     string `json:"title" validate:"required"`
	Summary   string `json:"summary"`
	Content   string `json:"content" validate:"required"`
	Thumbnail string `json:"thumbnail"`
	Category  string `json:"category" validate:"required"`
	Date      string `json:"date" validate:"required"`
}

type ReqUpdateNews struct {
	Title     *string `json:"title"`
	Summary   *string `json:"summary"`
	Content   *string `json:"content"`
	Thumbnail *string `json:"thumbnail"`
	Category  *string `json:"category"`
	Date      *string `json:"date"`
}

type ReqListNews struct {
	Page     int    `query:"page"`
	Limit    int    `query:"limit"`
	Category string `query:"category"`
}
