package request

type ReqUpdateMe struct {
	Nickname string `json:"nickname" validate:"required,min=2,max=20"`
}
