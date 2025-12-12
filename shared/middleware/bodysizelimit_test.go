package middleware

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
)

func TestBodySizeLimit(t *testing.T) {
	tests := []struct {
		name       string
		maxBytes   int64
		bodySize   int
		expectPass bool
	}{
		{
			name:       "small body within limit",
			maxBytes:   1024, // 1KB
			bodySize:   512,  // 512 bytes
			expectPass: true,
		},
		{
			name:       "body exactly at limit",
			maxBytes:   1024, // 1KB
			bodySize:   1024, // 1KB
			expectPass: true,
		},
		{
			name:       "zero limit uses default",
			maxBytes:   0,
			bodySize:   512,
			expectPass: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := echo.New()

			// Create body with specified size
			body := strings.Repeat("a", tt.bodySize)
			req := httptest.NewRequest(http.MethodPost, "/test", bytes.NewBufferString(body))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			// Create handler that reads the body
			handler := BodySizeLimit(tt.maxBytes)(func(c echo.Context) error {
				return c.String(http.StatusOK, "success")
			})

			err := handler(c)

			if tt.expectPass {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			} else {
				if err == nil {
					t.Errorf("Expected error for body size %d with limit %d, but got none", tt.bodySize, tt.maxBytes)
				}
			}
		})
	}
}

func TestDefaultBodySizeLimit(t *testing.T) {
	e := echo.New()

	// Create a body smaller than default limit (10MB)
	smallBody := strings.Repeat("a", 1024) // 1KB
	req := httptest.NewRequest(http.MethodPost, "/test", bytes.NewBufferString(smallBody))
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler := DefaultBodySizeLimit()(func(c echo.Context) error {
		return c.String(http.StatusOK, "success")
	})

	err := handler(c)
	if err != nil {
		t.Errorf("Unexpected error with small body: %v", err)
	}
}

func TestFileUploadBodySizeLimit(t *testing.T) {
	e := echo.New()

	// Create a body smaller than file upload limit (100MB)
	smallBody := strings.Repeat("a", 1024*1024) // 1MB
	req := httptest.NewRequest(http.MethodPost, "/upload", bytes.NewBufferString(smallBody))
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler := FileUploadBodySizeLimit()(func(c echo.Context) error {
		return c.String(http.StatusOK, "success")
	})

	err := handler(c)
	if err != nil {
		t.Errorf("Unexpected error with 1MB upload: %v", err)
	}
}
