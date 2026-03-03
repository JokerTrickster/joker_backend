package handler

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/JokerTrickster/joker_backend/services/morandoranService/features/news/model/entity"
	"github.com/JokerTrickster/joker_backend/services/morandoranService/features/news/model/request"
	"github.com/JokerTrickster/joker_backend/services/morandoranService/features/news/usecase"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/mock"
)

func TestUpdateHandler_Update_Success(t *testing.T) {
	t.Log("Update: valid request -> 200 OK")
	e := setupNewsTestEcho()
	mockRepo := new(mockNewsRepository)
	uc := usecase.NewUpdateUseCase(mockRepo, defaultTimeout)
	h := NewUpdateHandler(uc)

	title := "Updated Title"
	reqBody := &request.ReqUpdateNews{Title: &title}
	body := newsMustJSON(t, reqBody)
	now := time.Now()
	mockRepo.On("Update", mock.Anything, uint(1), mock.AnythingOfType("map[string]interface {}")).Return(&entity.News{
		ID: 1, Title: "Updated Title", Summary: "S", Content: "C", Thumbnail: "", Category: "cat", Date: "2024-01-15",
		CreatedAt: now, UpdatedAt: now,
	}, nil)

	req := httptest.NewRequest(http.MethodPut, "/news/1", bytes.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/news/:id")
	c.SetParamNames("id")
	c.SetParamValues("1")

	err := h.Update(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), "news-001")
	assert.Contains(t, rec.Body.String(), "Updated Title")
	mockRepo.AssertExpectations(t)
}

func TestUpdateHandler_Update_InvalidID(t *testing.T) {
	t.Log("Update: invalid ID format -> 400 Bad Request")
	e := setupNewsTestEcho()
	mockRepo := new(mockNewsRepository)
	uc := usecase.NewUpdateUseCase(mockRepo, defaultTimeout)
	h := NewUpdateHandler(uc)

	reqBody := &request.ReqUpdateNews{}
	body := newsMustJSON(t, reqBody)
	req := httptest.NewRequest(http.MethodPut, "/news/xyz", bytes.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/news/:id")
	c.SetParamNames("id")
	c.SetParamValues("xyz")

	err := h.Update(c)
	require.Error(t, err)
	he, ok := err.(*echo.HTTPError)
	require.True(t, ok)
	assert.Equal(t, http.StatusBadRequest, he.Code)
	assert.Equal(t, "INVALID_ID", he.Message)
	mockRepo.AssertNotCalled(t, "Update")
}

func TestUpdateHandler_Update_NotFound(t *testing.T) {
	t.Log("Update: not found -> 404 Not Found")
	e := setupNewsTestEcho()
	mockRepo := new(mockNewsRepository)
	uc := usecase.NewUpdateUseCase(mockRepo, defaultTimeout)
	h := NewUpdateHandler(uc)

	title := "Updated"
	reqBody := &request.ReqUpdateNews{Title: &title}
	body := newsMustJSON(t, reqBody)
	mockRepo.On("Update", mock.Anything, uint(999), mock.AnythingOfType("map[string]interface {}")).Return(nil, assert.AnError)

	req := httptest.NewRequest(http.MethodPut, "/news/999", bytes.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/news/:id")
	c.SetParamNames("id")
	c.SetParamValues("999")

	err := h.Update(c)
	require.Error(t, err)
	he, ok := err.(*echo.HTTPError)
	require.True(t, ok)
	assert.Equal(t, http.StatusNotFound, he.Code)
	assert.Equal(t, "NOT_FOUND", he.Message)
	mockRepo.AssertExpectations(t)
}
