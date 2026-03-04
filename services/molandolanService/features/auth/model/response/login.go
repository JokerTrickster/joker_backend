package response

type ResLogin struct {
	Token string      `json:"token"`
	User  ResUserInfo `json:"user"`
}

type ResUserInfo struct {
	ID       string `json:"id"`
	Nickname string `json:"nickname"`
	Email    string `json:"email"`
	Role     string `json:"role"`
}
