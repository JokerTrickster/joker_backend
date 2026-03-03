package usecase

import (
	"context"
	"testing"
	"time"

	"github.com/JokerTrickster/joker_backend/services/lottoDefenseService/features/towerDefense/model/entity"
	"github.com/JokerTrickster/joker_backend/services/lottoDefenseService/features/towerDefense/model/request"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type mockTDGameRepository struct {
	mock.Mock
}

func (m *mockTDGameRepository) Create(ctx context.Context, result *entity.TDGameResult) error {
	args := m.Called(ctx, result)
	return args.Error(0)
}

func (m *mockTDGameRepository) GetByID(ctx context.Context, id uint) (*entity.TDGameResult, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.TDGameResult), args.Error(1)
}

func (m *mockTDGameRepository) GetHistory(ctx context.Context, userID uint, gameMode string, limit, offset int) ([]entity.TDGameResult, int64, error) {
	args := m.Called(ctx, userID, gameMode, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]entity.TDGameResult), args.Get(1).(int64), args.Error(2)
}

func (m *mockTDGameRepository) GetHighestRound(ctx context.Context, userID uint, gameMode string) (uint, error) {
	args := m.Called(ctx, userID, gameMode)
	return args.Get(0).(uint), args.Error(1)
}

func (m *mockTDGameRepository) GetWeeklyRankings(ctx context.Context, gameMode string, limit int) ([]entity.TDGameResult, error) {
	args := m.Called(ctx, gameMode, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]entity.TDGameResult), args.Error(1)
}

func TestTDGameUseCase_SaveSingleResult_Success(t *testing.T) {
	t.Log("SaveSingleResult: success - updates stats")
	mockGameRepo := new(mockTDGameRepository)
	mockUserRepo := new(mockTDUserRepository)
	uc := NewTDGameUseCase(mockGameRepo, mockUserRepo, 5*time.Second)
	ctx := context.Background()
	userID := uint(1)
	req := &request.SaveGameResultRequest{
		GameMode:       "single",
		RoundsReached:  10,
		MonstersKilled: 50,
		GoldEarned:     100,
		Result:         "victory",
	}

	mockGameRepo.On("Create", ctx, mock.AnythingOfType("*entity.TDGameResult")).Run(func(args mock.Arguments) {
		r := args.Get(1).(*entity.TDGameResult)
		r.ID = 1
	}).Return(nil)
	mockUserRepo.On("GetStats", ctx, userID).Return(&entity.TDUserStats{
		UserID:             userID,
		SingleHighestRound: 5,
		SingleTotalGames:   2,
		SingleTotalKills:   20,
	}, nil)
	mockUserRepo.On("UpdateStats", ctx, mock.AnythingOfType("*entity.TDUserStats")).Return(nil)

	resp, err := uc.SaveSingleResult(ctx, userID, req)
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, uint(1), resp.GameID)
	assert.Equal(t, uint(10), resp.NewHighestRound)
	mockGameRepo.AssertExpectations(t)
	mockUserRepo.AssertExpectations(t)
}

func TestTDGameUseCase_GetGameHistory_Success(t *testing.T) {
	t.Log("GetGameHistory: success")
	mockGameRepo := new(mockTDGameRepository)
	mockUserRepo := new(mockTDUserRepository)
	uc := NewTDGameUseCase(mockGameRepo, mockUserRepo, 5*time.Second)
	ctx := context.Background()
	userID := uint(1)
	req := &request.GameHistoryRequest{GameMode: "single", Limit: 10, Offset: 0}

	mockGameRepo.On("GetHistory", ctx, userID, "single", 10, 0).
		Return([]entity.TDGameResult{
			{ID: 1, GameMode: "single", RoundsReached: 10, MonstersKilled: 50, GoldEarned: 100, Result: "victory"},
		}, int64(1), nil)

	resp, err := uc.GetGameHistory(ctx, userID, req)
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, int64(1), resp.Total)
	assert.Len(t, resp.Games, 1)
	mockGameRepo.AssertExpectations(t)
}

func TestTDGameUseCase_GetUserStats_Success(t *testing.T) {
	t.Log("GetUserStats: success")
	mockGameRepo := new(mockTDGameRepository)
	mockUserRepo := new(mockTDUserRepository)
	uc := NewTDGameUseCase(mockGameRepo, mockUserRepo, 5*time.Second)
	ctx := context.Background()
	userID := uint(1)

	mockUserRepo.On("GetStats", ctx, userID).Return(&entity.TDUserStats{
		UserID:             userID,
		SingleHighestRound: 10,
		SingleTotalGames:   5,
		SingleTotalKills:   100,
		CoopHighestRound:   8,
		CoopTotalGames:    3,
		CoopTotalKills:    60,
		CoopWins:          2,
		TotalGoldEarned:   500,
		CurrentGold:       200,
	}, nil)

	resp, err := uc.GetUserStats(ctx, userID)
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, uint(10), resp.Single.HighestRound)
	assert.Equal(t, uint(5), resp.Single.TotalGames)
	mockUserRepo.AssertExpectations(t)
}

func TestTDGameUseCase_GetWeeklyRankings_Success(t *testing.T) {
	t.Log("GetWeeklyRankings: success")
	mockGameRepo := new(mockTDGameRepository)
	mockUserRepo := new(mockTDUserRepository)
	uc := NewTDGameUseCase(mockGameRepo, mockUserRepo, 5*time.Second)
	ctx := context.Background()

	mockGameRepo.On("GetWeeklyRankings", ctx, "single", 10).
		Return([]entity.TDGameResult{
			{ID: 1, UserID: 1, RoundsReached: 15, GameMode: "single", Result: "victory", User: &entity.TDUser{Username: "player1"}},
		}, nil)

	resp, err := uc.GetWeeklyRankings(ctx, "single")
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, "single", resp.GameMode)
	mockGameRepo.AssertExpectations(t)
}
