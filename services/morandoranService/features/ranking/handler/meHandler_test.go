package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/JokerTrickster/joker_backend/services/morandoranService/features/ranking/model/entity"
	"github.com/JokerTrickster/joker_backend/services/morandoranService/features/ranking/usecase"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/mock"
)

func TestMeHandler_MyRanking_Success(t *testing.T) {
	t.Log("MyRanking: valid request -> 200 OK")
	e := setupRankingTestEcho()
	mockRepo := new(mockRankingRepository)
	uc := usecase.NewMeUseCase(mockRepo, rankingDefaultTimeout)
	h := NewMeHandler(uc)

	ranking := &entity.Ranking{
		ID: 1, UserID: rankingTestUserID, GameType: "puzzle", Nickname: "Me", ClearTimeMs: 4000,
	}
	mockRepo.On("FindByUserAndGame", mock.Anything, rankingTestUserID, "puzzle").Return(ranking, nil)
	mockRepo.On("GetRank", mock.Anything, "puzzle", uint(4000)).Return(3, nil)

	req := httptest.NewRequest(http.MethodGet, "/ranking/puzzle/me", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/ranking/:gameType/me")
	c.SetParamNames("gameType")
	c.SetParamValues("puzzle")
	setupRankingAuthContext(c)

	err := h.MyRanking(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), "3")
	assert.Contains(t, rec.Body.String(), "Me")
	mockRepo.AssertExpectations(t)
}

func TestMeHandler_MyRanking_Unauthorized(t *testing.T) {
	t.Log("MyRanking: no userID -> 401 Unauthorized")
	e := setupRankingTestEcho()
	mockRepo := new(mockRankingRepository)
	uc := usecase.NewMeUseCase(mockRepo, rankingDefaultTimeout)
	h := NewMeHandler(uc)

	req := httptest.NewRequest(http.MethodGet, "/ranking/puzzle/me", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/ranking/:gameType/me")
	c.SetParamNames("gameType")
	c.SetParamValues("puzzle")
	// Do NOT set userID

	err := h.MyRanking(c)
	require.Error(t, err)
	he, ok := err.(*echo.HTTPError)
	require.True(t, ok)
	assert.Equal(t, http.StatusUnauthorized, he.Code)
	mockRepo.AssertNotCalled(t, "FindByUserAndGame")
}

func TestMeHandler_MyRanking_NotFound(t *testing.T) {
	t.Log("MyRanking: not found -> 404")
	e := setupRankingTestEcho()
	mockRepo := new(mockRankingRepository)
	uc := usecase.NewMeUseCase(mockRepo, rankingDefaultTimeout)
	h := NewMeHandler(uc)

	mockRepo.On("FindByUserAndGame", mock.Anything, rankingTestUserID, "puzzle").Return(nil, assert.AnError)

	req := httptest.NewRequest(http.MethodGet, "/ranking/puzzle/me", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/ranking/:gameType/me")
	c.SetParamNames("gameType")
	c.SetParamValues("puzzle")
	setupRankingAuthContext(c)

	err := h.MyRanking(c)
	require.Error(t, err)
	he, ok := err.(*echo.HTTPError)
	require.True(t, ok)
	assert.Equal(t, http.StatusNotFound, he.Code)
	mockRepo.AssertExpectations(t)
}
