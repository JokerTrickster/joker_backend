package handler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/JokerTrickster/joker_backend/services/morandoranService/features/gallery/model/entity"
	_interface "github.com/JokerTrickster/joker_backend/services/morandoranService/features/gallery/model/interface"
	"github.com/JokerTrickster/joker_backend/services/morandoranService/features/gallery/usecase"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockGalleryLikeRepository struct {
	toggleLikeFunc func(ctx context.Context, userID, galleryID uint) (bool, int, error)
}

func (m *mockGalleryLikeRepository) List(ctx context.Context, page, limit int) ([]entity.GalleryPost, int64, error) {
	return nil, 0, nil
}
func (m *mockGalleryLikeRepository) FindByID(ctx context.Context, id uint) (*entity.GalleryPost, error) {
	return nil, nil
}
func (m *mockGalleryLikeRepository) Create(ctx context.Context, post *entity.GalleryPost) (*entity.GalleryPost, error) {
	return nil, nil
}
func (m *mockGalleryLikeRepository) Delete(ctx context.Context, id uint) error {
	return nil
}
func (m *mockGalleryLikeRepository) IsLiked(ctx context.Context, userID, galleryID uint) (bool, error) {
	return false, nil
}
func (m *mockGalleryLikeRepository) ToggleLike(ctx context.Context, userID, galleryID uint) (bool, int, error) {
	if m.toggleLikeFunc != nil {
		return m.toggleLikeFunc(ctx, userID, galleryID)
	}
	return false, 0, nil
}
func (m *mockGalleryLikeRepository) ListComments(ctx context.Context, galleryID uint, page, limit int) ([]entity.GalleryComment, int64, error) {
	return nil, 0, nil
}
func (m *mockGalleryLikeRepository) FindCommentByID(ctx context.Context, id uint) (*entity.GalleryComment, error) {
	return nil, nil
}
func (m *mockGalleryLikeRepository) CreateComment(ctx context.Context, comment *entity.GalleryComment) (*entity.GalleryComment, error) {
	return nil, nil
}
func (m *mockGalleryLikeRepository) DeleteComment(ctx context.Context, id uint) error {
	return nil
}
func (m *mockGalleryLikeRepository) GetAuthorNickname(ctx context.Context, userID uint) (string, error) {
	return "", nil
}

var _ _interface.IGalleryRepository = (*mockGalleryLikeRepository)(nil)

func TestLikeHandler_Like(t *testing.T) {
	tests := []struct {
		name           string
		paramID        string
		setUserID      bool
		mockToggleLike func(ctx context.Context, userID, galleryID uint) (bool, int, error)
		wantStatus     int
		wantBodyHas    string
		checkErrReturn bool
	}{
		{
			name:    "success: toggle like returns 200 with isLiked and likeCount",
			paramID: "1",
			setUserID: true,
			mockToggleLike: func(ctx context.Context, userID, galleryID uint) (bool, int, error) {
				return true, 6, nil
			},
			wantStatus:  http.StatusOK,
			wantBodyHas: "isLiked",
		},
		{
			name:           "invalid id returns 400",
			paramID:        "invalid",
			setUserID:      true,
			wantStatus:     http.StatusBadRequest,
			checkErrReturn: true,
		},
		{
			name:           "no userID returns 401",
			paramID:        "1",
			setUserID:      false,
			wantStatus:     http.StatusUnauthorized,
			checkErrReturn: true,
		},
		{
			name:    "usecase error: toggle fails returns 500",
			paramID: "1",
			setUserID: true,
			mockToggleLike: func(ctx context.Context, userID, galleryID uint) (bool, int, error) {
				return false, 0, assert.AnError
			},
			wantStatus:     http.StatusInternalServerError,
			checkErrReturn: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Logf("Running case: %s", tt.name)
			e := setupGalleryTestEcho()
			mockRepo := &mockGalleryLikeRepository{toggleLikeFunc: tt.mockToggleLike}
			uc := usecase.NewLikeUseCase(mockRepo, 10*time.Second)
			h := NewLikeHandler(uc)

			req := httptest.NewRequest(http.MethodPost, "/api/gallery/"+tt.paramID+"/like", nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			c.SetParamNames("id")
			c.SetParamValues(tt.paramID)
			if tt.setUserID {
				setupGalleryAuthContext(c)
			}

			err := h.Like(c)
			if tt.checkErrReturn {
				require.Error(t, err)
				if he, ok := err.(*echo.HTTPError); ok {
					assert.Equal(t, tt.wantStatus, he.Code)
				}
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.wantStatus, rec.Code)
			if tt.wantBodyHas != "" {
				assert.Contains(t, rec.Body.String(), tt.wantBodyHas)
			}
			t.Logf("Response: status=%d body=%s", rec.Code, rec.Body.String())
		})
	}
}
