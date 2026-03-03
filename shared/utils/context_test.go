package utils

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
)

func TestGetUserIDFromContext_Uint(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	c.Set("userID", uint(42))

	userID, err := GetUserIDFromContext(c)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if userID != 42 {
		t.Errorf("Expected userID 42, got %d", userID)
	}
	t.Logf("uint type: got userID=%d", userID)
}

func TestGetUserIDFromContext_Int(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	c.Set("userID", int(7))

	userID, err := GetUserIDFromContext(c)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if userID != 7 {
		t.Errorf("Expected userID 7, got %d", userID)
	}
	t.Logf("int type: got userID=%d", userID)
}

func TestGetUserIDFromContext_Int64(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	c.Set("userID", int64(100))

	userID, err := GetUserIDFromContext(c)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if userID != 100 {
		t.Errorf("Expected userID 100, got %d", userID)
	}
	t.Logf("int64 type: got userID=%d", userID)
}

func TestGetUserIDFromContext_Float64(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	c.Set("userID", float64(55))

	userID, err := GetUserIDFromContext(c)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if userID != 55 {
		t.Errorf("Expected userID 55, got %d", userID)
	}
	t.Logf("float64 type: got userID=%d", userID)
}

func TestGetUserIDFromContext_Missing(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	_, err := GetUserIDFromContext(c)
	if err == nil {
		t.Fatal("Missing userID should return error")
	}

	httpErr, ok := err.(*echo.HTTPError)
	if !ok {
		t.Fatalf("Expected *echo.HTTPError, got %T", err)
	}
	if httpErr.Code != http.StatusUnauthorized {
		t.Errorf("Expected 401, got %d", httpErr.Code)
	}
	t.Logf("Missing userID correctly rejected: %v", err)
}

func TestGetUserIDFromContext_InvalidType(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	c.Set("userID", "not-a-number")

	_, err := GetUserIDFromContext(c)
	if err == nil {
		t.Fatal("Invalid type should return error")
	}

	httpErr, ok := err.(*echo.HTTPError)
	if !ok {
		t.Fatalf("Expected *echo.HTTPError, got %T", err)
	}
	if httpErr.Code != http.StatusUnauthorized {
		t.Errorf("Expected 401, got %d", httpErr.Code)
	}
	t.Logf("Invalid type correctly rejected: %v", err)
}
