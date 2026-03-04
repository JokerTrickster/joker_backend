package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/JokerTrickster/joker_backend/services/morandoranService/features/auth/model/entity"
	"github.com/JokerTrickster/joker_backend/services/morandoranService/features/auth/usecase"
	"github.com/JokerTrickster/joker_backend/shared/jwt"
	"github.com/JokerTrickster/joker_backend/shared/utils"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

type mockAuthRepository struct {
	findByEmailFunc func(ctx context.Context, email string) (*entity.MorandoranUser, error)
}

func (m *mockAuthRepository) FindUserByEmail(ctx context.Context, email string) (*entity.MorandoranUser, error) {
	if m.findByEmailFunc != nil {
		return m.findByEmailFunc(ctx, email)
	}
	return nil, nil
}

func (m *mockAuthRepository) FindUserByID(ctx context.Context, userID uint) (*entity.MorandoranUser, error) {
	return nil, nil
}

func initJWTForTest(t *testing.T) {
	t.Helper()
	os.Setenv("IS_LOCAL", "true")
	err := jwt.InitJwt()
	require.NoError(t, err, "JWT init should succeed for login tests")
}

func setupLoginEcho() *echo.Echo {
	e := echo.New()
	e.Validator = utils.NewValidator()
	return e
}

func mustJSON(t *testing.T, v interface{}) []byte {
	t.Helper()
	b, err := json.Marshal(v)
	require.NoError(t, err)
	return b
}

func TestLoginHandler_Login(t *testing.T) {
	initJWTForTest(t)

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	require.NoError(t, err, "bcrypt hash should succeed")

	tests := []struct {
		name           string
		body           []byte
		mockFindByEmail func(ctx context.Context, email string) (*entity.MorandoranUser, error)
		wantStatus     int
		wantBodyHas    string
		checkErrReturn bool
	}{
		{
			name: "success: valid credentials returns 200 and token",
			body: mustJSON(t, map[string]string{
				"email":    "user@example.com",
				"password": "password123",
			}),
			mockFindByEmail: func(ctx context.Context, email string) (*entity.MorandoranUser, error) {
				return &entity.MorandoranUser{
					ID:       1,
					Nickname: "testuser",
					Email:    "user@example.com",
					Password: string(hashedPassword),
					Role:     "user",
				}, nil
			},
			wantStatus:  http.StatusOK,
			wantBodyHas: "token",
		},
		{
			name: "invalid credentials: user not found returns 401",
			body: mustJSON(t, map[string]string{
				"email":    "nonexistent@example.com",
				"password": "password123",
			}),
			mockFindByEmail: func(ctx context.Context, email string) (*entity.MorandoranUser, error) {
				return nil, assert.AnError
			},
			wantStatus:     http.StatusUnauthorized,
			checkErrReturn: true,
		},
		{
			name: "invalid credentials: wrong password returns 401",
			body: mustJSON(t, map[string]string{
				"email":    "user@example.com",
				"password": "wrongpassword",
			}),
			mockFindByEmail: func(ctx context.Context, email string) (*entity.MorandoranUser, error) {
				return &entity.MorandoranUser{
					ID:       1,
					Nickname: "testuser",
					Email:    "user@example.com",
					Password: string(hashedPassword),
					Role:     "user",
				}, nil
			},
			wantStatus:     http.StatusUnauthorized,
			checkErrReturn: true,
		},
		{
			name: "validation error: missing email returns 400",
			body: mustJSON(t, map[string]string{
				"password": "password123",
			}),
			mockFindByEmail: nil,
			wantStatus:      http.StatusBadRequest,
			checkErrReturn:  true,
		},
		{
			name: "validation error: invalid email format returns 400",
			body: mustJSON(t, map[string]string{
				"email":    "not-an-email",
				"password": "password123",
			}),
			mockFindByEmail: nil,
			wantStatus:      http.StatusBadRequest,
			checkErrReturn:  true,
		},
		{
			name: "validation error: missing password returns 400",
			body: mustJSON(t, map[string]string{
				"email": "user@example.com",
			}),
			mockFindByEmail: nil,
			wantStatus:      http.StatusBadRequest,
			checkErrReturn:  true,
		},
		{
			name:           "bind error: malformed JSON returns 400",
			body:           []byte(`{"email": "user@example.com", "password":`),
			mockFindByEmail: nil,
			wantStatus:      http.StatusBadRequest,
			checkErrReturn: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Logf("Running case: %s", tt.name)
			e := setupLoginEcho()
			mockRepo := &mockAuthRepository{}
			if tt.mockFindByEmail != nil {
				mockRepo.findByEmailFunc = tt.mockFindByEmail
			}
			uc := usecase.NewLoginUseCase(mockRepo, 10*time.Second)
			h := NewLoginHandler(uc)

			req := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewReader(tt.body))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			err := h.Login(c)
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
