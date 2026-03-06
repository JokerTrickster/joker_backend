package handler

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/JokerTrickster/joker_backend/services/molandolanService/features/auth/usecase"
	"github.com/JokerTrickster/joker_backend/shared/logger"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

type OAuthHandler struct {
	UseCase      *usecase.OAuthUseCase
	GoogleConfig *oauth2.Config
	KakaoConfig  *oauth2.Config
	FrontendURL  string
}

func NewOAuthHandler(uc *usecase.OAuthUseCase) *OAuthHandler {
	frontendURL := os.Getenv("FRONTEND_URL")
	if frontendURL == "" {
		frontendURL = "http://localhost:3000"
	}

	h := &OAuthHandler{
		UseCase:     uc,
		FrontendURL: frontendURL,
	}

	if clientID := os.Getenv("GOOGLE_CLIENT_ID"); clientID != "" {
		h.GoogleConfig = &oauth2.Config{
			ClientID:     clientID,
			ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
			RedirectURL:  os.Getenv("GOOGLE_REDIRECT_URL"),
			Scopes:       []string{"openid", "email", "profile"},
			Endpoint:     google.Endpoint,
		}
		if h.GoogleConfig.RedirectURL == "" {
			h.GoogleConfig.RedirectURL = fmt.Sprintf("%s/api/auth/google/callback", os.Getenv("API_BASE_URL"))
		}
	}

	if clientID := os.Getenv("KAKAO_CLIENT_ID"); clientID != "" {
		h.KakaoConfig = &oauth2.Config{
			ClientID:     clientID,
			ClientSecret: os.Getenv("KAKAO_CLIENT_SECRET"),
			RedirectURL:  os.Getenv("KAKAO_REDIRECT_URL"),
			Scopes:       []string{"profile_nickname", "account_email", "profile_image"},
			Endpoint: oauth2.Endpoint{
				AuthURL:  "https://kauth.kakao.com/oauth/authorize",
				TokenURL: "https://kauth.kakao.com/oauth/token",
			},
		}
		if h.KakaoConfig.RedirectURL == "" {
			h.KakaoConfig.RedirectURL = fmt.Sprintf("%s/api/auth/kakao/callback", os.Getenv("API_BASE_URL"))
		}
	}

	return h
}

func (h *OAuthHandler) Redirect(c echo.Context) error {
	provider := c.Param("provider")
	var config *oauth2.Config

	switch provider {
	case "google":
		config = h.GoogleConfig
	case "kakao":
		config = h.KakaoConfig
	default:
		return echo.NewHTTPError(http.StatusBadRequest, "unsupported provider")
	}

	if config == nil {
		return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("%s OAuth not configured", provider))
	}

	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to generate state")
	}
	state := base64.URLEncoding.EncodeToString(b)

	c.SetCookie(&http.Cookie{
		Name:     "oauth_state",
		Value:    state,
		Path:     "/",
		MaxAge:   int(10 * time.Minute / time.Second),
		HttpOnly: true,
		Secure:   os.Getenv("ENV") != "local",
		SameSite: http.SameSiteLaxMode,
	})

	url := config.AuthCodeURL(state, oauth2.AccessTypeOffline)
	return c.Redirect(http.StatusFound, url)
}

func (h *OAuthHandler) Callback(c echo.Context) error {
	provider := c.Param("provider")
	code := c.QueryParam("code")
	if code == "" {
		return c.Redirect(http.StatusFound, h.FrontendURL+"/login?error=no_code")
	}

	// Check unsupported provider early—no state validation needed for unsupported providers
	switch provider {
	case "google", "kakao":
		// fall through to state validation
	default:
		return c.Redirect(http.StatusFound, h.FrontendURL+"/login?error=unsupported_provider")
	}

	stateParam := c.QueryParam("state")
	cookie, err := c.Cookie("oauth_state")
	if err != nil || cookie.Value == "" || cookie.Value != stateParam {
		return c.Redirect(http.StatusFound, h.FrontendURL+"/login?error=invalid_state")
	}
	c.SetCookie(&http.Cookie{
		Name:   "oauth_state",
		Value:  "",
		Path:   "/",
		MaxAge: -1,
	})

	ctx := c.Request().Context()
	var email, nickname string
	var profileImage *string

	switch provider {
	case "google":
		e, n, p, err := h.googleUserInfo(ctx, code)
		if err != nil {
			logger.Error("Google OAuth callback failed", zap.Error(err))
			return c.Redirect(http.StatusFound, h.FrontendURL+"/login?error=oauth_failed")
		}
		email, nickname, profileImage = e, n, p
	case "kakao":
		e, n, p, err := h.kakaoUserInfo(ctx, code)
		if err != nil {
			logger.Error("Kakao OAuth callback failed", zap.Error(err))
			return c.Redirect(http.StatusFound, h.FrontendURL+"/login?error=oauth_failed")
		}
		email, nickname, profileImage = e, n, p
	}

	res, err := h.UseCase.HandleCallback(ctx, email, nickname, provider, profileImage)
	if err != nil {
		logger.Error("OAuth user creation failed", zap.Error(err))
		return c.Redirect(http.StatusFound, h.FrontendURL+"/login?error=server_error")
	}

	return c.Redirect(http.StatusFound, fmt.Sprintf("%s/login/callback?token=%s", h.FrontendURL, res.Token))
}

func (h *OAuthHandler) googleUserInfo(ctx context.Context, code string) (string, string, *string, error) {
	token, err := h.GoogleConfig.Exchange(ctx, code)
	if err != nil {
		return "", "", nil, fmt.Errorf("exchange failed: %w", err)
	}

	client := h.GoogleConfig.Client(ctx, token)
	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		return "", "", nil, fmt.Errorf("userinfo request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", "", nil, fmt.Errorf("read userinfo body failed: %w", err)
	}
	var info struct {
		Email   string `json:"email"`
		Name    string `json:"name"`
		Picture string `json:"picture"`
	}
	if err := json.Unmarshal(body, &info); err != nil {
		return "", "", nil, fmt.Errorf("parse userinfo failed: %w", err)
	}

	var pic *string
	if info.Picture != "" {
		pic = &info.Picture
	}
	return info.Email, info.Name, pic, nil
}

func (h *OAuthHandler) kakaoUserInfo(ctx context.Context, code string) (string, string, *string, error) {
	token, err := h.KakaoConfig.Exchange(ctx, code)
	if err != nil {
		return "", "", nil, fmt.Errorf("exchange failed: %w", err)
	}

	client := h.KakaoConfig.Client(ctx, token)
	resp, err := client.Get("https://kapi.kakao.com/v2/user/me")
	if err != nil {
		return "", "", nil, fmt.Errorf("user info request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", "", nil, fmt.Errorf("read user info body failed: %w", err)
	}
	var info struct {
		KakaoAccount struct {
			Email   string `json:"email"`
			Profile struct {
				Nickname string `json:"nickname"`
				Image    string `json:"profile_image_url"`
			} `json:"profile"`
		} `json:"kakao_account"`
	}
	if err := json.Unmarshal(body, &info); err != nil {
		return "", "", nil, fmt.Errorf("parse user info failed: %w", err)
	}

	var pic *string
	if info.KakaoAccount.Profile.Image != "" {
		pic = &info.KakaoAccount.Profile.Image
	}

	email := info.KakaoAccount.Email
	nickname := info.KakaoAccount.Profile.Nickname
	if nickname == "" {
		nickname = "사용자"
	}

	return email, nickname, pic, nil
}
