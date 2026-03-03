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

type mockSignupAuthUseCase struct {
	signupFunc func(c context.Context, req *request.ReqSignUp) (response.ResSignUp, error)
}

func (m *mockSignupAuthUseCase) Signup(c context.Context, req *request.ReqSignUp) (response.ResSignUp, error) {
	if m.signupFunc != nil {
		return m.signupFunc(c, req)
	}
	return response.ResSignUp{}, nil
}

func setupSignupEcho() *echo.Echo {
	e := echo.New()
	e.Validator = utils.NewValidator()
	return e
}

func TestSignupAuthHandler_Signup(t *testing.T) {
	tests := []struct {
		name           string
		body           []byte
		mockSignup     func(c context.Context, req *request.ReqSignUp) (response.ResSignUp, error)
		wantStatus     int
		wantBodyHas    string
		checkErrReturn bool
	}{
		{
			name: "success: valid request returns 200 with tokens",
			body: mustJSON(t, map[string]interface{}{
				"email":       "test@example.com",
				"password":    "password123",
				"serviceType": "game",
				"name":        "Test User",
			}),
			mockSignup: func(c context.Context, req *request.ReqSignUp) (response.ResSignUp, error) {
				return response.ResSignUp{
					AccessToken:  "access-token-abc",
					RefreshToken: "refresh-token-xyz",
				}, nil
			},
			wantStatus:  http.StatusOK,
			wantBodyHas: "access-token-abc",
		},
		{
			name: "validation error: missing name",
			body: mustJSON(t, map[string]interface{}{
				"email":       "test@example.com",
				"password":    "password123",
				"serviceType": "game",
			}),
			mockSignup: func(c context.Context, req *request.ReqSignUp) (response.ResSignUp, error) {
				return response.ResSignUp{}, nil
			},
			wantStatus:     http.StatusBadRequest,
			checkErrReturn: true,
		},
		{
			name: "validation error: missing email",
			body: mustJSON(t, map[string]interface{}{
				"password":    "password123",
				"serviceType": "game",
				"name":        "Test User",
			}),
			mockSignup: func(c context.Context, req *request.ReqSignUp) (response.ResSignUp, error) {
				return response.ResSignUp{}, nil
			},
			wantStatus:     http.StatusBadRequest,
			checkErrReturn: true,
		},
		{
			name: "usecase error: duplicate email",
			body: mustJSON(t, map[string]interface{}{
				"email":       "existing@example.com",
				"password":    "password123",
				"serviceType": "game",
				"name":        "Test User",
			}),
			mockSignup: func(c context.Context, req *request.ReqSignUp) (response.ResSignUp, error) {
				return response.ResSignUp{}, echo.NewHTTPError(http.StatusConflict, "email already exists")
			},
			wantStatus:     http.StatusConflict,
			checkErrReturn: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Logf("Running case: %s", tt.name)
			e := setupSignupEcho()
			mockUC := &mockSignupAuthUseCase{signupFunc: tt.mockSignup}
			handler := &SignupAuthHandler{UseCase: mockUC}

			req := httptest.NewRequest(http.MethodPost, "/v0.1/auth/signup", bytes.NewReader(tt.body))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			err := handler.Signup(c)
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
