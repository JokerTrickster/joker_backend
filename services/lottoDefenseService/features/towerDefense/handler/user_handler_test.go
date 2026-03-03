package handler

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/JokerTrickster/joker_backend/services/lottoDefenseService/features/towerDefense/model/response"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestTDUserHandler_GetMe_Success(t *testing.T) {
	t.Log("GetMe: set auth context, mock GetUserInfo returns response -> 200 with success=true")
	e := setupTDTestEcho()
	mockAuthUC := new(mockTDAuthUseCase)
	mockGameUC := new(mockTDGameUseCase)
	h := &TDUserHandler{authUC: mockAuthUC, gameUC: mockGameUC}

	mockAuthUC.On("GetUserInfo", mock.Anything, tdTestUserID).
		Return(&response.UserInfoResponse{
			User:  &response.UserData{ID: 1, Username: "testuser", Email: "test@example.com"},
			Stats: &response.UserStatsData{},
		}, nil)

	req := httptest.NewRequest(http.MethodGet, "/users/me", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	setupTDAuthContext(c)

	err := h.GetMe(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), "\"success\":true")
	assert.Contains(t, rec.Body.String(), "\"data\"")
	mockAuthUC.AssertExpectations(t)
	mockGameUC.AssertNotCalled(t, "GetUserStats")
}

func TestTDUserHandler_GetMe_NoAuth(t *testing.T) {
	t.Log("GetMe: don't set userID in context -> error")
	e := setupTDTestEcho()
	mockAuthUC := new(mockTDAuthUseCase)
	mockGameUC := new(mockTDGameUseCase)
	h := &TDUserHandler{authUC: mockAuthUC, gameUC: mockGameUC}

	req := httptest.NewRequest(http.MethodGet, "/users/me", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.GetMe(c)
	require.Error(t, err)
	he, ok := err.(*echo.HTTPError)
	require.True(t, ok)
	assert.Equal(t, http.StatusUnauthorized, he.Code)
	mockAuthUC.AssertNotCalled(t, "GetUserInfo")
}

func TestTDUserHandler_GetMe_UseCaseError(t *testing.T) {
	t.Log("GetMe: mock GetUserInfo returns error -> 500")
	e := setupTDTestEcho()
	mockAuthUC := new(mockTDAuthUseCase)
	mockGameUC := new(mockTDGameUseCase)
	h := &TDUserHandler{authUC: mockAuthUC, gameUC: mockGameUC}

	mockAuthUC.On("GetUserInfo", mock.Anything, tdTestUserID).
		Return(nil, errors.New("database error"))

	req := httptest.NewRequest(http.MethodGet, "/users/me", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	setupTDAuthContext(c)

	err := h.GetMe(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	assert.Contains(t, rec.Body.String(), "\"error\":\"database error\"")
	mockAuthUC.AssertExpectations(t)
}

func TestTDUserHandler_GetStats_Success(t *testing.T) {
	t.Log("GetStats: set auth context, mock GetUserStats returns response -> 200")
	e := setupTDTestEcho()
	mockAuthUC := new(mockTDAuthUseCase)
	mockGameUC := new(mockTDGameUseCase)
	h := &TDUserHandler{authUC: mockAuthUC, gameUC: mockGameUC}

	mockGameUC.On("GetUserStats", mock.Anything, tdTestUserID).
		Return(&response.UserStatsResponse{
			Single: response.SingleStats{HighestRound: 10, TotalGames: 5, TotalKills: 100},
			Coop:   response.CoopStats{HighestRound: 5, TotalGames: 3, Wins: 2},
			Gold:   response.GoldStats{TotalEarned: 500, Current: 100},
		}, nil)

	req := httptest.NewRequest(http.MethodGet, "/users/me/stats", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	setupTDAuthContext(c)

	err := h.GetStats(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), "\"success\":true")
	assert.Contains(t, rec.Body.String(), "\"data\"")
	mockGameUC.AssertExpectations(t)
	mockAuthUC.AssertNotCalled(t, "GetUserInfo")
}

func TestTDUserHandler_GetStats_NoAuth(t *testing.T) {
	t.Log("GetStats: don't set userID -> error")
	e := setupTDTestEcho()
	mockAuthUC := new(mockTDAuthUseCase)
	mockGameUC := new(mockTDGameUseCase)
	h := &TDUserHandler{authUC: mockAuthUC, gameUC: mockGameUC}

	req := httptest.NewRequest(http.MethodGet, "/users/me/stats", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.GetStats(c)
	require.Error(t, err)
	he, ok := err.(*echo.HTTPError)
	require.True(t, ok)
	assert.Equal(t, http.StatusUnauthorized, he.Code)
	mockGameUC.AssertNotCalled(t, "GetUserStats")
}

func TestTDUserHandler_GetStats_UseCaseError(t *testing.T) {
	t.Log("GetStats: mock GetUserStats returns error -> 500")
	e := setupTDTestEcho()
	mockAuthUC := new(mockTDAuthUseCase)
	mockGameUC := new(mockTDGameUseCase)
	h := &TDUserHandler{authUC: mockAuthUC, gameUC: mockGameUC}

	mockGameUC.On("GetUserStats", mock.Anything, tdTestUserID).
		Return(nil, errors.New("stats lookup failed"))

	req := httptest.NewRequest(http.MethodGet, "/users/me/stats", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	setupTDAuthContext(c)

	err := h.GetStats(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	assert.Contains(t, rec.Body.String(), "\"error\":\"stats lookup failed\"")
	mockGameUC.AssertExpectations(t)
}
