package handler

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/JokerTrickster/joker_backend/services/authService/features/auth/model/request"
	"github.com/JokerTrickster/joker_backend/services/authService/features/auth/model/response"
	"github.com/JokerTrickster/joker_backend/shared/utils"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockRefreshTokenUseCase struct {
	refreshFunc func(c context.Context, req *request.ReqRefreshToken) (response.ResRefreshToken, error)
}

func (m *mockRefreshTokenUseCase) RefreshToken(c context.Context, req *request.ReqRefreshToken) (response.ResRefreshToken, error) {
	if m.refreshFunc != nil {
		return m.refreshFunc(c, req)
	}
	return response.ResRefreshToken{}, nil
}

func setupRefreshTokenEcho() *echo.Echo {
	e := echo.New()
	e.Validator = utils.NewValidator()
	return e
}

func TestRefreshTokenHandler_RefreshToken(t *testing.T) {
	tests := []struct {
		name           string
		body           []byte
		mockRefresh    func(c context.Context, req *request.ReqRefreshToken) (response.ResRefreshToken, error)
		wantStatus     int
		wantBodyHas    string
		checkErrReturn bool
	}{
		{
			name: "success: valid refresh token returns 200",
			body: mustJSON(t, map[string]interface{}{
				"refreshToken": "valid-refresh-token-xyz",
			}),
			mockRefresh: func(c context.Context, req *request.ReqRefreshToken) (response.ResRefreshToken, error) {
				return response.ResRefreshToken{
					AccessToken:  "new-access-token",
					RefreshToken: "new-refresh-token",
				}, nil
			},
			wantStatus:  http.StatusOK,
			wantBodyHas: "new-access-token",
		},
		{
			name: "validation error: missing token",
			body: mustJSON(t, map[string]interface{}{}),
			mockRefresh: func(c context.Context, req *request.ReqRefreshToken) (response.ResRefreshToken, error) {
				return response.ResRefreshToken{}, nil
			},
			wantStatus:     http.StatusBadRequest,
			checkErrReturn: true,
		},
		{
			name: "usecase error: invalid token",
			body: mustJSON(t, map[string]interface{}{
				"refreshToken": "invalid-token",
			}),
			mockRefresh: func(c context.Context, req *request.ReqRefreshToken) (response.ResRefreshToken, error) {
				return response.ResRefreshToken{}, echo.NewHTTPError(http.StatusUnauthorized, "token invalid or expired")
			},
			wantStatus:     http.StatusUnauthorized,
			checkErrReturn: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Logf("Running case: %s", tt.name)
			e := setupRefreshTokenEcho()
			mockUC := &mockRefreshTokenUseCase{refreshFunc: tt.mockRefresh}
			handler := &RefreshTokenHandler{UseCase: mockUC}

			req := httptest.NewRequest(http.MethodPost, "/v0.1/auth/refresh", bytes.NewReader(tt.body))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			err := handler.RefreshToken(c)
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
