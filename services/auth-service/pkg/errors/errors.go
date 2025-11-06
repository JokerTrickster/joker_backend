package errors

import (
	"fmt"
	"net/http"
)

// AppError represents a custom application error
type AppError struct {
	Code       string `json:"code"`
	Message    string `json:"message"`
	HTTPStatus int    `json:"-"`
	Err        error  `json:"-"`
}

// Error implements the error interface
func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

// Unwrap returns the wrapped error
func (e *AppError) Unwrap() error {
	return e.Err
}

// Common error codes
const (
	ErrCodeBadRequest          = "BAD_REQUEST"
	ErrCodeUnauthorized        = "UNAUTHORIZED"
	ErrCodeForbidden           = "FORBIDDEN"
	ErrCodeNotFound            = "NOT_FOUND"
	ErrCodeConflict            = "CONFLICT"
	ErrCodeValidation          = "VALIDATION_ERROR"
	ErrCodeInternalServer      = "INTERNAL_SERVER_ERROR"
	ErrCodeServiceUnavailable  = "SERVICE_UNAVAILABLE"
	ErrCodeDatabaseError       = "DATABASE_ERROR"
	ErrCodeInvalidInput        = "INVALID_INPUT"
	ErrCodeResourceExists      = "RESOURCE_EXISTS"
	ErrCodeResourceNotFound    = "RESOURCE_NOT_FOUND"
	ErrCodeUnprocessableEntity = "UNPROCESSABLE_ENTITY"
)

// New creates a new AppError
func New(code, message string, httpStatus int) *AppError {
	return &AppError{
		Code:       code,
		Message:    message,
		HTTPStatus: httpStatus,
	}
}

// Wrap wraps an existing error with additional context
func Wrap(err error, code, message string, httpStatus int) *AppError {
	return &AppError{
		Code:       code,
		Message:    message,
		HTTPStatus: httpStatus,
		Err:        err,
	}
}

// Predefined errors
func BadRequest(message string) *AppError {
	return New(ErrCodeBadRequest, message, http.StatusBadRequest)
}

func Unauthorized(message string) *AppError {
	return New(ErrCodeUnauthorized, message, http.StatusUnauthorized)
}

func Forbidden(message string) *AppError {
	return New(ErrCodeForbidden, message, http.StatusForbidden)
}

func NotFound(message string) *AppError {
	return New(ErrCodeNotFound, message, http.StatusNotFound)
}

func Conflict(message string) *AppError {
	return New(ErrCodeConflict, message, http.StatusConflict)
}

func ValidationError(message string) *AppError {
	return New(ErrCodeValidation, message, http.StatusBadRequest)
}

func InternalServerError(message string) *AppError {
	return New(ErrCodeInternalServer, message, http.StatusInternalServerError)
}

func ServiceUnavailable(message string) *AppError {
	return New(ErrCodeServiceUnavailable, message, http.StatusServiceUnavailable)
}

func DatabaseError(err error) *AppError {
	return Wrap(err, ErrCodeDatabaseError, "Database operation failed", http.StatusInternalServerError)
}

func InvalidInput(message string) *AppError {
	return New(ErrCodeInvalidInput, message, http.StatusBadRequest)
}

func ResourceExists(resourceType string) *AppError {
	return New(
		ErrCodeResourceExists,
		fmt.Sprintf("%s already exists", resourceType),
		http.StatusConflict,
	)
}

func ResourceNotFound(resourceType string) *AppError {
	return New(
		ErrCodeResourceNotFound,
		fmt.Sprintf("%s not found", resourceType),
		http.StatusNotFound,
	)
}

func UnprocessableEntity(message string) *AppError {
	return New(ErrCodeUnprocessableEntity, message, http.StatusUnprocessableEntity)
}
