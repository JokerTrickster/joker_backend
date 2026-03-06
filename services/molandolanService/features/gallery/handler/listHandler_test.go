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

type mockGalleryListRepository struct {
	listFunc func(ctx context.Context, page, limit int) ([]entity.GalleryPost, int64, error)
}

func (m *mockGalleryListRepository) List(ctx context.Context, page, limit int) ([]entity.GalleryPost, int64, error) {
	if m.listFunc != nil {
		return m.listFunc(ctx, page, limit)
	}
	return nil, 0, nil
}
func (m *mockGalleryListRepository) FindByID(ctx context.Context, id uint) (*entity.GalleryPost, error) {
	return nil, nil
}
func (m *mockGalleryListRepository) Create(ctx context.Context, post *entity.GalleryPost) (*entity.GalleryPost, error) {
	return nil, nil
}
func (m *mockGalleryListRepository) Delete(ctx context.Context, id uint) error {
	return nil
}
func (m *mockGalleryListRepository) IsLiked(ctx context.Context, userID, galleryID uint) (bool, error) {
	return false, nil
}
func (m *mockGalleryListRepository) ToggleLike(ctx context.Context, userID, galleryID uint) (bool, int, error) {
	return false, 0, nil
}
func (m *mockGalleryListRepository) ListComments(ctx context.Context, galleryID uint, page, limit int) ([]entity.GalleryComment, int64, error) {
	return nil, 0, nil
}
func (m *mockGalleryListRepository) FindCommentByID(ctx context.Context, id uint) (*entity.GalleryComment, error) {
	return nil, nil
}
func (m *mockGalleryListRepository) CreateComment(ctx context.Context, comment *entity.GalleryComment) (*entity.GalleryComment, error) {
	return nil, nil
}
func (m *mockGalleryListRepository) DeleteComment(ctx context.Context, id uint) error {
	return nil
}
func (m *mockGalleryListRepository) GetAuthorNickname(ctx context.Context, userID uint) (string, error) {
	return "", nil
}
func (m *mockGalleryListRepository) GetAuthorInfo(ctx context.Context, userID uint) (string, *string, error) {
	return "", nil, nil
}
func (m *mockGalleryListRepository) IsLikedBatch(ctx context.Context, userID uint, galleryIDs []uint) (map[uint]bool, error) {
	return map[uint]bool{}, nil
}
func (m *mockGalleryListRepository) GetUserRole(ctx context.Context, userID uint) (string, error) {
	return "user", nil
}

var _ _interface.IGalleryRepository = (*mockGalleryListRepository)(nil)

func TestListHandler_List(t *testing.T) {
	tests := []struct {
		name           string
		query          string
		mockList       func(ctx context.Context, page, limit int) ([]entity.GalleryPost, int64, error)
		wantStatus     int
		wantBodyHas    string
		checkErrReturn bool
	}{
		{
			name:  "success: returns 200 with items and pagination",
			query: "",
			mockList: func(ctx context.Context, page, limit int) ([]entity.GalleryPost, int64, error) {
				return []entity.GalleryPost{
					{ID: 1, AuthorID: 1, MediaType: "image", ThumbnailURL: "https://x.com/t.jpg", Caption: "cap", LikeCount: 2, CommentCount: 1},
				}, 1, nil
			},
			wantStatus:  http.StatusOK,
			wantBodyHas: "items",
		},
		{
			name:  "success: with page and limit query params",
			query: "?page=2&limit=10",
			mockList: func(ctx context.Context, page, limit int) ([]entity.GalleryPost, int64, error) {
				assert.Equal(t, 2, page)
				assert.Equal(t, 10, limit)
				return nil, 0, nil
			},
			wantStatus:  http.StatusOK,
			wantBodyHas: "pagination",
		},
		{
			name:  "usecase error: list fails returns 500",
			query: "",
			mockList: func(ctx context.Context, page, limit int) ([]entity.GalleryPost, int64, error) {
				return nil, 0, assert.AnError
			},
			wantStatus:     http.StatusInternalServerError,
			checkErrReturn: true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Logf("Running case: %s", tt.name)
			e := setupGalleryTestEcho()
			mockRepo := &mockGalleryListRepository{listFunc: tt.mockList}
			uc := usecase.NewListUseCase(mockRepo, 10*time.Second)
			h := NewListHandler(uc)

			path := "/api/gallery"
			if tt.query != "" {
				path += tt.query
			}
			req := httptest.NewRequest(http.MethodGet, path, nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			err := h.List(c)
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

func TestListHandler_List_Authenticated(t *testing.T) {
	t.Log("List with authenticated user: userID set in context -> 200 with isLiked")
	e := setupGalleryTestEcho()
	mockRepo := &mockGalleryListRepository{
		listFunc: func(ctx context.Context, page, limit int) ([]entity.GalleryPost, int64, error) {
			return []entity.GalleryPost{
				{ID: 1, AuthorID: 1, MediaType: "image", ThumbnailURL: "t1", Caption: "cap1"},
			}, 1, nil
		},
	}
	uc := usecase.NewListUseCase(mockRepo, 10*time.Second)
	h := NewListHandler(uc)

	req := httptest.NewRequest(http.MethodGet, "/api/gallery", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("userID", uint(1))

	err := h.List(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), "isLiked")
	assert.Contains(t, rec.Body.String(), "author")
	assert.Contains(t, rec.Body.String(), "caption")
	t.Logf("Authenticated list response: %s", rec.Body.String())
}
