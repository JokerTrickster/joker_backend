package handler

import (
	"bytes"
	"context"
	"encoding/json"
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

type mockSigninAuthUseCase struct {
	signinFunc func(c context.Context, req *request.ReqSignIn) (response.ResSignIn, error)
}

func (m *mockSigninAuthUseCase) Signin(c context.Context, req *request.ReqSignIn) (response.ResSignIn, error) {
	if m.signinFunc != nil {
		return m.signinFunc(c, req)
	}
	return response.ResSignIn{}, nil
}

func setupSigninEcho() *echo.Echo {
	e := echo.New()
	e.Validator = utils.NewValidator()
	return e
}

func TestSigninAuthHandler_Signin(t *testing.T) {
	tests := []struct {
		name           string
		body           []byte
		mockSignin     func(c context.Context, req *request.ReqSignIn) (response.ResSignIn, error)
		wantStatus     int
		wantBodyHas    string
		checkErrReturn bool
	}{
		{
			name: "success: valid JSON body, mock usecase returns tokens",
			body: mustJSON(t, map[string]interface{}{
				"email":       "test@example.com",
				"password":    "password123",
				"serviceType": "game",
			}),
			mockSignin: func(c context.Context, req *request.ReqSignIn) (response.ResSignIn, error) {
				return response.ResSignIn{
					AccessToken:  "access-token-123",
					RefreshToken: "refresh-token-456",
				}, nil
			},
			wantStatus:  http.StatusOK,
			wantBodyHas: "access-token-123",
		},
		{
			name: "invalid JSON: malformed body",
			body: []byte(`{"email": "test@example.com", "password": "short"`),
			mockSignin: func(c context.Context, req *request.ReqSignIn) (response.ResSignIn, error) {
				return response.ResSignIn{}, nil
			},
			wantStatus:     http.StatusBadRequest,
			checkErrReturn: true,
		},
		{
			name: "validation error: missing email",
			body: mustJSON(t, map[string]interface{}{
				"password": "password123",
			}),
			mockSignin: func(c context.Context, req *request.ReqSignIn) (response.ResSignIn, error) {
				return response.ResSignIn{}, nil
			},
			wantStatus:     http.StatusBadRequest,
			checkErrReturn: true,
		},
		{
			name: "validation error: invalid email format",
			body: mustJSON(t, map[string]interface{}{
				"email":    "not-an-email",
				"password": "password123",
			}),
			mockSignin: func(c context.Context, req *request.ReqSignIn) (response.ResSignIn, error) {
				return response.ResSignIn{}, nil
			},
			wantStatus:     http.StatusBadRequest,
			checkErrReturn: true,
		},
		{
			name: "validation error: password too short",
			body: mustJSON(t, map[string]interface{}{
				"email":    "test@example.com",
				"password": "12345",
			}),
			mockSignin: func(c context.Context, req *request.ReqSignIn) (response.ResSignIn, error) {
				return response.ResSignIn{}, nil
			},
			wantStatus:     http.StatusBadRequest,
			checkErrReturn: true,
		},
		{
			name: "usecase error: returns error",
			body: mustJSON(t, map[string]interface{}{
				"email":    "test@example.com",
				"password": "password123",
			}),
			mockSignin: func(c context.Context, req *request.ReqSignIn) (response.ResSignIn, error) {
				return response.ResSignIn{}, echo.NewHTTPError(http.StatusUnauthorized, "user not found")
			},
			wantStatus:     http.StatusUnauthorized,
			checkErrReturn: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Logf("Running case: %s", tt.name)
			e := setupSigninEcho()
			mockUC := &mockSigninAuthUseCase{signinFunc: tt.mockSignin}
			handler := &SigninAuthHandler{UseCase: mockUC}

			req := httptest.NewRequest(http.MethodPost, "/v0.1/auth/signin", bytes.NewReader(tt.body))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			err := handler.Signin(c)
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

func mustJSON(t *testing.T, v interface{}) []byte {
	t.Helper()
	b, err := json.Marshal(v)
	require.NoError(t, err)
	return b
}
