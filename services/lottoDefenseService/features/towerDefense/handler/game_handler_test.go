package handler

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/JokerTrickster/joker_backend/services/lottoDefenseService/features/towerDefense/model/request"
	"github.com/JokerTrickster/joker_backend/services/lottoDefenseService/features/towerDefense/model/response"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type mockTDGameUseCase struct {
	mock.Mock
}

func (m *mockTDGameUseCase) SaveSingleResult(ctx context.Context, userID uint, req *request.SaveGameResultRequest) (*response.GameResultResponse, error) {
	args := m.Called(ctx, userID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*response.GameResultResponse), args.Error(1)
}

func (m *mockTDGameUseCase) GetGameHistory(ctx context.Context, userID uint, req *request.GameHistoryRequest) (*response.GameHistoryResponse, error) {
	args := m.Called(ctx, userID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*response.GameHistoryResponse), args.Error(1)
}

func (m *mockTDGameUseCase) GetUserStats(ctx context.Context, userID uint) (*response.UserStatsResponse, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*response.UserStatsResponse), args.Error(1)
}

func (m *mockTDGameUseCase) GetWeeklyRankings(ctx context.Context, gameMode string) (*response.RankingResponse, error) {
	args := m.Called(ctx, gameMode)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*response.RankingResponse), args.Error(1)
}

func TestTDGameHandler_SaveSingleResult_Success(t *testing.T) {
	t.Log("SaveSingleResult: success -> 201")
	e := setupTDTestEcho()
	mockUC := new(mockTDGameUseCase)
	h := &TDGameHandler{uc: mockUC}

	body := tdMustJSON(t, &request.SaveGameResultRequest{
		GameMode:       "single",
		RoundsReached:  10,
		MonstersKilled: 50,
		GoldEarned:     100,
		Result:         "victory",
	})
	mockUC.On("SaveSingleResult", mock.Anything, tdTestUserID, mock.AnythingOfType("*request.SaveGameResultRequest")).
		Return(&response.GameResultResponse{GameID: 1, NewHighestRound: 10, Rewards: []response.Reward{}}, nil)

	req := httptest.NewRequest(http.MethodPost, "/game/single/result", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	setupTDAuthContext(c)

	err := h.SaveSingleResult(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusCreated, rec.Code)
	mockUC.AssertExpectations(t)
}

func TestTDGameHandler_SaveSingleResult_NoUserID(t *testing.T) {
	t.Log("SaveSingleResult: no userID -> 401")
	e := setupTDTestEcho()
	mockUC := new(mockTDGameUseCase)
	h := &TDGameHandler{uc: mockUC}

	body := tdMustJSON(t, &request.SaveGameResultRequest{
		GameMode:       "single",
		RoundsReached:  10,
		MonstersKilled: 50,
		GoldEarned:     100,
		Result:         "victory",
	})
	req := httptest.NewRequest(http.MethodPost, "/game/single/result", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.SaveSingleResult(c)
	require.Error(t, err)
	he, ok := err.(*echo.HTTPError)
	require.True(t, ok)
	assert.Equal(t, http.StatusUnauthorized, he.Code)
	mockUC.AssertNotCalled(t, "SaveSingleResult")
}

func TestTDGameHandler_SaveSingleResult_ValidationError(t *testing.T) {
	t.Log("SaveSingleResult: validation error -> 400")
	e := setupTDTestEcho()
	mockUC := new(mockTDGameUseCase)
	h := &TDGameHandler{uc: mockUC}

	body := tdMustJSON(t, map[string]interface{}{
		"game_mode": "invalid",
		"rounds_reached": 0,
	})
	req := httptest.NewRequest(http.MethodPost, "/game/single/result", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	setupTDAuthContext(c)

	err := h.SaveSingleResult(c)
	require.Error(t, err)
	he, ok := err.(*echo.HTTPError)
	require.True(t, ok)
	assert.Equal(t, http.StatusBadRequest, he.Code)
	mockUC.AssertNotCalled(t, "SaveSingleResult")
}

func TestTDGameHandler_GetHistory_Success(t *testing.T) {
	t.Log("GetHistory: success -> 200")
	e := setupTDTestEcho()
	mockUC := new(mockTDGameUseCase)
	h := &TDGameHandler{uc: mockUC}

	mockUC.On("GetGameHistory", mock.Anything, tdTestUserID, mock.AnythingOfType("*request.GameHistoryRequest")).
		Return(&response.GameHistoryResponse{Total: 0, Games: []response.GameHistoryItem{}}, nil)

	req := httptest.NewRequest(http.MethodGet, "/game/history", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	setupTDAuthContext(c)

	err := h.GetHistory(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	mockUC.AssertExpectations(t)
}

func TestTDGameHandler_GetHistory_NoUserID(t *testing.T) {
	t.Log("GetHistory: no userID -> 401")
	e := setupTDTestEcho()
	mockUC := new(mockTDGameUseCase)
	h := &TDGameHandler{uc: mockUC}

	req := httptest.NewRequest(http.MethodGet, "/game/history", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.GetHistory(c)
	require.Error(t, err)
	he, ok := err.(*echo.HTTPError)
	require.True(t, ok)
	assert.Equal(t, http.StatusUnauthorized, he.Code)
	mockUC.AssertNotCalled(t, "GetGameHistory")
}

func TestTDGameHandler_GetRankings_Success(t *testing.T) {
	t.Log("GetRankings: success -> 200")
	e := setupTDTestEcho()
	mockUC := new(mockTDGameUseCase)
	h := &TDGameHandler{uc: mockUC}

	mockUC.On("GetWeeklyRankings", mock.Anything, "single").
		Return(&response.RankingResponse{GameMode: "single", Rankings: []response.RankingItem{}}, nil)

	req := httptest.NewRequest(http.MethodGet, "/rankings/single", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/rankings/:mode")
	c.SetParamNames("mode")
	c.SetParamValues("single")

	err := h.GetRankings(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	mockUC.AssertExpectations(t)
}

func TestTDGameHandler_GetRankings_InvalidMode(t *testing.T) {
	t.Log("GetRankings: invalid mode -> 400")
	e := setupTDTestEcho()
	mockUC := new(mockTDGameUseCase)
	h := &TDGameHandler{uc: mockUC}

	req := httptest.NewRequest(http.MethodGet, "/rankings/invalid", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/rankings/:mode")
	c.SetParamNames("mode")
	c.SetParamValues("invalid")

	err := h.GetRankings(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
	mockUC.AssertNotCalled(t, "GetWeeklyRankings")
}
