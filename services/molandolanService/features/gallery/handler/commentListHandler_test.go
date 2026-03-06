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

type mockGalleryCommentListRepository struct {
	listCommentsFunc func(ctx context.Context, galleryID uint, page, limit int) ([]entity.GalleryComment, int64, error)
	getAuthorNickname func(ctx context.Context, userID uint) (string, error)
}

func (m *mockGalleryCommentListRepository) List(ctx context.Context, page, limit int) ([]entity.GalleryPost, int64, error) {
	return nil, 0, nil
}
func (m *mockGalleryCommentListRepository) FindByID(ctx context.Context, id uint) (*entity.GalleryPost, error) {
	return nil, nil
}
func (m *mockGalleryCommentListRepository) Create(ctx context.Context, post *entity.GalleryPost) (*entity.GalleryPost, error) {
	return nil, nil
}
func (m *mockGalleryCommentListRepository) Delete(ctx context.Context, id uint) error {
	return nil
}
func (m *mockGalleryCommentListRepository) IsLiked(ctx context.Context, userID, galleryID uint) (bool, error) {
	return false, nil
}
func (m *mockGalleryCommentListRepository) ToggleLike(ctx context.Context, userID, galleryID uint) (bool, int, error) {
	return false, 0, nil
}
func (m *mockGalleryCommentListRepository) ListComments(ctx context.Context, galleryID uint, page, limit int) ([]entity.GalleryComment, int64, error) {
	if m.listCommentsFunc != nil {
		return m.listCommentsFunc(ctx, galleryID, page, limit)
	}
	return nil, 0, nil
}
func (m *mockGalleryCommentListRepository) FindCommentByID(ctx context.Context, id uint) (*entity.GalleryComment, error) {
	return nil, nil
}
func (m *mockGalleryCommentListRepository) CreateComment(ctx context.Context, comment *entity.GalleryComment) (*entity.GalleryComment, error) {
	return nil, nil
}
func (m *mockGalleryCommentListRepository) DeleteComment(ctx context.Context, id uint) error {
	return nil
}
func (m *mockGalleryCommentListRepository) GetAuthorNickname(ctx context.Context, userID uint) (string, error) {
	if m.getAuthorNickname != nil {
		return m.getAuthorNickname(ctx, userID)
	}
	return "user", nil
}
func (m *mockGalleryCommentListRepository) GetAuthorInfo(ctx context.Context, userID uint) (string, *string, error) {
	if m.getAuthorNickname != nil {
		nick, _ := m.getAuthorNickname(ctx, userID)
		return nick, nil, nil
	}
	return "user", nil, nil
}
func (m *mockGalleryCommentListRepository) IsLikedBatch(ctx context.Context, userID uint, galleryIDs []uint) (map[uint]bool, error) {
	return map[uint]bool{}, nil
}
func (m *mockGalleryCommentListRepository) GetUserRole(ctx context.Context, userID uint) (string, error) {
	return "user", nil
}

var _ _interface.IGalleryRepository = (*mockGalleryCommentListRepository)(nil)

func TestCommentListHandler_List(t *testing.T) {
	tests := []struct {
		name             string
		paramID          string
		query            string
		mockListComments func(ctx context.Context, galleryID uint, page, limit int) ([]entity.GalleryComment, int64, error)
		wantStatus       int
		wantBodyHas      string
		checkErrReturn   bool
	}{
		{
			name:    "success: returns 200 with comments and pagination",
			paramID: "1",
			query:   "",
			mockListComments: func(ctx context.Context, galleryID uint, page, limit int) ([]entity.GalleryComment, int64, error) {
				return []entity.GalleryComment{
					{ID: 1, GalleryID: 1, AuthorID: 1, Content: "nice"},
				}, 1, nil
			},
			wantStatus:  http.StatusOK,
			wantBodyHas: "items",
		},
		{
			name:           "invalid gallery id returns 400",
			paramID:        "invalid",
			query:          "",
			wantStatus:     http.StatusBadRequest,
			checkErrReturn: true,
		},
		{
			name:    "usecase error: list fails returns 500",
			paramID: "1",
			mockListComments: func(ctx context.Context, galleryID uint, page, limit int) ([]entity.GalleryComment, int64, error) {
				return nil, 0, assert.AnError
			},
			wantStatus:     http.StatusInternalServerError,
			checkErrReturn: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Logf("Running case: %s", tt.name)
			e := setupGalleryTestEcho()
			mockRepo := &mockGalleryCommentListRepository{listCommentsFunc: tt.mockListComments}
			uc := usecase.NewCommentListUseCase(mockRepo, 10*time.Second)
			h := NewCommentListHandler(uc)

			path := "/api/gallery/" + tt.paramID + "/comments"
			if tt.query != "" {
				path += tt.query
			}
			req := httptest.NewRequest(http.MethodGet, path, nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			c.SetParamNames("id")
			c.SetParamValues(tt.paramID)

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
