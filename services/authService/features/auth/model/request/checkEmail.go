package request

type ReqCheckEmail struct {
	Email    string `json:"email" validate:"required,email"`
	Provider string `json:"provider" validate:"required,oneof=email google"`
}
