package handler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/JokerTrickster/joker_backend/shared/utils"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockLogoutAuthUseCase struct {
	logoutFunc func(c context.Context, userID uint) error
}

func (m *mockLogoutAuthUseCase) Logout(c context.Context, userID uint) error {
	if m.logoutFunc != nil {
		return m.logoutFunc(c, userID)
	}
	return nil
}

func setupLogoutEcho() *echo.Echo {
	e := echo.New()
	e.Validator = utils.NewValidator()
	return e
}

func TestLogoutAuthHandler_Logout(t *testing.T) {
	tests := []struct {
		name           string
		setupContext   func(c echo.Context)
		mockLogout     func(c context.Context, userID uint) error
		wantStatus     int
		checkErrReturn bool
	}{
		{
			name: "success: userID in context returns 200",
			setupContext: func(c echo.Context) {
				c.Set("userID", uint(42))
			},
			mockLogout: func(c context.Context, userID uint) error {
				return nil
			},
			wantStatus: http.StatusOK,
		},
		{
			name: "no userID in context returns 401",
			setupContext: func(c echo.Context) {
				// do not set userID
			},
			mockLogout:     nil,
			wantStatus:     http.StatusUnauthorized,
			checkErrReturn: true,
		},
		{
			name: "invalid userID type returns 401",
			setupContext: func(c echo.Context) {
				c.Set("userID", "not-a-uint")
			},
			mockLogout:     nil,
			wantStatus:     http.StatusUnauthorized,
			checkErrReturn: true,
		},
		{
			name: "usecase error returns error",
			setupContext: func(c echo.Context) {
				c.Set("userID", uint(99))
			},
			mockLogout: func(c context.Context, userID uint) error {
				return echo.NewHTTPError(http.StatusInternalServerError, "logout failed")
			},
			wantStatus:     http.StatusInternalServerError,
			checkErrReturn: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Logf("Running case: %s", tt.name)
			e := setupLogoutEcho()
			mockUC := &mockLogoutAuthUseCase{logoutFunc: tt.mockLogout}
			handler := &LogoutAuthHandler{UseCase: mockUC}

			req := httptest.NewRequest(http.MethodPost, "/v0.1/auth/logout", nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			if tt.setupContext != nil {
				tt.setupContext(c)
			}

			err := handler.Logout(c)
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
			t.Logf("Response: status=%d body=%s", rec.Code, rec.Body.String())
		})
	}
}
