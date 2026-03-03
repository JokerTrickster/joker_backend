package usecase

import (
	"context"
	"testing"
	"time"

	"github.com/JokerTrickster/joker_backend/services/morandoranService/features/ranking/model/entity"
	"github.com/JokerTrickster/joker_backend/services/morandoranService/features/ranking/model/request"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/mock"
)

type mockListRankingRepo struct {
	mock.Mock
}

func (m *mockListRankingRepo) List(ctx context.Context, gameType string, page, limit int) ([]entity.Ranking, int64, error) {
	args := m.Called(ctx, gameType, page, limit)
	if args.Get(0) == nil {
		return nil, 0, args.Error(2)
	}
	return args.Get(0).([]entity.Ranking), args.Get(1).(int64), args.Error(2)
}

func (m *mockListRankingRepo) FindByUserAndGame(ctx context.Context, userID uint, gameType string) (*entity.Ranking, error) {
	args := m.Called(ctx, userID, gameType)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Ranking), args.Error(1)
}

func (m *mockListRankingRepo) Create(ctx context.Context, ranking *entity.Ranking) error {
	args := m.Called(ctx, ranking)
	return args.Error(0)
}

func (m *mockListRankingRepo) Update(ctx context.Context, ranking *entity.Ranking) error {
	args := m.Called(ctx, ranking)
	return args.Error(0)
}

func (m *mockListRankingRepo) Delete(ctx context.Context, id uint) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *mockListRankingRepo) GetRank(ctx context.Context, gameType string, clearTimeMs uint) (int, error) {
	args := m.Called(ctx, gameType, clearTimeMs)
	return args.Int(0), args.Error(1)
}

func TestListRankingUseCase_List_Success(t *testing.T) {
	t.Log("List: success")
	mockRepo := new(mockListRankingRepo)
	uc := NewListUseCase(mockRepo, 5*time.Second)
	ctx := context.Background()
	req := &request.ReqListRanking{Limit: 5}

	items := []entity.Ranking{
		{ID: 1, UserID: 1, GameType: "puzzle", Nickname: "A", ClearTimeMs: 3000},
		{ID: 2, UserID: 2, GameType: "puzzle", Nickname: "B", ClearTimeMs: 4000},
	}
	mockRepo.On("List", mock.Anything, "puzzle", 1, 5).Return(items, int64(2), nil)

	res, err := uc.List(ctx, "puzzle", req)
	require.NoError(t, err)
	require.NotNil(t, res)
	assert.Len(t, res.Rankings, 2)
	assert.Equal(t, 1, res.Rankings[0].Rank)
	assert.Equal(t, "A", res.Rankings[0].Nickname)
	assert.Equal(t, uint(3000), res.Rankings[0].ClearTimeMs)
	mockRepo.AssertExpectations(t)
}

func TestListRankingUseCase_List_DefaultLimit(t *testing.T) {
	t.Log("List: limit<1 uses default 5")
	mockRepo := new(mockListRankingRepo)
	uc := NewListUseCase(mockRepo, 5*time.Second)
	ctx := context.Background()
	req := &request.ReqListRanking{Limit: 0}

	mockRepo.On("List", mock.Anything, "puzzle", 1, 5).Return([]entity.Ranking{}, int64(0), nil)

	res, err := uc.List(ctx, "puzzle", req)
	require.NoError(t, err)
	require.NotNil(t, res)
	mockRepo.AssertExpectations(t)
}
