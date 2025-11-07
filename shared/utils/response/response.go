package response

// BaseResponse is the base structure for API responses
type BaseResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
	Error   *ErrorInfo  `json:"error,omitempty"`
}

// ErrorInfo contains error details
type ErrorInfo struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

// SuccessResponse creates a success response
func SuccessResponse(data interface{}, message string) BaseResponse {
	return BaseResponse{
		Success: true,
		Message: message,
		Data:    data,
	}
}

// ErrorResponse creates an error response
func ErrorResponse(code, message, details string) BaseResponse {
	return BaseResponse{
		Success: false,
		Error: &ErrorInfo{
			Code:    code,
			Message: message,
			Details: details,
		},
	}
}

// ResSignIn response for sign in
type ResSignIn struct {
	Token        string `json:"token"`
	RefreshToken string `json:"refreshToken,omitempty"`
	UserID       int64  `json:"userId"`
	Email        string `json:"email"`
	Name         string `json:"name"`
}

// ResSignUp response for sign up
type ResSignUp struct {
	UserID int64  `json:"userId"`
	Email  string `json:"email"`
	Name   string `json:"name"`
}

// ResUser response for user information
type ResUser struct {
	UserID    int64  `json:"userId"`
	Email     string `json:"email"`
	Name      string `json:"name"`
	CreatedAt int64  `json:"createdAt"`
	UpdatedAt int64  `json:"updatedAt"`
}

// ResPagination response for paginated data
type ResPagination struct {
	Items      interface{} `json:"items"`
	TotalCount int         `json:"totalCount"`
	Page       int         `json:"page"`
	PageSize   int         `json:"pageSize"`
	TotalPages int         `json:"totalPages"`
}