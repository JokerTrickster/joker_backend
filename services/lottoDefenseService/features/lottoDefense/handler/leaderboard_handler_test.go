package handler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/JokerTrickster/joker_backend/services/lottoDefenseService/features/lottoDefense/model/response"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type mockLeaderboardUseCase struct {
	mock.Mock
}

func (m *mockLeaderboardUseCase) GetLeaderboard(ctx context.Context, limit int) (*response.LeaderboardResponse, error) {
	args := m.Called(ctx, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*response.LeaderboardResponse), args.Error(1)
}

func TestLeaderboardHandler_GetLeaderboard_Success(t *testing.T) {
	t.Log("GetLeaderboard: success -> 200")
	e := setupTestEcho()
	mockUC := new(mockLeaderboardUseCase)
	h := &LeaderboardHandler{uc: mockUC}

	mockUC.On("GetLeaderboard", mock.Anything, 10).
		Return(&response.LeaderboardResponse{
			Entries: []response.LeaderboardEntry{
				{Rank: 1, UserID: 1, Name: "Alice", Score: 1000},
				{Rank: 2, UserID: 2, Name: "Bob", Score: 900},
			},
		}, nil)

	req := httptest.NewRequest(http.MethodGet, "/leaderboard", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.GetLeaderboard(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), "Alice")
	mockUC.AssertExpectations(t)
}

func TestLeaderboardHandler_GetLeaderboard_WithLimitParam(t *testing.T) {
	t.Log("GetLeaderboard: with limit param -> 200")
	e := setupTestEcho()
	mockUC := new(mockLeaderboardUseCase)
	h := &LeaderboardHandler{uc: mockUC}

	mockUC.On("GetLeaderboard", mock.Anything, 25).
		Return(&response.LeaderboardResponse{Entries: []response.LeaderboardEntry{}}, nil)

	req := httptest.NewRequest(http.MethodGet, "/leaderboard?limit=25", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.GetLeaderboard(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	mockUC.AssertExpectations(t)
}

func TestLeaderboardHandler_GetLeaderboard_InvalidLimitIgnored(t *testing.T) {
	t.Log("GetLeaderboard: invalid limit (negative) -> uses default 10")
	e := setupTestEcho()
	mockUC := new(mockLeaderboardUseCase)
	h := &LeaderboardHandler{uc: mockUC}

	mockUC.On("GetLeaderboard", mock.Anything, 10).
		Return(&response.LeaderboardResponse{Entries: []response.LeaderboardEntry{}}, nil)

	req := httptest.NewRequest(http.MethodGet, "/leaderboard?limit=-5", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.GetLeaderboard(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	mockUC.AssertExpectations(t)
}

func TestLeaderboardHandler_GetLeaderboard_Over100UsesDefault(t *testing.T) {
	t.Log("GetLeaderboard: limit>100 uses default 10")
	e := setupTestEcho()
	mockUC := new(mockLeaderboardUseCase)
	h := &LeaderboardHandler{uc: mockUC}

	mockUC.On("GetLeaderboard", mock.Anything, 10).
		Return(&response.LeaderboardResponse{Entries: []response.LeaderboardEntry{}}, nil)

	req := httptest.NewRequest(http.MethodGet, "/leaderboard?limit=999", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.GetLeaderboard(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	mockUC.AssertExpectations(t)
}

func TestLeaderboardHandler_GetLeaderboard_NonNumericLimitIgnored(t *testing.T) {
	t.Log("GetLeaderboard: non-numeric limit -> uses default 10")
	e := setupTestEcho()
	mockUC := new(mockLeaderboardUseCase)
	h := &LeaderboardHandler{uc: mockUC}

	mockUC.On("GetLeaderboard", mock.Anything, 10).
		Return(&response.LeaderboardResponse{Entries: []response.LeaderboardEntry{}}, nil)

	req := httptest.NewRequest(http.MethodGet, "/leaderboard?limit=abc", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.GetLeaderboard(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	mockUC.AssertExpectations(t)
}

func TestLeaderboardHandler_GetLeaderboard_UseCaseError(t *testing.T) {
	t.Log("GetLeaderboard: usecase error -> 500")
	e := setupTestEcho()
	mockUC := new(mockLeaderboardUseCase)
	h := &LeaderboardHandler{uc: mockUC}

	mockUC.On("GetLeaderboard", mock.Anything, 10).
		Return(nil, assert.AnError)

	req := httptest.NewRequest(http.MethodGet, "/leaderboard", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.GetLeaderboard(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	mockUC.AssertExpectations(t)
}
