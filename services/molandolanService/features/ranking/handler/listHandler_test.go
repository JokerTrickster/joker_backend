package handler

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/JokerTrickster/joker_backend/services/molandolanService/features/ranking/model/entity"
	"github.com/JokerTrickster/joker_backend/services/molandolanService/features/ranking/usecase"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestListHandler_List_Success(t *testing.T) {
	t.Log("List: valid request -> 200 OK")
	e := setupRankingTestEcho()
	mockRepo := new(mockRankingRepository)
	uc := usecase.NewListUseCase(mockRepo, rankingDefaultTimeout)
	h := NewListHandler(uc)

	items := []entity.Ranking{
		{ID: 1, UserID: 1, GameType: "puzzle", Nickname: "Alice", ClearTimeMs: 3000},
		{ID: 2, UserID: 2, GameType: "puzzle", Nickname: "Bob", ClearTimeMs: 5000},
	}
	mockRepo.On("List", mock.Anything, "puzzle", 1, 5).Return(items, int64(2), nil)

	req := httptest.NewRequest(http.MethodGet, "/ranking/puzzle?limit=5", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/ranking/:gameType")
	c.SetParamNames("gameType")
	c.SetParamValues("puzzle")

	err := h.List(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), "Alice")
	assert.Contains(t, rec.Body.String(), "Bob")
	mockRepo.AssertExpectations(t)
}

func TestListHandler_List_UseCaseError(t *testing.T) {
	t.Log("List: usecase error -> 500")
	e := setupRankingTestEcho()
	mockRepo := new(mockRankingRepository)
	uc := usecase.NewListUseCase(mockRepo, rankingDefaultTimeout)
	h := NewListHandler(uc)

	mockRepo.On("List", mock.Anything, "puzzle", 1, 5).Return(nil, int64(0), assert.AnError)

	req := httptest.NewRequest(http.MethodGet, "/ranking/puzzle?limit=5", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/ranking/:gameType")
	c.SetParamNames("gameType")
	c.SetParamValues("puzzle")

	err := h.List(c)
	require.Error(t, err)
	he, ok := err.(*echo.HTTPError)
	require.True(t, ok)
	assert.Equal(t, http.StatusInternalServerError, he.Code)
	mockRepo.AssertExpectations(t)
}

func TestListHandler_List_BadRequest_BindError(t *testing.T) {
	t.Log("List: invalid bind (e.g. POST with malformed body) -> 400")
	e := setupRankingTestEcho()
	mockRepo := new(mockRankingRepository)
	uc := usecase.NewListUseCase(mockRepo, rankingDefaultTimeout)
	h := NewListHandler(uc)

	req := httptest.NewRequest(http.MethodPost, "/ranking/puzzle", bytes.NewReader([]byte("invalid json")))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/ranking/:gameType")
	c.SetParamNames("gameType")
	c.SetParamValues("puzzle")

	err := h.List(c)
	require.Error(t, err)
	he, ok := err.(*echo.HTTPError)
	require.True(t, ok)
	assert.Equal(t, http.StatusBadRequest, he.Code)
	mockRepo.AssertNotCalled(t, "List")
}

func TestListHandler_List_DefaultLimitPagination(t *testing.T) {
	t.Log("List: limit=0 uses default 5")
	e := setupRankingTestEcho()
	mockRepo := new(mockRankingRepository)
	uc := usecase.NewListUseCase(mockRepo, rankingDefaultTimeout)
	h := NewListHandler(uc)

	mockRepo.On("List", mock.Anything, "puzzle", 1, 5).Return([]entity.Ranking{}, int64(0), nil)

	req := httptest.NewRequest(http.MethodGet, "/ranking/puzzle?limit=0", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/ranking/:gameType")
	c.SetParamNames("gameType")
	c.SetParamValues("puzzle")

	err := h.List(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	mockRepo.AssertExpectations(t)
}

func TestListHandler_List_MaxLimitClamped(t *testing.T) {
	t.Log("List: limit>100 clamped to 100")
	e := setupRankingTestEcho()
	mockRepo := new(mockRankingRepository)
	uc := usecase.NewListUseCase(mockRepo, rankingDefaultTimeout)
	h := NewListHandler(uc)

	mockRepo.On("List", mock.Anything, "puzzle", 1, 100).Return([]entity.Ranking{}, int64(0), nil)

	req := httptest.NewRequest(http.MethodGet, "/ranking/puzzle?limit=150", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/ranking/:gameType")
	c.SetParamNames("gameType")
	c.SetParamValues("puzzle")

	err := h.List(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	mockRepo.AssertExpectations(t)
}
