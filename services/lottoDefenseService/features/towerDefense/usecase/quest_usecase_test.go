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
	"gorm.io/gorm"
)

type mockTDQuestRepository struct {
	mock.Mock
}

func (m *mockTDQuestRepository) Create(ctx context.Context, quest *entity.TDQuest) error {
	args := m.Called(ctx, quest)
	return args.Error(0)
}

func (m *mockTDQuestRepository) GetByID(ctx context.Context, id uint) (*entity.TDQuest, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.TDQuest), args.Error(1)
}

func (m *mockTDQuestRepository) GetActiveQuests(ctx context.Context, userID uint) ([]entity.TDQuest, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]entity.TDQuest), args.Error(1)
}

func (m *mockTDQuestRepository) UpdateProgress(ctx context.Context, questID uint, increment uint) error {
	args := m.Called(ctx, questID, increment)
	return args.Error(0)
}

func (m *mockTDQuestRepository) CompleteQuest(ctx context.Context, questID uint) error {
	args := m.Called(ctx, questID)
	return args.Error(0)
}

func (m *mockTDQuestRepository) ClaimQuest(ctx context.Context, questID uint) error {
	args := m.Called(ctx, questID)
	return args.Error(0)
}

func TestTDQuestUseCase_GetActiveQuests_Success(t *testing.T) {
	t.Log("GetActiveQuests: success")
	mockQuestRepo := new(mockTDQuestRepository)
	mockUserRepo := new(mockTDUserRepository)
	uc := NewTDQuestUseCase(mockQuestRepo, mockUserRepo, 5*time.Second)
	ctx := context.Background()
	userID := uint(1)

	mockQuestRepo.On("GetActiveQuests", ctx, userID).
		Return([]entity.TDQuest{
			{ID: 1, UserID: userID, QuestType: "kill", QuestName: "Kill 10", TargetCount: 10, CurrentCount: 5, Status: "active", CreatedAt: time.Now()},
		}, nil)

	resp, err := uc.GetActiveQuests(ctx, userID)
	require.NoError(t, err)
	require.Len(t, resp, 1)
	assert.Equal(t, uint(1), resp[0].QuestID)
	assert.Equal(t, "Kill 10", resp[0].QuestName)
	mockQuestRepo.AssertExpectations(t)
}

func TestTDQuestUseCase_UpdateQuestProgress_Success(t *testing.T) {
	t.Log("UpdateQuestProgress: success - does not complete (target not reached)")
	mockQuestRepo := new(mockTDQuestRepository)
	mockUserRepo := new(mockTDUserRepository)
	uc := NewTDQuestUseCase(mockQuestRepo, mockUserRepo, 5*time.Second)
	ctx := context.Background()
	userID := uint(1)
	questID := uint(1)
	req := &request.UpdateQuestProgressRequest{Increment: 5}

	quest := &entity.TDQuest{ID: questID, UserID: userID, TargetCount: 20, CurrentCount: 5, Status: "active", CreatedAt: time.Now()}
	mockQuestRepo.On("GetByID", ctx, questID).Return(quest, nil).Once()
	mockQuestRepo.On("UpdateProgress", ctx, questID, uint(5)).Return(nil).Once()
	mockQuestRepo.On("GetByID", ctx, questID).Return(&entity.TDQuest{ID: questID, UserID: userID, TargetCount: 20, CurrentCount: 10, Status: "active", CreatedAt: time.Now()}, nil).Once()

	resp, err := uc.UpdateQuestProgress(ctx, userID, questID, req)
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, uint(10), resp.CurrentCount)
	assert.Equal(t, "active", resp.Status)
	mockQuestRepo.AssertExpectations(t)
	mockQuestRepo.AssertNotCalled(t, "CompleteQuest")
}

func TestTDQuestUseCase_UpdateQuestProgress_AutoCompletes(t *testing.T) {
	t.Log("UpdateQuestProgress: success - auto-completes quest when target reached")
	mockQuestRepo := new(mockTDQuestRepository)
	mockUserRepo := new(mockTDUserRepository)
	uc := NewTDQuestUseCase(mockQuestRepo, mockUserRepo, 5*time.Second)
	ctx := context.Background()
	userID := uint(1)
	questID := uint(1)
	req := &request.UpdateQuestProgressRequest{Increment: 5}

	quest := &entity.TDQuest{ID: questID, UserID: userID, TargetCount: 10, CurrentCount: 5, Status: "active", CreatedAt: time.Now()}

	mockQuestRepo.On("GetByID", ctx, questID).Return(quest, nil).Once()
	mockQuestRepo.On("UpdateProgress", ctx, questID, uint(5)).Return(nil).Once()
	mockQuestRepo.On("GetByID", ctx, questID).Return(&entity.TDQuest{ID: questID, UserID: userID, TargetCount: 10, CurrentCount: 10, Status: "active", CreatedAt: time.Now()}, nil).Once()
	mockQuestRepo.On("CompleteQuest", ctx, questID).Return(nil).Once()

	resp, err := uc.UpdateQuestProgress(ctx, userID, questID, req)
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, uint(10), resp.CurrentCount)
	assert.Equal(t, "completed", resp.Status)
	mockQuestRepo.AssertExpectations(t)
}

func TestTDQuestUseCase_ClaimReward_Success(t *testing.T) {
	t.Log("ClaimReward: success")
	mockQuestRepo := new(mockTDQuestRepository)
	mockUserRepo := new(mockTDUserRepository)
	uc := NewTDQuestUseCase(mockQuestRepo, mockUserRepo, 5*time.Second)
	ctx := context.Background()
	userID := uint(1)
	questID := uint(1)

	quest := &entity.TDQuest{ID: questID, UserID: userID, QuestType: "kill", QuestName: "Kill 10", TargetCount: 10, CurrentCount: 10, Status: "completed", RewardGold: 100, CreatedAt: time.Now()}
	stats := &entity.TDUserStats{UserID: userID, CurrentGold: 50, QuestsCompleted: 0}

	mockQuestRepo.On("GetByID", ctx, questID).Return(quest, nil)
	mockQuestRepo.On("ClaimQuest", ctx, questID).Return(nil)
	mockUserRepo.On("GetStats", ctx, userID).Return(stats, nil)
	mockUserRepo.On("UpdateStats", ctx, mock.AnythingOfType("*entity.TDUserStats")).Return(nil)

	resp, err := uc.ClaimReward(ctx, userID, questID)
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, uint(150), resp.NewGold)
	mockQuestRepo.AssertExpectations(t)
	mockUserRepo.AssertExpectations(t)
}

func TestTDQuestUseCase_ClaimReward_QuestNotCompleted(t *testing.T) {
	t.Log("ClaimReward: quest not completed error")
	mockQuestRepo := new(mockTDQuestRepository)
	mockUserRepo := new(mockTDUserRepository)
	uc := NewTDQuestUseCase(mockQuestRepo, mockUserRepo, 5*time.Second)
	ctx := context.Background()
	userID := uint(1)
	questID := uint(1)

	quest := &entity.TDQuest{ID: questID, UserID: userID, Status: "active", TargetCount: 10, CurrentCount: 5, CreatedAt: time.Now()}
	mockQuestRepo.On("GetByID", ctx, questID).Return(quest, nil)

	resp, err := uc.ClaimReward(ctx, userID, questID)
	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "quest not completed")
	mockQuestRepo.AssertExpectations(t)
	mockQuestRepo.AssertNotCalled(t, "ClaimQuest")
}

func TestTDQuestUseCase_ClaimReward_QuestNotFound(t *testing.T) {
	t.Log("ClaimReward: quest not found")
	mockQuestRepo := new(mockTDQuestRepository)
	mockUserRepo := new(mockTDUserRepository)
	uc := NewTDQuestUseCase(mockQuestRepo, mockUserRepo, 5*time.Second)
	ctx := context.Background()
	userID := uint(1)
	questID := uint(999)

	mockQuestRepo.On("GetByID", ctx, questID).Return(nil, gorm.ErrRecordNotFound)

	resp, err := uc.ClaimReward(ctx, userID, questID)
	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "quest not found")
	mockQuestRepo.AssertExpectations(t)
}

func TestTDQuestUseCase_UpdateQuestProgress_WrongUser(t *testing.T) {
	t.Log("UpdateQuestProgress: quest belongs to different user -> quest not found")
	mockQuestRepo := new(mockTDQuestRepository)
	mockUserRepo := new(mockTDUserRepository)
	uc := NewTDQuestUseCase(mockQuestRepo, mockUserRepo, 5*time.Second)
	ctx := context.Background()
	userID := uint(2)
	questID := uint(1)
	req := &request.UpdateQuestProgressRequest{Increment: 5}

	quest := &entity.TDQuest{ID: questID, UserID: 1, TargetCount: 10, CurrentCount: 5, Status: "active", CreatedAt: time.Now()}
	mockQuestRepo.On("GetByID", ctx, questID).Return(quest, nil)

	resp, err := uc.UpdateQuestProgress(ctx, userID, questID, req)
	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "quest not found")
	mockQuestRepo.AssertExpectations(t)
	mockQuestRepo.AssertNotCalled(t, "UpdateProgress")
}

func TestTDQuestUseCase_GetActiveQuests_Empty(t *testing.T) {
	t.Log("GetActiveQuests: empty list")
	mockQuestRepo := new(mockTDQuestRepository)
	mockUserRepo := new(mockTDUserRepository)
	uc := NewTDQuestUseCase(mockQuestRepo, mockUserRepo, 5*time.Second)
	ctx := context.Background()
	userID := uint(1)

	mockQuestRepo.On("GetActiveQuests", ctx, userID).Return([]entity.TDQuest{}, nil)

	resp, err := uc.GetActiveQuests(ctx, userID)
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Empty(t, resp)
	mockQuestRepo.AssertExpectations(t)
}

func TestTDQuestUseCase_GetActiveQuests_TargetCountZero(t *testing.T) {
	t.Log("GetActiveQuests: target count 0 avoids division by zero")
	mockQuestRepo := new(mockTDQuestRepository)
	mockUserRepo := new(mockTDUserRepository)
	uc := NewTDQuestUseCase(mockQuestRepo, mockUserRepo, 5*time.Second)
	ctx := context.Background()
	userID := uint(1)

	quest := entity.TDQuest{ID: 1, UserID: userID, QuestType: "kill", TargetCount: 0, CurrentCount: 0, Status: "active", CreatedAt: time.Now()}
	mockQuestRepo.On("GetActiveQuests", ctx, userID).Return([]entity.TDQuest{quest}, nil)

	resp, err := uc.GetActiveQuests(ctx, userID)
	require.NoError(t, err)
	require.Len(t, resp, 1)
	assert.Equal(t, float64(0), resp[0].Progress)
	mockQuestRepo.AssertExpectations(t)
}
