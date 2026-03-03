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

type mockGalleryCreateRepository struct {
	createFunc         func(ctx context.Context, post *entity.GalleryPost) (*entity.GalleryPost, error)
	getAuthorNickname  func(ctx context.Context, userID uint) (string, error)
}

func (m *mockGalleryCreateRepository) List(ctx context.Context, page, limit int) ([]entity.GalleryPost, int64, error) {
	return nil, 0, nil
}
func (m *mockGalleryCreateRepository) FindByID(ctx context.Context, id uint) (*entity.GalleryPost, error) {
	return nil, nil
}
func (m *mockGalleryCreateRepository) Create(ctx context.Context, post *entity.GalleryPost) (*entity.GalleryPost, error) {
	if m.createFunc != nil {
		return m.createFunc(ctx, post)
	}
	return post, nil
}
func (m *mockGalleryCreateRepository) Delete(ctx context.Context, id uint) error {
	return nil
}
func (m *mockGalleryCreateRepository) IsLiked(ctx context.Context, userID, galleryID uint) (bool, error) {
	return false, nil
}
func (m *mockGalleryCreateRepository) ToggleLike(ctx context.Context, userID, galleryID uint) (bool, int, error) {
	return false, 0, nil
}
func (m *mockGalleryCreateRepository) ListComments(ctx context.Context, galleryID uint, page, limit int) ([]entity.GalleryComment, int64, error) {
	return nil, 0, nil
}
func (m *mockGalleryCreateRepository) FindCommentByID(ctx context.Context, id uint) (*entity.GalleryComment, error) {
	return nil, nil
}
func (m *mockGalleryCreateRepository) CreateComment(ctx context.Context, comment *entity.GalleryComment) (*entity.GalleryComment, error) {
	return nil, nil
}
func (m *mockGalleryCreateRepository) DeleteComment(ctx context.Context, id uint) error {
	return nil
}
func (m *mockGalleryCreateRepository) GetAuthorNickname(ctx context.Context, userID uint) (string, error) {
	if m.getAuthorNickname != nil {
		return m.getAuthorNickname(ctx, userID)
	}
	return "testuser", nil
}

var _ _interface.IGalleryRepository = (*mockGalleryCreateRepository)(nil)

func TestCreateHandler_Create(t *testing.T) {
	tests := []struct {
		name           string
		body           []byte
		setUserID      bool
		mockCreate     func(ctx context.Context, post *entity.GalleryPost) (*entity.GalleryPost, error)
		mockGetNickname func(ctx context.Context, userID uint) (string, error)
		wantStatus     int
		wantBodyHas    string
		checkErrReturn bool
	}{
		{
			name: "success: valid request creates gallery returns 201",
			body: mustJSON(t, map[string]string{
				"mediaUrl":     "https://example.com/media.jpg",
				"thumbnailUrl": "https://example.com/thumb.jpg",
				"mediaType":    "image",
				"caption":      "test caption",
			}),
			setUserID: true,
			mockCreate: func(ctx context.Context, post *entity.GalleryPost) (*entity.GalleryPost, error) {
				post.ID = 1
				return post, nil
			},
			mockGetNickname: func(ctx context.Context, userID uint) (string, error) {
				return "testuser", nil
			},
			wantStatus:  http.StatusCreated,
			wantBodyHas: "gallery-001",
		},
		{
			name:           "no userID in context returns 401",
			body:           mustJSON(t, map[string]string{"mediaUrl": "https://x.com/a.jpg", "thumbnailUrl": "https://x.com/t.jpg", "mediaType": "image"}),
			setUserID:      false,
			wantStatus:     http.StatusUnauthorized,
			checkErrReturn: true,
		},
		{
			name: "validation error: missing mediaUrl returns 400",
			body: mustJSON(t, map[string]string{
				"thumbnailUrl": "https://example.com/thumb.jpg",
				"mediaType":    "image",
			}),
			setUserID:      true,
			wantStatus:     http.StatusBadRequest,
			checkErrReturn: true,
		},
		{
			name: "validation error: missing thumbnailUrl returns 400",
			body: mustJSON(t, map[string]string{
				"mediaUrl":  "https://example.com/media.jpg",
				"mediaType": "image",
			}),
			setUserID:      true,
			wantStatus:     http.StatusBadRequest,
			checkErrReturn: true,
		},
		{
			name: "validation error: missing mediaType returns 400",
			body: mustJSON(t, map[string]string{
				"mediaUrl":     "https://example.com/media.jpg",
				"thumbnailUrl": "https://example.com/thumb.jpg",
			}),
			setUserID:      true,
			wantStatus:     http.StatusBadRequest,
			checkErrReturn: true,
		},
		{
			name: "usecase error: create fails returns 500",
			body: mustJSON(t, map[string]string{
				"mediaUrl":     "https://example.com/media.jpg",
				"thumbnailUrl": "https://example.com/thumb.jpg",
				"mediaType":    "image",
			}),
			setUserID: true,
			mockCreate: func(ctx context.Context, post *entity.GalleryPost) (*entity.GalleryPost, error) {
				return nil, assert.AnError
			},
			wantStatus:     http.StatusInternalServerError,
			checkErrReturn: true,
		},
		{
			name:           "bind error: malformed JSON returns 400",
			body:           []byte(`{"mediaUrl": "x", "thumbnailUrl": "y"`),
			setUserID:      true,
			wantStatus:     http.StatusBadRequest,
			checkErrReturn: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Logf("Running case: %s", tt.name)
			e := setupGalleryTestEcho()
			mockRepo := &mockGalleryCreateRepository{}
			if tt.mockCreate != nil {
				mockRepo.createFunc = tt.mockCreate
			}
			if tt.mockGetNickname != nil {
				mockRepo.getAuthorNickname = tt.mockGetNickname
			}
			uc := usecase.NewCreateUseCase(mockRepo, 10*time.Second)
			h := NewCreateHandler(uc)

			req := newJSONRequest(t, http.MethodPost, "/api/gallery", tt.body)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			if tt.setUserID {
				setupGalleryAuthContext(c)
			}

			err := h.Create(c)
			if tt.checkErrReturn {
				require.Error(t, err)
				if he, ok := err.(*echo.HTTPError); ok {
					assert.Equal(t, tt.wantStatus, he.Code)
					t.Logf("Error: code=%d message=%v", he.Code, he.Message)
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
