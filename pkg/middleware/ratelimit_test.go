package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
)

func TestRateLimiter(t *testing.T) {
	e := echo.New()
	limiter := NewRateLimiter(2, 2) // 2 requests per second, burst of 2

	handler := limiter.Middleware()(func(c echo.Context) error {
		return c.String(http.StatusOK, "ok")
	})

	// First request should succeed
	req1 := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec1 := httptest.NewRecorder()
	c1 := e.NewContext(req1, rec1)

	if err := handler(c1); err != nil {
		t.Errorf("First request should succeed: %v", err)
	}

	if rec1.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rec1.Code)
	}

	// Second request should succeed (within burst)
	req2 := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec2 := httptest.NewRecorder()
	c2 := e.NewContext(req2, rec2)

	if err := handler(c2); err != nil {
		t.Errorf("Second request should succeed: %v", err)
	}

	if rec2.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rec2.Code)
	}

	// Third request should be rate limited
	req3 := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec3 := httptest.NewRecorder()
	c3 := e.NewContext(req3, rec3)

	err := handler(c3)
	if err == nil {
		t.Error("Third request should be rate limited")
	}

	if httpErr, ok := err.(*echo.HTTPError); ok {
		if httpErr.Code != http.StatusTooManyRequests {
			t.Errorf("Expected status 429, got %d", httpErr.Code)
		}
	}
}

func TestRateLimiterDifferentIPs(t *testing.T) {
	e := echo.New()
	limiter := NewRateLimiter(1, 1) // 1 request per second, burst of 1

	handler := limiter.Middleware()(func(c echo.Context) error {
		return c.String(http.StatusOK, "ok")
	})

	// Request from IP1
	req1 := httptest.NewRequest(http.MethodGet, "/test", nil)
	req1.Header.Set("X-Real-IP", "192.168.1.1")
	rec1 := httptest.NewRecorder()
	c1 := e.NewContext(req1, rec1)

	if err := handler(c1); err != nil {
		t.Errorf("Request from IP1 should succeed: %v", err)
	}

	// Request from IP2 should succeed (different IP)
	req2 := httptest.NewRequest(http.MethodGet, "/test", nil)
	req2.Header.Set("X-Real-IP", "192.168.1.2")
	rec2 := httptest.NewRecorder()
	c2 := e.NewContext(req2, rec2)

	if err := handler(c2); err != nil {
		t.Errorf("Request from IP2 should succeed: %v", err)
	}

	if rec2.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rec2.Code)
	}
}

func TestRateLimiterReset(t *testing.T) {
	e := echo.New()
	limiter := NewRateLimiter(10, 1) // 10 requests per second, burst of 1

	handler := limiter.Middleware()(func(c echo.Context) error {
		return c.String(http.StatusOK, "ok")
	})

	// First request should succeed
	req1 := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec1 := httptest.NewRecorder()
	c1 := e.NewContext(req1, rec1)

	if err := handler(c1); err != nil {
		t.Errorf("First request should succeed: %v", err)
	}

	// Second request should fail (burst exceeded)
	req2 := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec2 := httptest.NewRecorder()
	c2 := e.NewContext(req2, rec2)

	err := handler(c2)
	if err == nil {
		t.Error("Second request should be rate limited")
	}

	// Wait for rate limiter to reset
	time.Sleep(150 * time.Millisecond)

	// Third request should succeed after waiting
	req3 := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec3 := httptest.NewRecorder()
	c3 := e.NewContext(req3, rec3)

	if err := handler(c3); err != nil {
		t.Errorf("Request after waiting should succeed: %v", err)
	}

	if rec3.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rec3.Code)
	}
}
