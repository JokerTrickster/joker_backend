package handler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/JokerTrickster/joker_backend/services/molandolanService/features/auth/model/entity"
	"github.com/JokerTrickster/joker_backend/services/molandolanService/features/auth/usecase"
	"github.com/JokerTrickster/joker_backend/shared/utils"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockMeAuthRepository struct {
	findByIDFunc func(ctx context.Context, userID uint) (*entity.MorandoranUser, error)
}

func (m *mockMeAuthRepository) FindUserByEmail(ctx context.Context, email string) (*entity.MorandoranUser, error) {
	return nil, nil
}

func (m *mockMeAuthRepository) FindUserByID(ctx context.Context, userID uint) (*entity.MorandoranUser, error) {
	if m.findByIDFunc != nil {
		return m.findByIDFunc(ctx, userID)
	}
	return nil, nil
}

func (m *mockMeAuthRepository) FindOrCreateByOAuth(ctx context.Context, email, nickname, provider string, profileImage *string) (*entity.MorandoranUser, error) {
	return nil, nil
}

func (m *mockMeAuthRepository) UpdateNickname(ctx context.Context, userID uint, nickname string) (*entity.MorandoranUser, error) {
	return nil, nil
}

func setupMeEcho() *echo.Echo {
	e := echo.New()
	e.Validator = utils.NewValidator()
	return e
}

func TestMeHandler_Me(t *testing.T) {
	tests := []struct {
		name           string
		setUserID      func(c echo.Context)
		mockFindByID   func(ctx context.Context, userID uint) (*entity.MorandoranUser, error)
		wantStatus     int
		wantBodyHas    string
		checkErrReturn bool
	}{
		{
			name: "success: valid userID returns 200 with profileImage and provider",
			setUserID: func(c echo.Context) {
				c.Set("userID", uint(1))
			},
			mockFindByID: func(ctx context.Context, userID uint) (*entity.MorandoranUser, error) {
				pic := "https://example.com/pic.jpg"
				return &entity.MorandoranUser{
					ID:           1,
					Nickname:     "testuser",
					Email:        "user@example.com",
					Role:         "user",
					Provider:     "google",
					ProfileImage: &pic,
				}, nil
			},
			wantStatus:  http.StatusOK,
			wantBodyHas: "profileImage",
		},
		{
			name:           "no userID in context returns 401",
			setUserID:      func(c echo.Context) { /* do not set userID */ },
			mockFindByID:   nil,
			wantStatus:     http.StatusUnauthorized,
			checkErrReturn: true,
		},
		{
			name: "user not found returns 404",
			setUserID: func(c echo.Context) {
				c.Set("userID", uint(999))
			},
			mockFindByID: func(ctx context.Context, userID uint) (*entity.MorandoranUser, error) {
				return nil, assert.AnError
			},
			wantStatus:     http.StatusNotFound,
			checkErrReturn: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Logf("Running case: %s", tt.name)
			e := setupMeEcho()
			mockRepo := &mockMeAuthRepository{}
			if tt.mockFindByID != nil {
				mockRepo.findByIDFunc = tt.mockFindByID
			}
			uc := usecase.NewMeUseCase(mockRepo, 10*time.Second)
			h := NewMeHandler(uc)

			req := httptest.NewRequest(http.MethodGet, "/api/auth/me", nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			if tt.setUserID != nil {
				tt.setUserID(c)
			}

			err := h.Me(c)
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
