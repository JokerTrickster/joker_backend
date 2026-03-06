package handler

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/JokerTrickster/joker_backend/services/molandolanService/features/auth/model/entity"
	_interface "github.com/JokerTrickster/joker_backend/services/molandolanService/features/auth/model/interface"
	"github.com/JokerTrickster/joker_backend/services/molandolanService/features/auth/usecase"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
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

func TestOAuthHandler_Callback_InvalidState_NoCookie(t *testing.T) {
	t.Log("Callback: valid code but no oauth_state cookie -> redirect invalid_state")
	e := echo.New()
	h := &OAuthHandler{FrontendURL: "http://localhost:3000"}

	req := httptest.NewRequest(http.MethodGet, "/api/auth/google/callback?code=abc&state=xyz", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("provider")
	c.SetParamValues("google")

	err := h.Callback(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusFound, rec.Code)
	location := rec.Header().Get("Location")
	assert.Contains(t, location, "/login?error=invalid_state")
	t.Logf("Redirect: %s", location)
}

func TestOAuthHandler_Callback_InvalidState_EmptyCookie(t *testing.T) {
	t.Log("Callback: empty oauth_state cookie value -> redirect invalid_state")
	e := echo.New()
	h := &OAuthHandler{FrontendURL: "http://localhost:3000"}

	req := httptest.NewRequest(http.MethodGet, "/api/auth/google/callback?code=abc&state=xyz", nil)
	req.AddCookie(&http.Cookie{Name: "oauth_state", Value: ""})
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("provider")
	c.SetParamValues("google")

	err := h.Callback(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusFound, rec.Code)
	location := rec.Header().Get("Location")
	assert.Contains(t, location, "/login?error=invalid_state")
	t.Logf("Redirect: %s", location)
}

func TestOAuthHandler_Callback_InvalidState_Mismatch(t *testing.T) {
	t.Log("Callback: cookie state != query state -> redirect invalid_state")
	e := echo.New()
	h := &OAuthHandler{FrontendURL: "http://localhost:3000"}

	req := httptest.NewRequest(http.MethodGet, "/api/auth/google/callback?code=abc&state=real_state", nil)
	req.AddCookie(&http.Cookie{Name: "oauth_state", Value: "tampered_state"})
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("provider")
	c.SetParamValues("google")

	err := h.Callback(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusFound, rec.Code)
	location := rec.Header().Get("Location")
	assert.Contains(t, location, "/login?error=invalid_state")
	t.Logf("Redirect: %s", location)
}

func TestOAuthHandler_Redirect_KakaoNotConfigured(t *testing.T) {
	t.Log("Redirect: kakao provider without config -> 500 error")
	e := echo.New()
	h := &OAuthHandler{
		FrontendURL:  "http://localhost:3000",
		KakaoConfig:  nil,
		GoogleConfig: nil,
	}

	req := httptest.NewRequest(http.MethodGet, "/api/auth/kakao", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("provider")
	c.SetParamValues("kakao")

	err := h.Redirect(c)
	require.Error(t, err)
	he, ok := err.(*echo.HTTPError)
	require.True(t, ok)
	assert.Equal(t, http.StatusInternalServerError, he.Code)
	t.Logf("Error: code=%d message=%v", he.Code, he.Message)
}

// mockOAuthTransport returns valid token and userinfo JSON for OAuth flow tests.
type mockOAuthTransport struct {
	tokenResp       []byte
	userinfoResp    []byte
	kakaoUserinfoResp []byte
}

func (m *mockOAuthTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	var body []byte
	if strings.Contains(r.URL.Path, "userinfo") {
		body = m.userinfoResp
	} else if strings.Contains(r.URL.Path, "user/me") && len(m.kakaoUserinfoResp) > 0 {
		body = m.kakaoUserinfoResp
	} else {
		body = m.tokenResp
	}
	return &http.Response{
		StatusCode: http.StatusOK,
		Header:     map[string][]string{"Content-Type": {"application/json"}},
		Body:       io.NopCloser(bytes.NewReader(body)),
	}, nil
}

type mockOAuthCallbackRepo struct {
	findOrCreateFunc func(ctx context.Context, email, nickname, provider string, profileImage *string) (*entity.MorandoranUser, error)
}

func (m *mockOAuthCallbackRepo) FindUserByEmail(ctx context.Context, email string) (*entity.MorandoranUser, error) {
	return nil, nil
}
func (m *mockOAuthCallbackRepo) FindUserByID(ctx context.Context, userID uint) (*entity.MorandoranUser, error) {
	return nil, nil
}
func (m *mockOAuthCallbackRepo) FindOrCreateByOAuth(ctx context.Context, email, nickname, provider string, profileImage *string) (*entity.MorandoranUser, error) {
	if m.findOrCreateFunc != nil {
		return m.findOrCreateFunc(ctx, email, nickname, provider, profileImage)
	}
	return nil, nil
}
func (m *mockOAuthCallbackRepo) UpdateNickname(ctx context.Context, userID uint, nickname string) (*entity.MorandoranUser, error) {
	return nil, nil
}

var _ _interface.IAuthRepository = (*mockOAuthCallbackRepo)(nil)

func TestNewOAuthHandler(t *testing.T) {
	t.Log("NewOAuthHandler: sets FrontendURL and configs from env")
	os.Setenv("FRONTEND_URL", "https://app.example.com")
	os.Setenv("GOOGLE_CLIENT_ID", "google-id")
	os.Setenv("GOOGLE_CLIENT_SECRET", "google-secret")
	defer func() {
		os.Unsetenv("FRONTEND_URL")
		os.Unsetenv("GOOGLE_CLIENT_ID")
		os.Unsetenv("GOOGLE_CLIENT_SECRET")
	}()

	mockRepo := &mockOAuthCallbackRepo{}
	uc := usecase.NewOAuthUseCase(mockRepo, 5*time.Second)
	h := NewOAuthHandler(uc)

	assert.Equal(t, "https://app.example.com", h.FrontendURL)
	require.NotNil(t, h.GoogleConfig)
	assert.Equal(t, "google-id", h.GoogleConfig.ClientID)
	assert.Equal(t, "google-secret", h.GoogleConfig.ClientSecret)
	t.Logf("FrontendURL=%s GoogleConfig.ClientID=%s", h.FrontendURL, h.GoogleConfig.ClientID)
}

func TestNewOAuthHandler_DefaultsFrontendURL(t *testing.T) {
	t.Log("NewOAuthHandler: defaults FrontendURL to localhost:3000 when FRONTEND_URL empty")
	orig := os.Getenv("FRONTEND_URL")
	os.Unsetenv("FRONTEND_URL")
	defer os.Setenv("FRONTEND_URL", orig)

	mockRepo := &mockOAuthCallbackRepo{}
	uc := usecase.NewOAuthUseCase(mockRepo, 5*time.Second)
	h := NewOAuthHandler(uc)

	assert.Equal(t, "http://localhost:3000", h.FrontendURL)
	t.Logf("FrontendURL=%s (default)", h.FrontendURL)
}

func TestOAuthHandler_Callback_UseCaseFails(t *testing.T) {
	t.Log("Callback: valid OAuth flow but UseCase.HandleCallback fails -> redirect server_error")
	tokenResp := []byte(`{"access_token":"mock_token","token_type":"Bearer","expires_in":3600}`)
	userinfoResp := []byte(`{"email":"test@example.com","name":"Test User","picture":"https://pic.example.com/img.jpg"}`)

	transport := &mockOAuthTransport{tokenResp: tokenResp, userinfoResp: userinfoResp}
	httpClient := &http.Client{Transport: transport}

	mockRepo := &mockOAuthCallbackRepo{
		findOrCreateFunc: func(ctx context.Context, email, nickname, provider string, profileImage *string) (*entity.MorandoranUser, error) {
			return nil, fmt.Errorf("db connection failed")
		},
	}
	uc := usecase.NewOAuthUseCase(mockRepo, 10*time.Second)

	state := "valid_state_12345"
	h := &OAuthHandler{
		UseCase:     uc,
		FrontendURL: "http://localhost:3000",
		GoogleConfig: &oauth2.Config{
			ClientID:     "test-client",
			ClientSecret: "secret",
			RedirectURL:  "http://localhost/callback",
			Scopes:       []string{"openid", "email", "profile"},
			Endpoint:     google.Endpoint,
		},
	}

	req := httptest.NewRequest(http.MethodGet, "/api/auth/google/callback?code=abc&state="+state, nil)
	req.AddCookie(&http.Cookie{Name: "oauth_state", Value: state})
	ctx := context.WithValue(req.Context(), oauth2.HTTPClient, httpClient)
	req = req.WithContext(ctx)

	e := echo.New()
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("provider")
	c.SetParamValues("google")

	err := h.Callback(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusFound, rec.Code)
	location := rec.Header().Get("Location")
	assert.Contains(t, location, "/login?error=server_error")
	t.Logf("Redirect: %s", location)
}

func TestOAuthHandler_Callback_UseCaseFails_Kakao(t *testing.T) {
	t.Log("Callback: Kakao OAuth flow, UseCase fails -> redirect server_error")
	tokenResp := []byte(`{"access_token":"mock_token","token_type":"Bearer","expires_in":3600}`)
	kakaoUserinfoResp := []byte(`{"kakao_account":{"email":"kakao@example.com","profile":{"nickname":"KakaoUser","profile_image_url":"https://pic.kakao.com/img.jpg"}}}`)

	transport := &mockOAuthTransport{
		tokenResp:         tokenResp,
		userinfoResp:      []byte(`{}`),
		kakaoUserinfoResp: kakaoUserinfoResp,
	}
	httpClient := &http.Client{Transport: transport}

	mockRepo := &mockOAuthCallbackRepo{
		findOrCreateFunc: func(ctx context.Context, email, nickname, provider string, profileImage *string) (*entity.MorandoranUser, error) {
			return nil, fmt.Errorf("db error")
		},
	}
	uc := usecase.NewOAuthUseCase(mockRepo, 10*time.Second)

	state := "kakao_state_xyz"
	h := &OAuthHandler{
		UseCase:     uc,
		FrontendURL: "http://localhost:3000",
		KakaoConfig: &oauth2.Config{
			ClientID:     "kakao-client",
			ClientSecret: "secret",
			RedirectURL:  "http://localhost/callback",
			Scopes:       []string{"profile_nickname", "account_email"},
			Endpoint: oauth2.Endpoint{
				AuthURL:  "https://kauth.kakao.com/oauth/authorize",
				TokenURL: "https://kauth.kakao.com/oauth/token",
			},
		},
	}

	req := httptest.NewRequest(http.MethodGet, "/api/auth/kakao/callback?code=xyz&state="+state, nil)
	req.AddCookie(&http.Cookie{Name: "oauth_state", Value: state})
	ctx := context.WithValue(req.Context(), oauth2.HTTPClient, httpClient)
	req = req.WithContext(ctx)

	e := echo.New()
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("provider")
	c.SetParamValues("kakao")

	err := h.Callback(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusFound, rec.Code)
	location := rec.Header().Get("Location")
	assert.Contains(t, location, "/login?error=server_error")
	t.Logf("Redirect: %s", location)
}

