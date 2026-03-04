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
	"github.com/JokerTrickster/joker_backend/shared/db/mysql"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockGalleryCommentDeleteRepository struct {
	findCommentFunc func(ctx context.Context, id uint) (*entity.GalleryComment, error)
	deleteFunc      func(ctx context.Context, id uint) error
}

func (m *mockGalleryCommentDeleteRepository) List(ctx context.Context, page, limit int) ([]entity.GalleryPost, int64, error) {
	return nil, 0, nil
}
func (m *mockGalleryCommentDeleteRepository) FindByID(ctx context.Context, id uint) (*entity.GalleryPost, error) {
	return nil, nil
}
func (m *mockGalleryCommentDeleteRepository) Create(ctx context.Context, post *entity.GalleryPost) (*entity.GalleryPost, error) {
	return nil, nil
}
func (m *mockGalleryCommentDeleteRepository) Delete(ctx context.Context, id uint) error {
	return nil
}
func (m *mockGalleryCommentDeleteRepository) IsLiked(ctx context.Context, userID, galleryID uint) (bool, error) {
	return false, nil
}
func (m *mockGalleryCommentDeleteRepository) ToggleLike(ctx context.Context, userID, galleryID uint) (bool, int, error) {
	return false, 0, nil
}
func (m *mockGalleryCommentDeleteRepository) ListComments(ctx context.Context, galleryID uint, page, limit int) ([]entity.GalleryComment, int64, error) {
	return nil, 0, nil
}
func (m *mockGalleryCommentDeleteRepository) FindCommentByID(ctx context.Context, id uint) (*entity.GalleryComment, error) {
	if m.findCommentFunc != nil {
		return m.findCommentFunc(ctx, id)
	}
	return nil, nil
}
func (m *mockGalleryCommentDeleteRepository) CreateComment(ctx context.Context, comment *entity.GalleryComment) (*entity.GalleryComment, error) {
	return nil, nil
}
func (m *mockGalleryCommentDeleteRepository) DeleteComment(ctx context.Context, id uint) error {
	if m.deleteFunc != nil {
		return m.deleteFunc(ctx, id)
	}
	return nil
}
func (m *mockGalleryCommentDeleteRepository) GetAuthorNickname(ctx context.Context, userID uint) (string, error) {
	return "", nil
}
func (m *mockGalleryCommentDeleteRepository) GetAuthorInfo(ctx context.Context, userID uint) (string, *string, error) {
	return "", nil, nil
}
func (m *mockGalleryCommentDeleteRepository) IsLikedBatch(ctx context.Context, userID uint, galleryIDs []uint) (map[uint]bool, error) {
	return map[uint]bool{}, nil
}

var _ _interface.IGalleryRepository = (*mockGalleryCommentDeleteRepository)(nil)

func TestCommentDeleteHandler_Delete_InvalidCommentID(t *testing.T) {
	e := setupGalleryTestEcho()
	mockRepo := &mockGalleryCommentDeleteRepository{}
	uc := usecase.NewCommentDeleteUseCase(mockRepo, 10*time.Second)
	h := NewCommentDeleteHandler(uc)

	req := httptest.NewRequest(http.MethodDelete, "/api/gallery/1/comments/invalid", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id", "commentId")
	c.SetParamValues("1", "invalid")
	setupGalleryAuthContext(c)

	err := h.Delete(c)
	require.Error(t, err)
	if he, ok := err.(*echo.HTTPError); ok {
		assert.Equal(t, http.StatusBadRequest, he.Code)
	}
	t.Logf("Invalid commentId returns 400: %v", err)
}

func TestCommentDeleteHandler_Delete_NoUserID(t *testing.T) {
	e := setupGalleryTestEcho()
	mockRepo := &mockGalleryCommentDeleteRepository{}
	uc := usecase.NewCommentDeleteUseCase(mockRepo, 10*time.Second)
	h := NewCommentDeleteHandler(uc)

	req := httptest.NewRequest(http.MethodDelete, "/api/gallery/1/comments/1", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id", "commentId")
	c.SetParamValues("1", "1")
	// do NOT set userID

	err := h.Delete(c)
	require.Error(t, err)
	if he, ok := err.(*echo.HTTPError); ok {
		assert.Equal(t, http.StatusUnauthorized, he.Code)
	}
	t.Logf("No userID returns 401: %v", err)
}

func TestCommentDeleteHandler_Delete_DBRequired(t *testing.T) {
	if mysql.GormMysqlDB == nil {
		t.Skip("Skipping: mysql.GormMysqlDB not initialized (DB-dependent test)")
	}
	t.Skip("Skipping: full comment delete flow requires DB for Raw role query - run as integration test when DB available")
}
