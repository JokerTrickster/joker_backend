package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/oauth2"
)

func TestOAuthHandler_Redirect_Google(t *testing.T) {
	t.Log("Redirect: google provider with config -> 302 redirect to Google OAuth URL")
	e := echo.New()
	h := &OAuthHandler{
		FrontendURL: "http://localhost:3000",
		GoogleConfig: &oauth2.Config{
			ClientID:    "test-client-id",
			RedirectURL: "http://localhost/callback",
			Scopes:      []string{"openid", "email", "profile"},
			Endpoint: oauth2.Endpoint{
				AuthURL:  "https://accounts.google.com/o/oauth2/auth",
				TokenURL: "https://oauth2.googleapis.com/token",
			},
		},
	}

	req := httptest.NewRequest(http.MethodGet, "/api/auth/google", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("provider")
	c.SetParamValues("google")

	err := h.Redirect(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusFound, rec.Code)
	location := rec.Header().Get("Location")
	assert.Contains(t, location, "accounts.google.com")
	assert.Contains(t, location, "test-client-id")
	t.Logf("Redirect location: %s", location)
}

func TestOAuthHandler_Redirect_Kakao(t *testing.T) {
	t.Log("Redirect: kakao provider with config -> 302 redirect to Kakao OAuth URL")
	e := echo.New()
	h := &OAuthHandler{
		FrontendURL: "http://localhost:3000",
		KakaoConfig: &oauth2.Config{
			ClientID:    "kakao-client-id",
			RedirectURL: "http://localhost/callback",
			Scopes:      []string{"profile_nickname", "account_email"},
			Endpoint: oauth2.Endpoint{
				AuthURL:  "https://kauth.kakao.com/oauth/authorize",
				TokenURL: "https://kauth.kakao.com/oauth/token",
			},
		},
	}

	req := httptest.NewRequest(http.MethodGet, "/api/auth/kakao", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("provider")
	c.SetParamValues("kakao")

	err := h.Redirect(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusFound, rec.Code)
	location := rec.Header().Get("Location")
	assert.Contains(t, location, "kauth.kakao.com")
	assert.Contains(t, location, "kakao-client-id")
	t.Logf("Redirect location: %s", location)
}

func TestOAuthHandler_Redirect_UnsupportedProvider(t *testing.T) {
	t.Log("Redirect: unsupported provider -> 400 error")
	e := echo.New()
	h := &OAuthHandler{FrontendURL: "http://localhost:3000"}

	req := httptest.NewRequest(http.MethodGet, "/api/auth/facebook", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("provider")
	c.SetParamValues("facebook")

	err := h.Redirect(c)
	require.Error(t, err)
	he, ok := err.(*echo.HTTPError)
	require.True(t, ok)
	assert.Equal(t, http.StatusBadRequest, he.Code)
	t.Logf("Error: code=%d message=%v", he.Code, he.Message)
}

func TestOAuthHandler_Redirect_NotConfigured(t *testing.T) {
	t.Log("Redirect: google provider without config -> 500 error")
	e := echo.New()
	h := &OAuthHandler{
		FrontendURL:  "http://localhost:3000",
		GoogleConfig: nil,
	}

	req := httptest.NewRequest(http.MethodGet, "/api/auth/google", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("provider")
	c.SetParamValues("google")

	err := h.Redirect(c)
	require.Error(t, err)
	he, ok := err.(*echo.HTTPError)
	require.True(t, ok)
	assert.Equal(t, http.StatusInternalServerError, he.Code)
	t.Logf("Error: code=%d message=%v", he.Code, he.Message)
}

func TestOAuthHandler_Callback_NoCode(t *testing.T) {
	t.Log("Callback: no code parameter -> redirect to frontend with error")
	e := echo.New()
	h := &OAuthHandler{FrontendURL: "http://localhost:3000"}

	req := httptest.NewRequest(http.MethodGet, "/api/auth/google/callback", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("provider")
	c.SetParamValues("google")

	err := h.Callback(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusFound, rec.Code)
	location := rec.Header().Get("Location")
	assert.Contains(t, location, "/login?error=no_code")
	t.Logf("Redirect: %s", location)
}

func TestOAuthHandler_Callback_UnsupportedProvider(t *testing.T) {
	t.Log("Callback: unsupported provider -> redirect to frontend with error")
	e := echo.New()
	h := &OAuthHandler{FrontendURL: "http://localhost:3000"}

	req := httptest.NewRequest(http.MethodGet, "/api/auth/facebook/callback?code=test", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("provider")
	c.SetParamValues("facebook")
	c.QueryParams().Set("code", "test")

	err := h.Callback(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusFound, rec.Code)
	location := rec.Header().Get("Location")
	assert.Contains(t, location, "/login?error=unsupported_provider")
	t.Logf("Redirect: %s", location)
}

