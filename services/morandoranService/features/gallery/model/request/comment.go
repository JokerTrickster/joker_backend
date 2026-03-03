package request

type ReqCreateComment struct {
	Content string `json:"content" validate:"required"`
}

type ReqListComment struct {
	Page  int `query:"page"`
	Limit int `query:"limit"`
}
