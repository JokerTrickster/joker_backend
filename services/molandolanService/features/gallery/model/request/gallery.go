package request

type ReqCreateGallery struct {
	MediaURL     string `json:"mediaUrl" validate:"required"`
	ThumbnailURL string `json:"thumbnailUrl" validate:"required"`
	MediaType    string `json:"mediaType" validate:"required"`
	Caption      string `json:"caption"`
}

type ReqListGallery struct {
	Page  int `query:"page"`
	Limit int `query:"limit"`
}
