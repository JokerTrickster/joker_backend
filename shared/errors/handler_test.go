package errors

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/JokerTrickster/joker_backend/shared/logger"
	"github.com/labstack/echo/v4"
)

func init() {
	logger.Init("error")
}

func TestCustomErrorHandler_AppError(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	appErr := New(ErrCodeBadRequest, "invalid input", http.StatusBadRequest)
	CustomErrorHandler(appErr, c)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}

	var body map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	errObj, ok := body["error"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected 'error' object in response")
	}
	if errObj["code"] != ErrCodeBadRequest {
		t.Errorf("Expected code %s, got %v", ErrCodeBadRequest, errObj["code"])
	}
	if errObj["message"] != "invalid input" {
		t.Errorf("Expected message 'invalid input', got %v", errObj["message"])
	}
}

func TestCustomErrorHandler_AppErrorInternal(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	appErr := InternalServerError("database connection failed")
	CustomErrorHandler(appErr, c)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("Expected status %d, got %d", http.StatusInternalServerError, rec.Code)
	}

	var body map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}
	errObj := body["error"].(map[string]interface{})
	if errObj["code"] != ErrCodeInternalServer {
		t.Errorf("Expected code %s, got %v", ErrCodeInternalServer, errObj["code"])
	}
}

func TestCustomErrorHandler_EchoHTTPError(t *testing.T) {
	tests := []struct {
		name         string
		httpError    *echo.HTTPError
		wantStatus   int
		wantErrCode  string
		wantMessage  string
	}{
		{
			name:        "bad request maps to BAD_REQUEST",
			httpError:   echo.NewHTTPError(http.StatusBadRequest, "invalid payload"),
			wantStatus:  http.StatusBadRequest,
			wantErrCode: ErrCodeBadRequest,
			wantMessage: "invalid payload",
		},
		{
			name:        "unauthorized maps to UNAUTHORIZED",
			httpError:   echo.NewHTTPError(http.StatusUnauthorized, "token expired"),
			wantStatus:  http.StatusUnauthorized,
			wantErrCode: ErrCodeUnauthorized,
			wantMessage: "token expired",
		},
		{
			name:        "forbidden maps to FORBIDDEN",
			httpError:   echo.NewHTTPError(http.StatusForbidden, "access denied"),
			wantStatus:  http.StatusForbidden,
			wantErrCode: ErrCodeForbidden,
			wantMessage: "access denied",
		},
		{
			name:        "not found maps to NOT_FOUND",
			httpError:   echo.NewHTTPError(http.StatusNotFound, "resource not found"),
			wantStatus:  http.StatusNotFound,
			wantErrCode: ErrCodeNotFound,
			wantMessage: "resource not found",
		},
		{
			name:        "conflict maps to CONFLICT",
			httpError:   echo.NewHTTPError(http.StatusConflict, "duplicate entry"),
			wantStatus:  http.StatusConflict,
			wantErrCode: ErrCodeConflict,
			wantMessage: "duplicate entry",
		},
		{
			name:        "unprocessable entity maps to UNPROCESSABLE_ENTITY",
			httpError:   echo.NewHTTPError(http.StatusUnprocessableEntity, "validation failed"),
			wantStatus:  http.StatusUnprocessableEntity,
			wantErrCode: ErrCodeUnprocessableEntity,
			wantMessage: "validation failed",
		},
		{
			name:        "too many requests maps to RATE_LIMIT_EXCEEDED",
			httpError:   echo.NewHTTPError(http.StatusTooManyRequests, "rate limit exceeded"),
			wantStatus:  http.StatusTooManyRequests,
			wantErrCode: "RATE_LIMIT_EXCEEDED",
			wantMessage: "rate limit exceeded",
		},
		{
			name:        "service unavailable maps to SERVICE_UNAVAILABLE",
			httpError:   echo.NewHTTPError(http.StatusServiceUnavailable, "maintenance"),
			wantStatus:  http.StatusServiceUnavailable,
			wantErrCode: ErrCodeServiceUnavailable,
			wantMessage: "maintenance",
		},
		{
			name:        "unknown status maps to INTERNAL_SERVER_ERROR",
			httpError:   echo.NewHTTPError(418, "I'm a teapot"),
			wantStatus:  418,
			wantErrCode: ErrCodeInternalServer,
			wantMessage: "I'm a teapot",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := echo.New()
			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			CustomErrorHandler(tt.httpError, c)

			if rec.Code != tt.wantStatus {
				t.Errorf("Expected status %d, got %d", tt.wantStatus, rec.Code)
			}

			var body map[string]interface{}
			if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
				t.Fatalf("Failed to parse response: %v", err)
			}
			errObj := body["error"].(map[string]interface{})
			if errObj["code"] != tt.wantErrCode {
				t.Errorf("Expected code %s, got %v", tt.wantErrCode, errObj["code"])
			}
			if errObj["message"] != tt.wantMessage {
				t.Errorf("Expected message %q, got %v", tt.wantMessage, errObj["message"])
			}
		})
	}
}

func TestCustomErrorHandler_UnknownError(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Plain error (not AppError, not echo.HTTPError) hits the "unknown error" branch
	CustomErrorHandler(errors.New("something broke unexpectedly"), c)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("Expected status %d for unknown error, got %d", http.StatusInternalServerError, rec.Code)
	}

	var body map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}
	errObj := body["error"].(map[string]interface{})
	if errObj["code"] != ErrCodeInternalServer {
		t.Errorf("Expected code %s for unknown error path, got %v", ErrCodeInternalServer, errObj["code"])
	}
	if errObj["message"] != "Internal server error" {
		t.Errorf("Expected default message, got %v", errObj["message"])
	}
}

func TestCustomErrorHandler_CommittedResponse(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	c.Response().WriteHeader(http.StatusOK)
	if !c.Response().Committed {
		t.Fatal("Response should be committed after WriteHeader")
	}

	appErr := New(ErrCodeBadRequest, "should not be written", http.StatusBadRequest)
	CustomErrorHandler(appErr, c)

	if rec.Code != http.StatusOK {
		t.Errorf("Committed response should not be overwritten, got %d", rec.Code)
	}
}
