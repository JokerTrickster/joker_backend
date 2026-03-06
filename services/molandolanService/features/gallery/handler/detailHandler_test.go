package handler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/JokerTrickster/joker_backend/services/molandolanService/features/gallery/model/entity"
	_interface "github.com/JokerTrickster/joker_backend/services/molandolanService/features/gallery/model/interface"
	"github.com/JokerTrickster/joker_backend/services/molandolanService/features/gallery/usecase"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockGalleryDetailRepository struct {
	findByIDFunc      func(ctx context.Context, id uint) (*entity.GalleryPost, error)
	isLikedFunc       func(ctx context.Context, userID, galleryID uint) (bool, error)
	getAuthorNickname func(ctx context.Context, userID uint) (string, error)
}

func (m *mockGalleryDetailRepository) List(ctx context.Context, page, limit int) ([]entity.GalleryPost, int64, error) {
	return nil, 0, nil
}
func (m *mockGalleryDetailRepository) FindByID(ctx context.Context, id uint) (*entity.GalleryPost, error) {
	if m.findByIDFunc != nil {
		return m.findByIDFunc(ctx, id)
	}
	return nil, nil
}
func (m *mockGalleryDetailRepository) Create(ctx context.Context, post *entity.GalleryPost) (*entity.GalleryPost, error) {
	return nil, nil
}
func (m *mockGalleryDetailRepository) Delete(ctx context.Context, id uint) error {
	return nil
}
func (m *mockGalleryDetailRepository) IsLiked(ctx context.Context, userID, galleryID uint) (bool, error) {
	if m.isLikedFunc != nil {
		return m.isLikedFunc(ctx, userID, galleryID)
	}
	return false, nil
}
func (m *mockGalleryDetailRepository) ToggleLike(ctx context.Context, userID, galleryID uint) (bool, int, error) {
	return false, 0, nil
}
func (m *mockGalleryDetailRepository) ListComments(ctx context.Context, galleryID uint, page, limit int) ([]entity.GalleryComment, int64, error) {
	return nil, 0, nil
}
func (m *mockGalleryDetailRepository) FindCommentByID(ctx context.Context, id uint) (*entity.GalleryComment, error) {
	return nil, nil
}
func (m *mockGalleryDetailRepository) CreateComment(ctx context.Context, comment *entity.GalleryComment) (*entity.GalleryComment, error) {
	return nil, nil
}
func (m *mockGalleryDetailRepository) DeleteComment(ctx context.Context, id uint) error {
	return nil
}
func (m *mockGalleryDetailRepository) GetAuthorNickname(ctx context.Context, userID uint) (string, error) {
	if m.getAuthorNickname != nil {
		return m.getAuthorNickname(ctx, userID)
	}
	return "testuser", nil
}
func (m *mockGalleryDetailRepository) GetAuthorInfo(ctx context.Context, userID uint) (string, *string, error) {
	if m.getAuthorNickname != nil {
		nick, _ := m.getAuthorNickname(ctx, userID)
		return nick, nil, nil
	}
	return "testuser", nil, nil
}
func (m *mockGalleryDetailRepository) IsLikedBatch(ctx context.Context, userID uint, galleryIDs []uint) (map[uint]bool, error) {
	return map[uint]bool{}, nil
}
func (m *mockGalleryDetailRepository) GetUserRole(ctx context.Context, userID uint) (string, error) {
	return "user", nil
}

var _ _interface.IGalleryRepository = (*mockGalleryDetailRepository)(nil)

func TestDetailHandler_Detail(t *testing.T) {
	tests := []struct {
		name           string
		paramID        string
		setUserID      bool
		mockFindByID   func(ctx context.Context, id uint) (*entity.GalleryPost, error)
		mockIsLiked    func(ctx context.Context, userID, galleryID uint) (bool, error)
		mockGetNickname func(ctx context.Context, userID uint) (string, error)
		wantStatus     int
		wantBodyHas    string
		checkErrReturn bool
	}{
		{
			name:    "success: valid id returns 200 and detail",
			paramID: "1",
			setUserID: true,
			mockFindByID: func(ctx context.Context, id uint) (*entity.GalleryPost, error) {
				return &entity.GalleryPost{
					ID:           1,
					AuthorID:     1,
					MediaType:    "image",
					MediaURL:     "https://example.com/media.jpg",
					ThumbnailURL: "https://example.com/thumb.jpg",
					Caption:      "test",
					LikeCount:    5,
					CommentCount: 2,
				}, nil
			},
			mockGetNickname: func(ctx context.Context, userID uint) (string, error) {
				return "testuser", nil
			},
			wantStatus:  http.StatusOK,
			wantBodyHas: "gallery-001",
		},
		{
			name:    "success: no userID (optional auth) returns 200 with isLiked false",
			paramID: "1",
			setUserID: false,
			mockFindByID: func(ctx context.Context, id uint) (*entity.GalleryPost, error) {
				return &entity.GalleryPost{
					ID: 1, AuthorID: 1, MediaType: "image",
					MediaURL: "https://x.com/a.jpg", ThumbnailURL: "https://x.com/t.jpg",
					LikeCount: 0, CommentCount: 0,
				}, nil
			},
			mockGetNickname: func(ctx context.Context, userID uint) (string, error) {
				return "author", nil
			},
			wantStatus:  http.StatusOK,
			wantBodyHas: "gallery-001",
		},
		{
			name:           "invalid id returns 400",
			paramID:        "invalid",
			setUserID:      true,
			wantStatus:     http.StatusBadRequest,
			checkErrReturn: true,
		},
		{
			name:    "not found: post does not exist returns 404",
			paramID: "999",
			setUserID: true,
			mockFindByID: func(ctx context.Context, id uint) (*entity.GalleryPost, error) {
				return nil, assert.AnError
			},
			wantStatus:     http.StatusNotFound,
			checkErrReturn: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Logf("Running case: %s", tt.name)
			e := setupGalleryTestEcho()
			mockRepo := &mockGalleryDetailRepository{}
			if tt.mockFindByID != nil {
				mockRepo.findByIDFunc = tt.mockFindByID
			}
			if tt.mockIsLiked != nil {
				mockRepo.isLikedFunc = tt.mockIsLiked
			}
			if tt.mockGetNickname != nil {
				mockRepo.getAuthorNickname = tt.mockGetNickname
			}
			uc := usecase.NewDetailUseCase(mockRepo, 10*time.Second)
			h := NewDetailHandler(uc)

			req := httptest.NewRequest(http.MethodGet, "/api/gallery/"+tt.paramID, nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			c.SetParamNames("id")
			c.SetParamValues(tt.paramID)
			if tt.setUserID {
				setupGalleryAuthContext(c)
			}

			err := h.Detail(c)
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
