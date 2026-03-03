package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/JokerTrickster/joker_backend/services/lottoDefenseService/features/lottoDefense/model/entity"
	_interface "github.com/JokerTrickster/joker_backend/services/lottoDefenseService/features/lottoDefense/model/interface"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type mockGameRoundRepository struct {
	mock.Mock
}

func (m *mockGameRoundRepository) Create(ctx context.Context, round *entity.GameRound) error {
	args := m.Called(ctx, round)
	return args.Error(0)
}

func (m *mockGameRoundRepository) GetByID(ctx context.Context, id uint) (*entity.GameRound, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.GameRound), args.Error(1)
}

func (m *mockGameRoundRepository) GetByIDAndUser(ctx context.Context, id, userID uint) (*entity.GameRound, error) {
	args := m.Called(ctx, id, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.GameRound), args.Error(1)
}

func (m *mockGameRoundRepository) Update(ctx context.Context, round *entity.GameRound) error {
	args := m.Called(ctx, round)
	return args.Error(0)
}

func (m *mockGameRoundRepository) ListByUserID(ctx context.Context, userID uint, limit int) ([]entity.GameRound, error) {
	args := m.Called(ctx, userID, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]entity.GameRound), args.Error(1)
}

func (m *mockGameRoundRepository) TopScores(ctx context.Context, limit int) ([]entity.GameRound, error) {
	args := m.Called(ctx, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]entity.GameRound), args.Error(1)
}

func (m *mockGameRoundRepository) Leaderboard(ctx context.Context, limit int) ([]_interface.LeaderboardRow, error) {
	args := m.Called(ctx, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]_interface.LeaderboardRow), args.Error(1)
}

func TestLeaderboardUseCase_GetLeaderboard_Success(t *testing.T) {
	t.Log("GetLeaderboard: success")
	mockRepo := new(mockGameRoundRepository)
	uc := NewLeaderboardUseCase(mockRepo)
	ctx := context.Background()

	mockRepo.On("Leaderboard", ctx, 10).
		Return([]_interface.LeaderboardRow{
			{UserID: 1, Name: "Alice", Score: 1000},
			{UserID: 2, Name: "Bob", Score: 900},
		}, nil)

	resp, err := uc.GetLeaderboard(ctx, 10)
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Len(t, resp.Entries, 2)
	assert.Equal(t, 1, resp.Entries[0].Rank)
	assert.Equal(t, uint(1), resp.Entries[0].UserID)
	assert.Equal(t, "Alice", resp.Entries[0].Name)
	assert.Equal(t, uint(1000), resp.Entries[0].Score)
	assert.Equal(t, 2, resp.Entries[1].Rank)
	mockRepo.AssertExpectations(t)
}

func TestLeaderboardUseCase_GetLeaderboard_EmptyResult(t *testing.T) {
	t.Log("GetLeaderboard: empty result")
	mockRepo := new(mockGameRoundRepository)
	uc := NewLeaderboardUseCase(mockRepo)
	ctx := context.Background()

	mockRepo.On("Leaderboard", ctx, 10).
		Return([]_interface.LeaderboardRow{}, nil)

	resp, err := uc.GetLeaderboard(ctx, 10)
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Empty(t, resp.Entries)
	mockRepo.AssertExpectations(t)
}

func TestLeaderboardUseCase_GetLeaderboard_DefaultLimitClamping(t *testing.T) {
	t.Log("GetLeaderboard: default limit clamping when limit <= 0")
	mockRepo := new(mockGameRoundRepository)
	uc := NewLeaderboardUseCase(mockRepo)
	ctx := context.Background()

	mockRepo.On("Leaderboard", ctx, 10).
		Return([]_interface.LeaderboardRow{}, nil)

	resp, err := uc.GetLeaderboard(ctx, 0)
	require.NoError(t, err)
	require.NotNil(t, resp)
	mockRepo.AssertExpectations(t)
}

func TestLeaderboardUseCase_GetLeaderboard_LimitOver100Clamped(t *testing.T) {
	t.Log("GetLeaderboard: limit > 100 clamped to 100")
	mockRepo := new(mockGameRoundRepository)
	uc := NewLeaderboardUseCase(mockRepo)
	ctx := context.Background()

	mockRepo.On("Leaderboard", ctx, 100).
		Return([]_interface.LeaderboardRow{}, nil)

	resp, err := uc.GetLeaderboard(ctx, 200)
	require.NoError(t, err)
	require.NotNil(t, resp)
	mockRepo.AssertExpectations(t)
}

func TestLeaderboardUseCase_GetLeaderboard_RepositoryError(t *testing.T) {
	t.Log("GetLeaderboard: repository error")
	mockRepo := new(mockGameRoundRepository)
	uc := NewLeaderboardUseCase(mockRepo)
	ctx := context.Background()

	mockRepo.On("Leaderboard", ctx, 10).
		Return(nil, errors.New("db error"))

	resp, err := uc.GetLeaderboard(ctx, 10)
	assert.Error(t, err)
	assert.Nil(t, resp)
	mockRepo.AssertExpectations(t)
}
