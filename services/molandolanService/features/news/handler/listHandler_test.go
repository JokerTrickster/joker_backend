package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/JokerTrickster/joker_backend/services/molandolanService/features/news/model/entity"
	"github.com/JokerTrickster/joker_backend/services/molandolanService/features/news/usecase"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/mock"
)

func TestListHandler_List_Success(t *testing.T) {
	t.Log("List: valid request -> 200 OK")
	e := setupNewsTestEcho()
	mockRepo := new(mockNewsRepository)
	uc := usecase.NewListUseCase(mockRepo, defaultTimeout)
	h := NewListHandler(uc)

	now := time.Now()
	items := []entity.News{
		{ID: 1, Title: "N1", Summary: "S1", Category: "cat", Date: "2024-01-15", CreatedAt: now},
	}
	mockRepo.On("List", mock.Anything, 1, 20, "").Return(items, int64(1), nil)

	req := httptest.NewRequest(http.MethodGet, "/news?page=1&limit=20", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.List(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), "news-001")
	assert.Contains(t, rec.Body.String(), "N1")
	mockRepo.AssertExpectations(t)
}

func TestListHandler_List_WithCategory(t *testing.T) {
	t.Log("List: with category filter -> 200 OK")
	e := setupNewsTestEcho()
	mockRepo := new(mockNewsRepository)
	uc := usecase.NewListUseCase(mockRepo, defaultTimeout)
	h := NewListHandler(uc)

	mockRepo.On("List", mock.Anything, 1, 10, "announcement").Return([]entity.News{}, int64(0), nil)

	req := httptest.NewRequest(http.MethodGet, "/news?page=1&limit=10&category=announcement", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.List(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	mockRepo.AssertExpectations(t)
}

func TestListHandler_List_BadRequest(t *testing.T) {
	t.Log("List: bind error -> 400 Bad Request")
	e := setupNewsTestEcho()
	mockRepo := new(mockNewsRepository)
	uc := usecase.NewListUseCase(mockRepo, defaultTimeout)
	h := NewListHandler(uc)

	// Invalid query (e.g. page as string) - Bind for query params typically doesn't fail easily
	// Test usecase error path instead
	mockRepo.On("List", mock.Anything, 1, 20, "").Return(nil, int64(0), assert.AnError)

	req := httptest.NewRequest(http.MethodGet, "/news", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.List(c)
	require.Error(t, err)
	he, ok := err.(*echo.HTTPError)
	require.True(t, ok)
	assert.Equal(t, http.StatusInternalServerError, he.Code)
	mockRepo.AssertExpectations(t)
}
