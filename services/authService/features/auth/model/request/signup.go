package request

type ReqSignUp struct {
	Email       string `json:"email" validate:"required,email"`
	Password    string `json:"password" validate:"required,min=6"`
	ServiceType string `json:"serviceType"`
	Name        string `json:"name" validate:"required"`
}
