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

type mockCommentCreateHandlerGalleryRepository struct {
	createCommentFunc  func(ctx context.Context, comment *entity.GalleryComment) (*entity.GalleryComment, error)
	getAuthorNickname  func(ctx context.Context, userID uint) (string, error)
}

func (m *mockCommentCreateHandlerGalleryRepository) List(ctx context.Context, page, limit int) ([]entity.GalleryPost, int64, error) {
	return nil, 0, nil
}
func (m *mockCommentCreateHandlerGalleryRepository) FindByID(ctx context.Context, id uint) (*entity.GalleryPost, error) {
	return nil, nil
}
func (m *mockCommentCreateHandlerGalleryRepository) Create(ctx context.Context, post *entity.GalleryPost) (*entity.GalleryPost, error) {
	return nil, nil
}
func (m *mockCommentCreateHandlerGalleryRepository) Delete(ctx context.Context, id uint) error {
	return nil
}
func (m *mockCommentCreateHandlerGalleryRepository) IsLiked(ctx context.Context, userID, galleryID uint) (bool, error) {
	return false, nil
}
func (m *mockCommentCreateHandlerGalleryRepository) ToggleLike(ctx context.Context, userID, galleryID uint) (bool, int, error) {
	return false, 0, nil
}
func (m *mockCommentCreateHandlerGalleryRepository) ListComments(ctx context.Context, galleryID uint, page, limit int) ([]entity.GalleryComment, int64, error) {
	return nil, 0, nil
}
func (m *mockCommentCreateHandlerGalleryRepository) FindCommentByID(ctx context.Context, id uint) (*entity.GalleryComment, error) {
	return nil, nil
}
func (m *mockCommentCreateHandlerGalleryRepository) CreateComment(ctx context.Context, comment *entity.GalleryComment) (*entity.GalleryComment, error) {
	if m.createCommentFunc != nil {
		return m.createCommentFunc(ctx, comment)
	}
	comment.ID = 1
	return comment, nil
}
func (m *mockCommentCreateHandlerGalleryRepository) DeleteComment(ctx context.Context, id uint) error {
	return nil
}
func (m *mockCommentCreateHandlerGalleryRepository) GetAuthorNickname(ctx context.Context, userID uint) (string, error) {
	if m.getAuthorNickname != nil {
		return m.getAuthorNickname(ctx, userID)
	}
	return "user", nil
}

var _ _interface.IGalleryRepository = (*mockCommentCreateHandlerGalleryRepository)(nil)

func TestCommentCreateHandler_Create_Success(t *testing.T) {
	t.Log("TestCommentCreateHandler_Create_Success: valid path param, auth, and body -> 201")
	mockRepo := &mockCommentCreateHandlerGalleryRepository{
		createCommentFunc: func(ctx context.Context, comment *entity.GalleryComment) (*entity.GalleryComment, error) {
			comment.ID = 1
			return comment, nil
		},
		getAuthorNickname: func(ctx context.Context, userID uint) (string, error) {
			return "author", nil
		},
	}
	uc := usecase.NewCommentCreateUseCase(mockRepo, 10*time.Second)
	h := NewCommentCreateHandler(uc)

	e := setupGalleryTestEcho()
	e.POST("/gallery/:id/comments", h.Create)

	body := mustJSON(t, map[string]string{"content": "nice post!"})
	req := newJSONRequest(t, http.MethodPost, "/gallery/1/comments", body)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("1")
	setupGalleryAuthContext(c)

	err := h.Create(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusCreated, rec.Code)
	assert.Contains(t, rec.Body.String(), "comment-001")
	assert.Contains(t, rec.Body.String(), "nice post!")
	t.Logf("Response: status=%d body=%s", rec.Code, rec.Body.String())
}

func TestCommentCreateHandler_Create_InvalidID(t *testing.T) {
	t.Log("TestCommentCreateHandler_Create_InvalidID: bad path param -> 400")
	mockRepo := &mockCommentCreateHandlerGalleryRepository{}
	uc := usecase.NewCommentCreateUseCase(mockRepo, 10*time.Second)
	h := NewCommentCreateHandler(uc)

	e := setupGalleryTestEcho()
	e.POST("/gallery/:id/comments", h.Create)

	body := mustJSON(t, map[string]string{"content": "hi"})
	req := newJSONRequest(t, http.MethodPost, "/gallery/invalid/comments", body)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("invalid")

	err := h.Create(c)
	require.Error(t, err)
	he, ok := err.(*echo.HTTPError)
	require.True(t, ok)
	assert.Equal(t, http.StatusBadRequest, he.Code)
	assert.Equal(t, "INVALID_ID", he.Message)
	t.Logf("Error: code=%d message=%v", he.Code, he.Message)
}

func TestCommentCreateHandler_Create_NoAuth(t *testing.T) {
	t.Log("TestCommentCreateHandler_Create_NoAuth: no userID in context -> 401")
	mockRepo := &mockCommentCreateHandlerGalleryRepository{}
	uc := usecase.NewCommentCreateUseCase(mockRepo, 10*time.Second)
	h := NewCommentCreateHandler(uc)

	e := setupGalleryTestEcho()
	e.POST("/gallery/:id/comments", h.Create)

	body := mustJSON(t, map[string]string{"content": "hi"})
	req := newJSONRequest(t, http.MethodPost, "/gallery/1/comments", body)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("1")
	// no setupGalleryAuthContext - userID not set

	err := h.Create(c)
	require.Error(t, err)
	he, ok := err.(*echo.HTTPError)
	require.True(t, ok)
	assert.Equal(t, http.StatusUnauthorized, he.Code)
	assert.Equal(t, "UNAUTHORIZED", he.Message)
	t.Logf("Error: code=%d message=%v", he.Code, he.Message)
}

func TestCommentCreateHandler_Create_EmptyContent(t *testing.T) {
	t.Log("TestCommentCreateHandler_Create_EmptyContent: empty content fails validation -> 400")
	mockRepo := &mockCommentCreateHandlerGalleryRepository{}
	uc := usecase.NewCommentCreateUseCase(mockRepo, 10*time.Second)
	h := NewCommentCreateHandler(uc)

	e := setupGalleryTestEcho()
	e.POST("/gallery/:id/comments", h.Create)

	body := mustJSON(t, map[string]string{"content": ""})
	req := newJSONRequest(t, http.MethodPost, "/gallery/1/comments", body)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("1")
	setupGalleryAuthContext(c)

	err := h.Create(c)
	require.Error(t, err)
	he, ok := err.(*echo.HTTPError)
	require.True(t, ok)
	assert.Equal(t, http.StatusBadRequest, he.Code)
	t.Logf("Error: code=%d message=%v", he.Code, he.Message)
}

func TestCommentCreateHandler_Create_UseCaseError(t *testing.T) {
	t.Log("TestCommentCreateHandler_Create_UseCaseError: repo returns error -> 500")
	mockRepo := &mockCommentCreateHandlerGalleryRepository{
		createCommentFunc: func(ctx context.Context, comment *entity.GalleryComment) (*entity.GalleryComment, error) {
			return nil, assert.AnError
		},
	}
	uc := usecase.NewCommentCreateUseCase(mockRepo, 10*time.Second)
	h := NewCommentCreateHandler(uc)

	e := setupGalleryTestEcho()
	e.POST("/gallery/:id/comments", h.Create)

	body := mustJSON(t, map[string]string{"content": "hi"})
	req := newJSONRequest(t, http.MethodPost, "/gallery/1/comments", body)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("1")
	setupGalleryAuthContext(c)

	err := h.Create(c)
	require.Error(t, err)
	he, ok := err.(*echo.HTTPError)
	require.True(t, ok)
	assert.Equal(t, http.StatusInternalServerError, he.Code)
	assert.Equal(t, "INTERNAL_ERROR", he.Message)
	t.Logf("Error: code=%d message=%v", he.Code, he.Message)
}
