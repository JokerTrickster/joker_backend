package request

type ReqCreateProduct struct {
	Name          string  `json:"name" validate:"required"`
	Price         uint    `json:"price" validate:"required"`
	OriginalPrice *uint   `json:"originalPrice"`
	Description   string  `json:"description" validate:"required"`
	Image         string  `json:"image" validate:"required"`
	Category      string  `json:"category" validate:"required"`
	Badge         *string `json:"badge"`
	InStock       *bool   `json:"inStock"`
}

type ReqUpdateProduct struct {
	Name          *string `json:"name"`
	Price         *uint   `json:"price"`
	OriginalPrice *uint   `json:"originalPrice"`
	Description   *string `json:"description"`
	Image         *string `json:"image"`
	Category      *string `json:"category"`
	Badge         *string `json:"badge"`
	InStock       *bool   `json:"inStock"`
}

type ReqListProduct struct {
	Page     int    `query:"page"`
	Limit    int    `query:"limit"`
	Category string `query:"category"`
	InStock  *bool  `query:"inStock"`
}
