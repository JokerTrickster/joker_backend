package errors

import (
	"errors"
	"net/http"
	"testing"
)

func TestNew(t *testing.T) {
	err := New(ErrCodeBadRequest, "test message", http.StatusBadRequest)

	if err.Code != ErrCodeBadRequest {
		t.Errorf("Expected code %s, got %s", ErrCodeBadRequest, err.Code)
	}

	if err.Message != "test message" {
		t.Errorf("Expected message 'test message', got '%s'", err.Message)
	}

	if err.HTTPStatus != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, err.HTTPStatus)
	}
}

func TestWrap(t *testing.T) {
	originalErr := errors.New("original error")
	wrappedErr := Wrap(originalErr, ErrCodeDatabaseError, "database error", http.StatusInternalServerError)

	if wrappedErr.Code != ErrCodeDatabaseError {
		t.Errorf("Expected code %s, got %s", ErrCodeDatabaseError, wrappedErr.Code)
	}

	if wrappedErr.Err != originalErr {
		t.Error("Wrapped error should contain original error")
	}

	if !errors.Is(wrappedErr, originalErr) {
		t.Error("Wrapped error should be identifiable with errors.Is")
	}
}

func TestAppErrorError(t *testing.T) {
	tests := []struct {
		name     string
		appErr   *AppError
		expected string
	}{
		{
			name: "error without wrapped error",
			appErr: &AppError{
				Code:    ErrCodeBadRequest,
				Message: "invalid input",
			},
			expected: "invalid input",
		},
		{
			name: "error with wrapped error",
			appErr: &AppError{
				Code:    ErrCodeDatabaseError,
				Message: "database failed",
				Err:     errors.New("connection refused"),
			},
			expected: "database failed: connection refused",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.appErr.Error()
			if result != tt.expected {
				t.Errorf("Expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

func TestPredefinedErrors(t *testing.T) {
	tests := []struct {
		name       string
		errFunc    func(string) *AppError
		message    string
		wantCode   string
		wantStatus int
	}{
		{
			name:       "BadRequest",
			errFunc:    BadRequest,
			message:    "bad request",
			wantCode:   ErrCodeBadRequest,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "Unauthorized",
			errFunc:    Unauthorized,
			message:    "unauthorized",
			wantCode:   ErrCodeUnauthorized,
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "Forbidden",
			errFunc:    Forbidden,
			message:    "forbidden",
			wantCode:   ErrCodeForbidden,
			wantStatus: http.StatusForbidden,
		},
		{
			name:       "NotFound",
			errFunc:    NotFound,
			message:    "not found",
			wantCode:   ErrCodeNotFound,
			wantStatus: http.StatusNotFound,
		},
		{
			name:       "Conflict",
			errFunc:    Conflict,
			message:    "conflict",
			wantCode:   ErrCodeConflict,
			wantStatus: http.StatusConflict,
		},
		{
			name:       "ValidationError",
			errFunc:    ValidationError,
			message:    "validation error",
			wantCode:   ErrCodeValidation,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "InternalServerError",
			errFunc:    InternalServerError,
			message:    "internal error",
			wantCode:   ErrCodeInternalServer,
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.errFunc(tt.message)

			if err.Code != tt.wantCode {
				t.Errorf("Expected code %s, got %s", tt.wantCode, err.Code)
			}

			if err.Message != tt.message {
				t.Errorf("Expected message '%s', got '%s'", tt.message, err.Message)
			}

			if err.HTTPStatus != tt.wantStatus {
				t.Errorf("Expected status %d, got %d", tt.wantStatus, err.HTTPStatus)
			}
		})
	}
}

func TestResourceErrors(t *testing.T) {
	t.Run("ResourceExists", func(t *testing.T) {
		err := ResourceExists("User")
		if err.Code != ErrCodeResourceExists {
			t.Errorf("Expected code %s, got %s", ErrCodeResourceExists, err.Code)
		}
		if err.Message != "User already exists" {
			t.Errorf("Expected message 'User already exists', got '%s'", err.Message)
		}
		if err.HTTPStatus != http.StatusConflict {
			t.Errorf("Expected status %d, got %d", http.StatusConflict, err.HTTPStatus)
		}
	})

	t.Run("ResourceNotFound", func(t *testing.T) {
		err := ResourceNotFound("Product")
		if err.Code != ErrCodeResourceNotFound {
			t.Errorf("Expected code %s, got %s", ErrCodeResourceNotFound, err.Code)
		}
		if err.Message != "Product not found" {
			t.Errorf("Expected message 'Product not found', got '%s'", err.Message)
		}
		if err.HTTPStatus != http.StatusNotFound {
			t.Errorf("Expected status %d, got %d", http.StatusNotFound, err.HTTPStatus)
		}
	})
}

func TestDatabaseError(t *testing.T) {
	originalErr := errors.New("connection timeout")
	err := DatabaseError(originalErr)

	if err.Code != ErrCodeDatabaseError {
		t.Errorf("Expected code %s, got %s", ErrCodeDatabaseError, err.Code)
	}

	if err.HTTPStatus != http.StatusInternalServerError {
		t.Errorf("Expected status %d, got %d", http.StatusInternalServerError, err.HTTPStatus)
	}

	if err.Err != originalErr {
		t.Error("Database error should wrap original error")
	}
}

func TestUnwrap(t *testing.T) {
	originalErr := errors.New("original")
	wrappedErr := Wrap(originalErr, ErrCodeDatabaseError, "wrapped", http.StatusInternalServerError)

	unwrapped := wrappedErr.Unwrap()
	if unwrapped != originalErr {
		t.Error("Unwrap should return the original error")
	}
}
