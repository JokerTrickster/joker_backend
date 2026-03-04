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
	"github.com/JokerTrickster/joker_backend/shared/db/mysql"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockGalleryDeleteRepository struct {
	findByIDFunc func(ctx context.Context, id uint) (*entity.GalleryPost, error)
	deleteFunc   func(ctx context.Context, id uint) error
}

func (m *mockGalleryDeleteRepository) List(ctx context.Context, page, limit int) ([]entity.GalleryPost, int64, error) {
	return nil, 0, nil
}
func (m *mockGalleryDeleteRepository) FindByID(ctx context.Context, id uint) (*entity.GalleryPost, error) {
	if m.findByIDFunc != nil {
		return m.findByIDFunc(ctx, id)
	}
	return nil, nil
}
func (m *mockGalleryDeleteRepository) Create(ctx context.Context, post *entity.GalleryPost) (*entity.GalleryPost, error) {
	return nil, nil
}
func (m *mockGalleryDeleteRepository) Delete(ctx context.Context, id uint) error {
	if m.deleteFunc != nil {
		return m.deleteFunc(ctx, id)
	}
	return nil
}
func (m *mockGalleryDeleteRepository) IsLiked(ctx context.Context, userID, galleryID uint) (bool, error) {
	return false, nil
}
func (m *mockGalleryDeleteRepository) ToggleLike(ctx context.Context, userID, galleryID uint) (bool, int, error) {
	return false, 0, nil
}
func (m *mockGalleryDeleteRepository) ListComments(ctx context.Context, galleryID uint, page, limit int) ([]entity.GalleryComment, int64, error) {
	return nil, 0, nil
}
func (m *mockGalleryDeleteRepository) FindCommentByID(ctx context.Context, id uint) (*entity.GalleryComment, error) {
	return nil, nil
}
func (m *mockGalleryDeleteRepository) CreateComment(ctx context.Context, comment *entity.GalleryComment) (*entity.GalleryComment, error) {
	return nil, nil
}
func (m *mockGalleryDeleteRepository) DeleteComment(ctx context.Context, id uint) error {
	return nil
}
func (m *mockGalleryDeleteRepository) GetAuthorNickname(ctx context.Context, userID uint) (string, error) {
	return "", nil
}

var _ _interface.IGalleryRepository = (*mockGalleryDeleteRepository)(nil)

func TestDeleteHandler_Delete_InvalidID(t *testing.T) {
	e := setupGalleryTestEcho()
	mockRepo := &mockGalleryDeleteRepository{}
	uc := usecase.NewDeleteUseCase(mockRepo, 10*time.Second)
	h := NewDeleteHandler(uc)

	req := httptest.NewRequest(http.MethodDelete, "/api/gallery/invalid", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("invalid")
	setupGalleryAuthContext(c)

	err := h.Delete(c)
	require.Error(t, err)
	if he, ok := err.(*echo.HTTPError); ok {
		assert.Equal(t, http.StatusBadRequest, he.Code)
		assert.Equal(t, "INVALID_ID", he.Message)
	}
	t.Logf("Invalid ID returns 400: %v", err)
}

func TestDeleteHandler_Delete_NoUserID(t *testing.T) {
	e := setupGalleryTestEcho()
	mockRepo := &mockGalleryDeleteRepository{}
	uc := usecase.NewDeleteUseCase(mockRepo, 10*time.Second)
	h := NewDeleteHandler(uc)

	req := httptest.NewRequest(http.MethodDelete, "/api/gallery/1", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("1")
	// do NOT set userID

	err := h.Delete(c)
	require.Error(t, err)
	if he, ok := err.(*echo.HTTPError); ok {
		assert.Equal(t, http.StatusUnauthorized, he.Code)
	}
	t.Logf("No userID returns 401: %v", err)
}

func TestDeleteHandler_Delete_DBRequired(t *testing.T) {
	// deleteHandler uses mysql.GormMysqlDB.Raw for role lookup - skip when DB unavailable
	if mysql.GormMysqlDB == nil {
		t.Skip("Skipping: mysql.GormMysqlDB not initialized (DB-dependent test)")
	}
	t.Skip("Skipping: full delete flow requires DB for Raw role query - run as integration test when DB available")
}
