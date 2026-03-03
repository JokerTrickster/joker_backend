package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/JokerTrickster/joker_backend/services/cloudRepositoryService/features/cloudRepository/model/entity"
	"github.com/JokerTrickster/joker_backend/services/cloudRepositoryService/features/cloudRepository/model/request"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockActivityHistoryRepository mocks IActivityHistoryCloudRepositoryRepository
type MockActivityHistoryRepository struct {
	mock.Mock
}

func (m *MockActivityHistoryRepository) GetMonthlyActivity(ctx context.Context, userID uint, year int, month int) ([]entity.ActivityLog, error) {
	args := m.Called(ctx, userID, year, month)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]entity.ActivityLog), args.Error(1)
}

func (m *MockActivityHistoryRepository) GetMonthlyUsedTags(ctx context.Context, userID uint, year int, month int) (map[string][]string, error) {
	args := m.Called(ctx, userID, year, month)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[string][]string), args.Error(1)
}

func TestGetActivityHistory_SuccessWithActivity(t *testing.T) {
	t.Logf("TestGetActivityHistory_SuccessWithActivity: verifying activity data aggregation by day")

	mockRepo := new(MockActivityHistoryRepository)
	uc := NewActivityHistoryCloudRepositoryUseCase(mockRepo, 5*time.Second)

	ctx := context.Background()
	userID := uint(1)
	req := &request.ActivityHistoryRequestDTO{Month: "2024-03"}

	now := time.Date(2024, 3, 15, 10, 0, 0, 0, time.UTC)
	mockActivities := []entity.ActivityLog{
		{UserID: userID, ActivityType: entity.ActivityTypeUpload, CreatedAt: now},
		{UserID: userID, ActivityType: entity.ActivityTypeDownload, CreatedAt: now},
		{UserID: userID, ActivityType: entity.ActivityTypeUpload, CreatedAt: now},
	}
	mockTags := map[string][]string{"2024-03-15": {"work", "project"}}

	mockRepo.On("GetMonthlyActivity", mock.Anything, userID, 2024, 3).Return(mockActivities, nil)
	mockRepo.On("GetMonthlyUsedTags", mock.Anything, userID, 2024, 3).Return(mockTags, nil)

	result, err := uc.GetActivityHistory(ctx, userID, req)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.NotEmpty(t, *result)
	dateStr := "2024-03-15"
	if daily, ok := (*result)[dateStr]; ok {
		assert.Equal(t, 2, daily.Uploads)
		assert.Equal(t, 1, daily.Downloads)
		assert.Contains(t, daily.Tags, "work")
		assert.Contains(t, daily.Tags, "project")
	}
	t.Logf("Success: activity history aggregated by day")

	mockRepo.AssertExpectations(t)
}

func TestGetActivityHistory_EmptyMonth(t *testing.T) {
	t.Logf("TestGetActivityHistory_EmptyMonth: verifying empty result for month with no activity")

	mockRepo := new(MockActivityHistoryRepository)
	uc := NewActivityHistoryCloudRepositoryUseCase(mockRepo, 5*time.Second)

	ctx := context.Background()
	userID := uint(1)
	req := &request.ActivityHistoryRequestDTO{Month: "2024-06"}

	mockRepo.On("GetMonthlyActivity", mock.Anything, userID, 2024, 6).Return([]entity.ActivityLog{}, nil)
	mockRepo.On("GetMonthlyUsedTags", mock.Anything, userID, 2024, 6).Return(map[string][]string{}, nil)

	result, err := uc.GetActivityHistory(ctx, userID, req)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Empty(t, *result)
	t.Logf("Success: empty month returns empty result")

	mockRepo.AssertExpectations(t)
}

func TestGetActivityHistory_DefaultMonth(t *testing.T) {
	t.Logf("TestGetActivityHistory_DefaultMonth: verifying default to current month when not provided")

	mockRepo := new(MockActivityHistoryRepository)
	uc := NewActivityHistoryCloudRepositoryUseCase(mockRepo, 5*time.Second)

	ctx := context.Background()
	userID := uint(1)
	req := &request.ActivityHistoryRequestDTO{Month: ""}

	now := time.Now()
	mockRepo.On("GetMonthlyActivity", mock.Anything, userID, now.Year(), int(now.Month())).Return([]entity.ActivityLog{}, nil)
	mockRepo.On("GetMonthlyUsedTags", mock.Anything, userID, now.Year(), int(now.Month())).Return(map[string][]string{}, nil)

	result, err := uc.GetActivityHistory(ctx, userID, req)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	t.Logf("Success: default month (current) used when empty")

	mockRepo.AssertExpectations(t)
}

func TestGetActivityHistory_InvalidMonthFormat(t *testing.T) {
	t.Logf("TestGetActivityHistory_InvalidMonthFormat: verifying error for invalid month format")

	mockRepo := new(MockActivityHistoryRepository)
	uc := NewActivityHistoryCloudRepositoryUseCase(mockRepo, 5*time.Second)

	ctx := context.Background()
	userID := uint(1)
	req := &request.ActivityHistoryRequestDTO{Month: "invalid"}

	result, err := uc.GetActivityHistory(ctx, userID, req)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "invalid month format")
	t.Logf("Expected: repository should not be called for invalid format")

	mockRepo.AssertNotCalled(t, "GetMonthlyActivity")
}

func TestGetActivityHistory_RepoError(t *testing.T) {
	t.Logf("TestGetActivityHistory_RepoError: verifying error propagation from repository")

	mockRepo := new(MockActivityHistoryRepository)
	uc := NewActivityHistoryCloudRepositoryUseCase(mockRepo, 5*time.Second)

	ctx := context.Background()
	userID := uint(1)
	req := &request.ActivityHistoryRequestDTO{Month: "2024-03"}

	mockRepo.On("GetMonthlyActivity", mock.Anything, userID, 2024, 3).Return(nil, errors.New("db error"))

	result, err := uc.GetActivityHistory(ctx, userID, req)

	assert.Error(t, err)
	assert.Nil(t, result)
	t.Logf("Expected: repository error propagated")

	mockRepo.AssertExpectations(t)
}
