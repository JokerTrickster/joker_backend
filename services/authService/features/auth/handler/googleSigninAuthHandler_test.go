package handler

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/JokerTrickster/joker_backend/services/authService/features/auth/model/response"
	"github.com/JokerTrickster/joker_backend/shared/utils"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockGoogleSigninAuthUseCase struct {
	googleSigninFunc func(ctx context.Context, idToken string) (response.ResGoogleSignin, error)
}

func (m *mockGoogleSigninAuthUseCase) GoogleSignin(ctx context.Context, idToken string) (response.ResGoogleSignin, error) {
	if m.googleSigninFunc != nil {
		return m.googleSigninFunc(ctx, idToken)
	}
	return response.ResGoogleSignin{}, nil
}

func setupGoogleSigninEcho() *echo.Echo {
	e := echo.New()
	e.Validator = utils.NewValidator()
	return e
}

func TestGoogleSigninAuthHandler_GoogleSignin(t *testing.T) {
	tests := []struct {
		name             string
		body             []byte
		mockGoogleSignin func(ctx context.Context, idToken string) (response.ResGoogleSignin, error)
		wantStatus       int
		wantBodyHas      string
		checkErrReturn   bool
	}{
		{
			name: "success: valid idToken returns 200",
			body: mustJSON(t, map[string]interface{}{
				"idToken": "valid-google-id-token-xyz",
			}),
			mockGoogleSignin: func(ctx context.Context, idToken string) (response.ResGoogleSignin, error) {
				return response.ResGoogleSignin{
					AccessToken:  "google-access-token",
					RefreshToken: "google-refresh-token",
				}, nil
			},
			wantStatus:  http.StatusOK,
			wantBodyHas: "google-access-token",
		},
		{
			name: "validation error: missing idToken",
			body: mustJSON(t, map[string]interface{}{}),
			mockGoogleSignin: func(ctx context.Context, idToken string) (response.ResGoogleSignin, error) {
				return response.ResGoogleSignin{}, nil
			},
			wantStatus:     http.StatusBadRequest,
			checkErrReturn: true,
		},
		{
			name: "usecase error: invalid token",
			body: mustJSON(t, map[string]interface{}{
				"idToken": "invalid-id-token",
			}),
			mockGoogleSignin: func(ctx context.Context, idToken string) (response.ResGoogleSignin, error) {
				return response.ResGoogleSignin{}, echo.NewHTTPError(http.StatusUnauthorized, "invalid google token")
			},
			wantStatus:     http.StatusUnauthorized,
			checkErrReturn: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Logf("Running case: %s", tt.name)
			e := setupGoogleSigninEcho()
			mockUC := &mockGoogleSigninAuthUseCase{googleSigninFunc: tt.mockGoogleSignin}
			handler := &GoogleSigninAuthHandler{UseCase: mockUC}

			req := httptest.NewRequest(http.MethodPost, "/v0.1/auth/google/signin", bytes.NewReader(tt.body))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			err := handler.GoogleSignin(c)
			if tt.checkErrReturn {
				require.Error(t, err, "expected handler to return error")
				if he, ok := err.(*echo.HTTPError); ok {
					assert.Equal(t, tt.wantStatus, he.Code, "HTTP status mismatch")
					t.Logf("Error response: code=%d message=%v", he.Code, he.Message)
				}
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.wantStatus, rec.Code, "HTTP status mismatch")
			if tt.wantBodyHas != "" {
				assert.Contains(t, rec.Body.String(), tt.wantBodyHas, "Response body should contain expected value")
			}
			t.Logf("Response: status=%d body=%s", rec.Code, rec.Body.String())
		})
	}
}
