package handler

import (
	"bytes"
	"context"
	"errors"
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

type mockCheckEmailAuthUseCase struct {
	checkEmailFunc func(ctx context.Context, req *request.ReqCheckEmail) (*response.ResCheckEmail, error)
}

func (m *mockCheckEmailAuthUseCase) CheckEmail(ctx context.Context, req *request.ReqCheckEmail) (*response.ResCheckEmail, error) {
	if m.checkEmailFunc != nil {
		return m.checkEmailFunc(ctx, req)
	}
	return nil, nil
}

func setupCheckEmailEcho() *echo.Echo {
	e := echo.New()
	e.Validator = utils.NewValidator()
	return e
}

func TestCheckEmailAuthHandler_CheckEmail(t *testing.T) {
	tests := []struct {
		name           string
		body           []byte
		mockCheckEmail func(ctx context.Context, req *request.ReqCheckEmail) (*response.ResCheckEmail, error)
		wantStatus     int
		wantBodyHas    string
		checkErrReturn bool
	}{
		{
			name: "success: email exists",
			body: mustJSON(t, map[string]interface{}{
				"email":    "existing@example.com",
				"provider": "email",
			}),
			mockCheckEmail: func(ctx context.Context, req *request.ReqCheckEmail) (*response.ResCheckEmail, error) {
				return &response.ResCheckEmail{
					Email:     "existing@example.com",
					Exists:    true,
					Available: false,
				}, nil
			},
			wantStatus:  http.StatusOK,
			wantBodyHas: "existing@example.com",
		},
		{
			name: "success: email available",
			body: mustJSON(t, map[string]interface{}{
				"email":    "new@example.com",
				"provider": "email",
			}),
			mockCheckEmail: func(ctx context.Context, req *request.ReqCheckEmail) (*response.ResCheckEmail, error) {
				return &response.ResCheckEmail{
					Email:     "new@example.com",
					Exists:    false,
					Available: true,
				}, nil
			},
			wantStatus:  http.StatusOK,
			wantBodyHas: "new@example.com",
		},
		{
			name: "invalid JSON returns 400",
			body: []byte(`{"email": "test@example.com", "provider": "email"`),
			mockCheckEmail: func(ctx context.Context, req *request.ReqCheckEmail) (*response.ResCheckEmail, error) {
				return nil, nil
			},
			wantStatus:     http.StatusBadRequest,
			checkErrReturn: true,
		},
		{
			name: "validation error: invalid provider",
			body: mustJSON(t, map[string]interface{}{
				"email":    "test@example.com",
				"provider": "invalid_provider",
			}),
			mockCheckEmail: func(ctx context.Context, req *request.ReqCheckEmail) (*response.ResCheckEmail, error) {
				return nil, nil
			},
			wantStatus:     http.StatusBadRequest,
			checkErrReturn: true,
		},
		{
			name: "usecase error returns 500",
			body: mustJSON(t, map[string]interface{}{
				"email":    "test@example.com",
				"provider": "email",
			}),
			mockCheckEmail: func(ctx context.Context, req *request.ReqCheckEmail) (*response.ResCheckEmail, error) {
				return nil, errors.New("usecase internal error")
			},
			wantStatus:     http.StatusInternalServerError,
			checkErrReturn: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Logf("Running case: %s", tt.name)
			e := setupCheckEmailEcho()
			mockUC := &mockCheckEmailAuthUseCase{checkEmailFunc: tt.mockCheckEmail}
			handler := &CheckEmailAuthHandler{UseCase: mockUC}

			req := httptest.NewRequest(http.MethodPost, "/v0.1/auth/check-email", bytes.NewReader(tt.body))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			err := handler.CheckEmail(c)
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
