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

func TestDetailHandler_Detail_Success(t *testing.T) {
	t.Log("Detail: valid ID -> 200 OK")
	e := setupNewsTestEcho()
	mockRepo := new(mockNewsRepository)
	uc := usecase.NewDetailUseCase(mockRepo, defaultTimeout)
	h := NewDetailHandler(uc)

	now := time.Now()
	mockRepo.On("FindByID", mock.Anything, uint(1)).Return(&entity.News{
		ID: 1, Title: "Title", Summary: "Sum", Content: "Body", Thumbnail: "", Category: "cat", Date: "2024-01-15",
		CreatedAt: now, UpdatedAt: now,
	}, nil)

	req := httptest.NewRequest(http.MethodGet, "/news/1", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/news/:id")
	c.SetParamNames("id")
	c.SetParamValues("1")

	err := h.Detail(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), "news-001")
	assert.Contains(t, rec.Body.String(), "Title")
	mockRepo.AssertExpectations(t)
}

func TestDetailHandler_Detail_InvalidID(t *testing.T) {
	t.Log("Detail: invalid ID format -> 400 Bad Request")
	e := setupNewsTestEcho()
	mockRepo := new(mockNewsRepository)
	uc := usecase.NewDetailUseCase(mockRepo, defaultTimeout)
	h := NewDetailHandler(uc)

	req := httptest.NewRequest(http.MethodGet, "/news/invalid", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/news/:id")
	c.SetParamNames("id")
	c.SetParamValues("invalid")

	err := h.Detail(c)
	require.Error(t, err)
	he, ok := err.(*echo.HTTPError)
	require.True(t, ok)
	assert.Equal(t, http.StatusBadRequest, he.Code)
	assert.Equal(t, "INVALID_ID", he.Message)
	mockRepo.AssertNotCalled(t, "FindByID")
}

func TestDetailHandler_Detail_NotFound(t *testing.T) {
	t.Log("Detail: not found -> 404 Not Found")
	e := setupNewsTestEcho()
	mockRepo := new(mockNewsRepository)
	uc := usecase.NewDetailUseCase(mockRepo, defaultTimeout)
	h := NewDetailHandler(uc)

	mockRepo.On("FindByID", mock.Anything, uint(999)).Return(nil, assert.AnError)

	req := httptest.NewRequest(http.MethodGet, "/news/999", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/news/:id")
	c.SetParamNames("id")
	c.SetParamValues("999")

	err := h.Detail(c)
	require.Error(t, err)
	he, ok := err.(*echo.HTTPError)
	require.True(t, ok)
	assert.Equal(t, http.StatusNotFound, he.Code)
	assert.Equal(t, "NOT_FOUND", he.Message)
	mockRepo.AssertExpectations(t)
}
