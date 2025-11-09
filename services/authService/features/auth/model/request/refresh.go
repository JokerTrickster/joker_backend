package request

type ReqRefreshToken struct {
	RefreshToken string `json:"refreshToken" validate:"required"`
}
