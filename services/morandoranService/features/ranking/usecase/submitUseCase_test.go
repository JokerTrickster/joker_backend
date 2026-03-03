package usecase

import (
	"context"
	"testing"
	"time"

	"github.com/JokerTrickster/joker_backend/services/morandoranService/features/ranking/model/entity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

type mockSubmitRankingRepo struct {
	mock.Mock
}

func (m *mockSubmitRankingRepo) List(ctx context.Context, gameType string, page, limit int) ([]entity.Ranking, int64, error) {
	args := m.Called(ctx, gameType, page, limit)
	if args.Get(0) == nil {
		return nil, 0, args.Error(2)
	}
	return args.Get(0).([]entity.Ranking), args.Get(1).(int64), args.Error(2)
}

func (m *mockSubmitRankingRepo) FindByUserAndGame(ctx context.Context, userID uint, gameType string) (*entity.Ranking, error) {
	args := m.Called(ctx, userID, gameType)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Ranking), args.Error(1)
}

func (m *mockSubmitRankingRepo) Create(ctx context.Context, ranking *entity.Ranking) error {
	args := m.Called(ctx, ranking)
	return args.Error(0)
}

func (m *mockSubmitRankingRepo) Update(ctx context.Context, ranking *entity.Ranking) error {
	args := m.Called(ctx, ranking)
	return args.Error(0)
}

func (m *mockSubmitRankingRepo) Delete(ctx context.Context, id uint) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *mockSubmitRankingRepo) GetRank(ctx context.Context, gameType string, clearTimeMs uint) (int, error) {
	args := m.Called(ctx, gameType, clearTimeMs)
	return args.Int(0), args.Error(1)
}

func TestSubmitUseCase_Submit_NewRecord(t *testing.T) {
	t.Log("Submit: new record - creates and returns rank")
	mockRepo := new(mockSubmitRankingRepo)
	uc := NewSubmitUseCase(mockRepo, 5*time.Second)
	ctx := context.Background()
	userID := uint(1)
	nickname := "Alice"
	gameType := "puzzle"
	clearTimeMs := uint(5000)

	mockRepo.On("FindByUserAndGame", mock.Anything, userID, gameType).Return(nil, gorm.ErrRecordNotFound)
	mockRepo.On("Create", mock.Anything, mock.AnythingOfType("*entity.Ranking")).Run(func(args mock.Arguments) {
		r := args.Get(1).(*entity.Ranking)
		r.ID = 1
	}).Return(nil)
	mockRepo.On("GetRank", mock.Anything, gameType, clearTimeMs).Return(1, nil)

	res, err := uc.Submit(ctx, userID, nickname, gameType, clearTimeMs)
	require.NoError(t, err)
	require.NotNil(t, res)
	assert.Equal(t, 1, res.Rank)
	assert.True(t, res.IsNewRecord)
	mockRepo.AssertExpectations(t)
}

func TestSubmitUseCase_Submit_UpdateFasterTime(t *testing.T) {
	t.Log("Submit: existing record, faster time - updates")
	mockRepo := new(mockSubmitRankingRepo)
	uc := NewSubmitUseCase(mockRepo, 5*time.Second)
	ctx := context.Background()
	userID := uint(1)
	nickname := "Alice"
	gameType := "puzzle"
	clearTimeMs := uint(3000)

	existing := &entity.Ranking{ID: 1, UserID: userID, GameType: gameType, Nickname: "Old", ClearTimeMs: 5000}
	mockRepo.On("FindByUserAndGame", mock.Anything, userID, gameType).Return(existing, nil)
	mockRepo.On("Update", mock.Anything, mock.AnythingOfType("*entity.Ranking")).Return(nil)
	mockRepo.On("GetRank", mock.Anything, gameType, clearTimeMs).Return(1, nil)

	res, err := uc.Submit(ctx, userID, nickname, gameType, clearTimeMs)
	require.NoError(t, err)
	require.NotNil(t, res)
	assert.Equal(t, 1, res.Rank)
	assert.True(t, res.IsNewRecord)
	mockRepo.AssertExpectations(t)
}

func TestSubmitUseCase_Submit_NoUpdateSlowerTime(t *testing.T) {
	t.Log("Submit: existing record, slower time - no update")
	mockRepo := new(mockSubmitRankingRepo)
	uc := NewSubmitUseCase(mockRepo, 5*time.Second)
	ctx := context.Background()
	userID := uint(1)
	nickname := "Alice"
	gameType := "puzzle"
	clearTimeMs := uint(7000)

	existing := &entity.Ranking{ID: 1, UserID: userID, GameType: gameType, Nickname: "Alice", ClearTimeMs: 5000}
	mockRepo.On("FindByUserAndGame", mock.Anything, userID, gameType).Return(existing, nil)
	mockRepo.On("GetRank", mock.Anything, gameType, clearTimeMs).Return(3, nil)

	res, err := uc.Submit(ctx, userID, nickname, gameType, clearTimeMs)
	require.NoError(t, err)
	require.NotNil(t, res)
	assert.Equal(t, 3, res.Rank)
	assert.False(t, res.IsNewRecord)
	mockRepo.AssertNotCalled(t, "Update")
	mockRepo.AssertExpectations(t)
}
