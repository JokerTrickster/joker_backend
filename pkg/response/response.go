package response

type SuccessResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Message string      `json:"message,omitempty"`
}

type ErrorResponse struct {
	Success bool       `json:"success"`
	Error   ErrorDetail `json:"error"`
}

type ErrorDetail struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func Success(data interface{}, message string) SuccessResponse {
	return SuccessResponse{
		Success: true,
		Data:    data,
		Message: message,
	}
}

func Error(code, message string) ErrorResponse {
	return ErrorResponse{
		Success: false,
		Error: ErrorDetail{
			Code:    code,
			Message: message,
		},
	}
}
