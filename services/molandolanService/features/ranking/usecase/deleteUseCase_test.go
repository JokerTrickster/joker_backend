package usecase

import (
	"context"
	"testing"
	"time"

	"github.com/JokerTrickster/joker_backend/services/molandolanService/features/ranking/model/entity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/mock"
)

type mockDeleteRankingRepo struct {
	mock.Mock
}

func (m *mockDeleteRankingRepo) List(ctx context.Context, gameType string, page, limit int) ([]entity.Ranking, int64, error) {
	args := m.Called(ctx, gameType, page, limit)
	if args.Get(0) == nil {
		return nil, 0, args.Error(2)
	}
	return args.Get(0).([]entity.Ranking), args.Get(1).(int64), args.Error(2)
}

func (m *mockDeleteRankingRepo) FindByUserAndGame(ctx context.Context, userID uint, gameType string) (*entity.Ranking, error) {
	args := m.Called(ctx, userID, gameType)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Ranking), args.Error(1)
}

func (m *mockDeleteRankingRepo) Create(ctx context.Context, ranking *entity.Ranking) error {
	args := m.Called(ctx, ranking)
	return args.Error(0)
}

func (m *mockDeleteRankingRepo) Update(ctx context.Context, ranking *entity.Ranking) error {
	args := m.Called(ctx, ranking)
	return args.Error(0)
}

func (m *mockDeleteRankingRepo) Delete(ctx context.Context, id uint) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *mockDeleteRankingRepo) GetRank(ctx context.Context, gameType string, clearTimeMs uint) (int, error) {
	args := m.Called(ctx, gameType, clearTimeMs)
	return args.Int(0), args.Error(1)
}

func TestDeleteRankingUseCase_Delete_Success(t *testing.T) {
	t.Log("Delete: success")
	mockRepo := new(mockDeleteRankingRepo)
	uc := NewDeleteUseCase(mockRepo, 5*time.Second)
	ctx := context.Background()

	mockRepo.On("Delete", mock.Anything, uint(1)).Return(nil)

	err := uc.Delete(ctx, 1)
	require.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestDeleteRankingUseCase_Delete_RepoError(t *testing.T) {
	t.Log("Delete: repo error")
	mockRepo := new(mockDeleteRankingRepo)
	uc := NewDeleteUseCase(mockRepo, 5*time.Second)
	ctx := context.Background()

	mockRepo.On("Delete", mock.Anything, uint(999)).Return(assert.AnError)

	err := uc.Delete(ctx, 999)
	assert.Error(t, err)
	mockRepo.AssertExpectations(t)
}
