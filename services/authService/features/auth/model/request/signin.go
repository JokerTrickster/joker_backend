package request

type ReqSignIn struct {
	Email       string `json:"email" validate:"required,email"`
	Password    string `json:"password" validate:"required,min=6"`
	ServiceType string `json:"serviceType" validate:"required,oneof=game"`
}
