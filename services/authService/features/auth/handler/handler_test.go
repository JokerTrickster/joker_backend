package handler

import (
	"testing"

	"github.com/JokerTrickster/joker_backend/shared/utils"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupHandlerTestEcho() *echo.Echo {
	e := echo.New()
	e.Validator = utils.NewValidator()
	return e
}

// NewAuthHandler uses real DB; we test individual handler constructors with mocks
// to verify route registration without DB dependency.

func TestNewSigninAuthHandler_DoesNotPanic(t *testing.T) {
	e := setupHandlerTestEcho()
	mock := &mockSigninAuthUseCase{}
	require.NotPanics(t, func() {
		NewSigninAuthHandler(e, mock)
	})
	routes := e.Routes()
	assert.NotEmpty(t, routes, "Signin route should be registered")
	t.Logf("Registered %d routes after NewSigninAuthHandler", len(routes))
}

func TestNewSignupAuthHandler_DoesNotPanic(t *testing.T) {
	e := setupHandlerTestEcho()
	mock := &mockSignupAuthUseCase{}
	require.NotPanics(t, func() {
		NewSignupAuthHandler(e, mock)
	})
	routes := e.Routes()
	assert.NotEmpty(t, routes, "Signup route should be registered")
	t.Logf("Registered %d routes after NewSignupAuthHandler", len(routes))
}

func TestNewRefreshTokenHandler_DoesNotPanic(t *testing.T) {
	e := setupHandlerTestEcho()
	mock := &mockRefreshTokenUseCase{}
	require.NotPanics(t, func() {
		NewRefreshTokenHandler(e, mock)
	})
	routes := e.Routes()
	assert.NotEmpty(t, routes, "RefreshToken route should be registered")
	t.Logf("Registered %d routes after NewRefreshTokenHandler", len(routes))
}

func TestNewLogoutAuthHandler_DoesNotPanic(t *testing.T) {
	e := setupHandlerTestEcho()
	mock := &mockLogoutAuthUseCase{}
	require.NotPanics(t, func() {
		NewLogoutAuthHandler(e, mock)
	})
	routes := e.Routes()
	assert.NotEmpty(t, routes, "Logout route should be registered")
	t.Logf("Registered %d routes after NewLogoutAuthHandler", len(routes))
}

func TestNewCheckEmailAuthHandler_DoesNotPanic(t *testing.T) {
	e := setupHandlerTestEcho()
	mock := &mockCheckEmailAuthUseCase{}
	require.NotPanics(t, func() {
		NewCheckEmailAuthHandler(e, mock)
	})
	routes := e.Routes()
	assert.NotEmpty(t, routes, "CheckEmail route should be registered")
	t.Logf("Registered %d routes after NewCheckEmailAuthHandler", len(routes))
}

func TestNewGoogleSigninAuthHandler_DoesNotPanic(t *testing.T) {
	e := setupHandlerTestEcho()
	mock := &mockGoogleSigninAuthUseCase{}
	require.NotPanics(t, func() {
		NewGoogleSigninAuthHandler(e, mock)
	})
	routes := e.Routes()
	assert.NotEmpty(t, routes, "GoogleSignin route should be registered")
	t.Logf("Registered %d routes after NewGoogleSigninAuthHandler", len(routes))
}
