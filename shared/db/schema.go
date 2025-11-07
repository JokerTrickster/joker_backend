package db

type Tokens struct {
	ID               int    `json:"id"`
	UserID           int    `json:"userID"`
	AccessToken      string `json:"accessToken"`
	RefreshToken     string `json:"refreshToken"`
	RefreshExpiredAt int64  `json:"refreshExpiredAt"`
}

type Users struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
	Provider string `json:"provider"` // google, kakao, naver, email
}
