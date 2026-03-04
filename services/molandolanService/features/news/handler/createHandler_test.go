package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/JokerTrickster/joker_backend/services/molandolanService/features/news/model/entity"
	"github.com/JokerTrickster/joker_backend/services/molandolanService/features/news/model/request"
	"github.com/JokerTrickster/joker_backend/services/molandolanService/features/news/model/response"
	"github.com/JokerTrickster/joker_backend/services/molandolanService/features/news/usecase"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/mock"
)

func TestCreateHandler_Create_Success(t *testing.T) {
	t.Log("Create: valid request -> 201 Created")
	e := setupNewsTestEcho()
	mockRepo := new(mockNewsRepository)
	uc := usecase.NewCreateUseCase(mockRepo, defaultTimeout)
	h := NewCreateHandler(uc)

	reqBody := &request.ReqCreateNews{
		Title:     "Test News",
		Summary:   "Summary",
		Content:   "Content body",
		Thumbnail: "https://example.com/img.png",
		Category:  "announcement",
		Date:      "2024-01-15",
	}
	body := newsMustJSON(t, reqBody)
	mockRepo.On("Create", mock.Anything, mock.AnythingOfType("*entity.News")).Return(&entity.News{
		ID: 1, Title: reqBody.Title, Summary: reqBody.Summary, Content: reqBody.Content,
		Thumbnail: reqBody.Thumbnail, Category: reqBody.Category, Date: reqBody.Date,
	}, nil)

	req := httptest.NewRequest(http.MethodPost, "/news", bytes.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.Create(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusCreated, rec.Code)

	var res response.ResNewsDetail
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &res))
	assert.Equal(t, "news-001", res.ID)
	assert.Equal(t, reqBody.Title, res.Title)
	t.Logf("Create response: %+v", res)
	mockRepo.AssertExpectations(t)
}

func TestCreateHandler_Create_BadRequest_BindError(t *testing.T) {
	t.Log("Create: invalid JSON -> 400 Bad Request")
	e := setupNewsTestEcho()
	mockRepo := new(mockNewsRepository)
	uc := usecase.NewCreateUseCase(mockRepo, defaultTimeout)
	h := NewCreateHandler(uc)

	req := httptest.NewRequest(http.MethodPost, "/news", bytes.NewReader([]byte("invalid json")))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.Create(c)
	require.Error(t, err)
	he, ok := err.(*echo.HTTPError)
	require.True(t, ok)
	assert.Equal(t, http.StatusBadRequest, he.Code)
	assert.Equal(t, "BAD_REQUEST", he.Message)
	mockRepo.AssertNotCalled(t, "Create")
}

func TestCreateHandler_Create_BadRequest_ValidationError(t *testing.T) {
	t.Log("Create: missing required fields -> 400 Bad Request")
	e := setupNewsTestEcho()
	mockRepo := new(mockNewsRepository)
	uc := usecase.NewCreateUseCase(mockRepo, defaultTimeout)
	h := NewCreateHandler(uc)

	reqBody := map[string]interface{}{"title": ""}
	body := newsMustJSON(t, reqBody)
	req := httptest.NewRequest(http.MethodPost, "/news", bytes.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.Create(c)
	require.Error(t, err)
	he, ok := err.(*echo.HTTPError)
	require.True(t, ok)
	assert.Equal(t, http.StatusBadRequest, he.Code)
	mockRepo.AssertNotCalled(t, "Create")
}

func TestCreateHandler_Create_InternalError(t *testing.T) {
	t.Log("Create: usecase error -> 500 Internal Server Error")
	e := setupNewsTestEcho()
	mockRepo := new(mockNewsRepository)
	uc := usecase.NewCreateUseCase(mockRepo, defaultTimeout)
	h := NewCreateHandler(uc)

	reqBody := &request.ReqCreateNews{
		Title: "Test", Summary: "", Content: "Content", Thumbnail: "", Category: "cat", Date: "2024-01-15",
	}
	body := newsMustJSON(t, reqBody)
	mockRepo.On("Create", mock.Anything, mock.AnythingOfType("*entity.News")).Return(nil, assert.AnError)

	req := httptest.NewRequest(http.MethodPost, "/news", bytes.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.Create(c)
	require.Error(t, err)
	he, ok := err.(*echo.HTTPError)
	require.True(t, ok)
	assert.Equal(t, http.StatusInternalServerError, he.Code)
	assert.Equal(t, "INTERNAL_ERROR", he.Message)
	mockRepo.AssertExpectations(t)
}
