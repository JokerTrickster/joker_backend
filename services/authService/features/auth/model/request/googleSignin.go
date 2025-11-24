package request

type ReqGoogleSignin struct {
	IdToken string `json:"idToken" validate:"required"`
}

