package response

type ResCheckEmail struct {
	Email     string `json:"email"`
	Exists    bool   `json:"exists"`
	Available bool   `json:"available"` // exists의 반대값 (사용 가능 여부)
}
