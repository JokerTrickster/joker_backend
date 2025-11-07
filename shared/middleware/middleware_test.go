package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/JokerTrickster/joker_backend/shared/logger"
)

func init() {
	logger.Init("error") // Initialize logger for tests
}

func TestRequestLogger(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler := RequestLogger()(func(c echo.Context) error {
		return c.String(http.StatusOK, "test")
	})

	if err := handler(c); err != nil {
		t.Errorf("RequestLogger middleware failed: %v", err)
	}

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rec.Code)
	}

	if c.Response().Header().Get(echo.HeaderXRequestID) == "" {
		t.Error("Request ID should be set")
	}
}

func TestRecovery(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/panic", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler := Recovery()(func(c echo.Context) error {
		panic("test panic")
	})

	// Recovery should catch the panic and not crash
	err := handler(c)
	if err != nil {
		t.Logf("Recovery caught error: %v", err)
	}

	// The handler should have set an error response
	if rec.Code != http.StatusInternalServerError && rec.Code != 0 {
		t.Logf("Response code after recovery: %d", rec.Code)
	}
}

func TestRequestID(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler := RequestID()(func(c echo.Context) error {
		return c.String(http.StatusOK, "test")
	})

	if err := handler(c); err != nil {
		t.Errorf("RequestID middleware failed: %v", err)
	}

	reqID := c.Response().Header().Get(echo.HeaderXRequestID)
	if reqID == "" {
		t.Error("Request ID should be set")
	}
}

func TestRequestIDWithExisting(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	existingID := "existing-request-id"
	req.Header.Set(echo.HeaderXRequestID, existingID)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler := RequestID()(func(c echo.Context) error {
		return c.String(http.StatusOK, "test")
	})

	if err := handler(c); err != nil {
		t.Errorf("RequestID middleware failed: %v", err)
	}

	reqID := c.Response().Header().Get(echo.HeaderXRequestID)
	if reqID != existingID {
		t.Errorf("Expected request ID %s, got %s", existingID, reqID)
	}
}

func TestTimeout(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/slow", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler := Timeout(100 * time.Millisecond)(func(c echo.Context) error {
		time.Sleep(200 * time.Millisecond)
		return c.String(http.StatusOK, "completed")
	})

	err := handler(c)
	if err == nil {
		t.Error("Expected timeout error, got nil")
	}

	if httpErr, ok := err.(*echo.HTTPError); ok {
		if httpErr.Code != http.StatusRequestTimeout {
			t.Errorf("Expected status 408, got %d", httpErr.Code)
		}
	}
}

func TestTimeoutSuccess(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/fast", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler := Timeout(200 * time.Millisecond)(func(c echo.Context) error {
		time.Sleep(50 * time.Millisecond)
		return c.String(http.StatusOK, "completed")
	})

	err := handler(c)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
}
