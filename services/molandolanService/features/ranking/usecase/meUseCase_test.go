package usecase

import (
	"context"
	"testing"
	"time"

	"github.com/JokerTrickster/joker_backend/services/molandolanService/features/ranking/model/entity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

type mockMeRankingRepo struct {
	mock.Mock
}

func (m *mockMeRankingRepo) List(ctx context.Context, gameType string, page, limit int) ([]entity.Ranking, int64, error) {
	args := m.Called(ctx, gameType, page, limit)
	if args.Get(0) == nil {
		return nil, 0, args.Error(2)
	}
	return args.Get(0).([]entity.Ranking), args.Get(1).(int64), args.Error(2)
}

func (m *mockMeRankingRepo) FindByUserAndGame(ctx context.Context, userID uint, gameType string) (*entity.Ranking, error) {
	args := m.Called(ctx, userID, gameType)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Ranking), args.Error(1)
}

func (m *mockMeRankingRepo) Create(ctx context.Context, ranking *entity.Ranking) error {
	args := m.Called(ctx, ranking)
	return args.Error(0)
}

func (m *mockMeRankingRepo) Update(ctx context.Context, ranking *entity.Ranking) error {
	args := m.Called(ctx, ranking)
	return args.Error(0)
}

func (m *mockMeRankingRepo) Delete(ctx context.Context, id uint) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *mockMeRankingRepo) GetRank(ctx context.Context, gameType string, clearTimeMs uint) (int, error) {
	args := m.Called(ctx, gameType, clearTimeMs)
	return args.Int(0), args.Error(1)
}

func TestMeRankingUseCase_MyRanking_Success(t *testing.T) {
	t.Log("MyRanking: success")
	mockRepo := new(mockMeRankingRepo)
	uc := NewMeUseCase(mockRepo, 5*time.Second)
	ctx := context.Background()

	ranking := &entity.Ranking{
		ID: 1, UserID: 1, GameType: "puzzle", Nickname: "Me", ClearTimeMs: 4500,
	}
	mockRepo.On("FindByUserAndGame", mock.Anything, uint(1), "puzzle").Return(ranking, nil)
	mockRepo.On("GetRank", mock.Anything, "puzzle", uint(4500)).Return(2, nil)

	res, err := uc.MyRanking(ctx, 1, "puzzle")
	require.NoError(t, err)
	require.NotNil(t, res)
	assert.Equal(t, 2, res.Rank)
	assert.Equal(t, 2, res.Entry.Rank)
	assert.Equal(t, "Me", res.Entry.Nickname)
	mockRepo.AssertExpectations(t)
}

func TestMeRankingUseCase_MyRanking_NotFound(t *testing.T) {
	t.Log("MyRanking: not found")
	mockRepo := new(mockMeRankingRepo)
	uc := NewMeUseCase(mockRepo, 5*time.Second)
	ctx := context.Background()

	mockRepo.On("FindByUserAndGame", mock.Anything, uint(1), "puzzle").Return(nil, gorm.ErrRecordNotFound)

	res, err := uc.MyRanking(ctx, 1, "puzzle")
	assert.Error(t, err)
	assert.Nil(t, res)
	mockRepo.AssertExpectations(t)
	mockRepo.AssertNotCalled(t, "GetRank")
}
