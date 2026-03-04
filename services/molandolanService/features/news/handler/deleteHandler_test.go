package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/JokerTrickster/joker_backend/services/morandoranService/features/news/usecase"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/mock"
)

func TestDeleteHandler_Delete_Success(t *testing.T) {
	t.Log("Delete: valid ID -> 200 OK")
	e := setupNewsTestEcho()
	mockRepo := new(mockNewsRepository)
	uc := usecase.NewDeleteUseCase(mockRepo, defaultTimeout)
	h := NewDeleteHandler(uc)

	mockRepo.On("Delete", mock.Anything, uint(1)).Return(nil)

	req := httptest.NewRequest(http.MethodDelete, "/news/1", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/news/:id")
	c.SetParamNames("id")
	c.SetParamValues("1")

	err := h.Delete(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), "삭제되었습니다")
	mockRepo.AssertExpectations(t)
}

func TestDeleteHandler_Delete_InvalidID(t *testing.T) {
	t.Log("Delete: invalid ID format -> 400 Bad Request")
	e := setupNewsTestEcho()
	mockRepo := new(mockNewsRepository)
	uc := usecase.NewDeleteUseCase(mockRepo, defaultTimeout)
	h := NewDeleteHandler(uc)

	req := httptest.NewRequest(http.MethodDelete, "/news/abc", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/news/:id")
	c.SetParamNames("id")
	c.SetParamValues("abc")

	err := h.Delete(c)
	require.Error(t, err)
	he, ok := err.(*echo.HTTPError)
	require.True(t, ok)
	assert.Equal(t, http.StatusBadRequest, he.Code)
	assert.Equal(t, "INVALID_ID", he.Message)
	mockRepo.AssertNotCalled(t, "Delete")
}

func TestDeleteHandler_Delete_NotFound(t *testing.T) {
	t.Log("Delete: not found -> 404 Not Found")
	e := setupNewsTestEcho()
	mockRepo := new(mockNewsRepository)
	uc := usecase.NewDeleteUseCase(mockRepo, defaultTimeout)
	h := NewDeleteHandler(uc)

	mockRepo.On("Delete", mock.Anything, uint(999)).Return(assert.AnError)

	req := httptest.NewRequest(http.MethodDelete, "/news/999", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/news/:id")
	c.SetParamNames("id")
	c.SetParamValues("999")

	err := h.Delete(c)
	require.Error(t, err)
	he, ok := err.(*echo.HTTPError)
	require.True(t, ok)
	assert.Equal(t, http.StatusNotFound, he.Code)
	assert.Equal(t, "NOT_FOUND", he.Message)
	mockRepo.AssertExpectations(t)
}
