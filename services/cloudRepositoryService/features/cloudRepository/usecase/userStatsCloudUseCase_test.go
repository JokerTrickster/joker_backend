package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetUserStats_SuccessWithStats(t *testing.T) {
	t.Logf("TestGetUserStats_SuccessWithStats: verifying stats retrieval with storage and monthly counts")

	mockRepo := new(MockUserStatsRepository)
	uc := NewUserStatsCloudRepositoryUseCase(mockRepo, 5*time.Second)

	ctx := context.Background()
	userID := uint(1)

	now := time.Now()
	mockRepo.On("GetTotalStorageUsed", mock.Anything, userID).Return(int64(2*1024*1024*1024), nil) // 2GB
	mockRepo.On("GetMonthlyUploadCount", mock.Anything, userID, now.Year(), int(now.Month())).Return(10, nil)
	mockRepo.On("GetMonthlyDownloadCount", mock.Anything, userID, now.Year(), int(now.Month())).Return(5, nil)
	mockRepo.On("GetMonthlyTagsCreatedCount", mock.Anything, userID, now.Year(), int(now.Month())).Return(3, nil)

	result, err := uc.GetUserStats(ctx, userID)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, int64(2*1024*1024*1024), result.Storage.Used)
	assert.Equal(t, int64(15*1024*1024*1024), result.Storage.Total)
	assert.Greater(t, result.Storage.Percentage, 0.0)
	assert.Less(t, result.Storage.Percentage, 100.0)
	assert.Equal(t, 10, result.MonthlyStats.Uploads)
	assert.Equal(t, 5, result.MonthlyStats.Downloads)
	assert.Equal(t, 3, result.MonthlyStats.TagsCreated)
	t.Logf("Success: user stats retrieved correctly")

	mockRepo.AssertExpectations(t)
}

func TestGetUserStats_ZeroStorage(t *testing.T) {
	t.Logf("TestGetUserStats_ZeroStorage: verifying zero storage returns 0 percentage")

	mockRepo := new(MockUserStatsRepository)
	uc := NewUserStatsCloudRepositoryUseCase(mockRepo, 5*time.Second)

	ctx := context.Background()
	userID := uint(1)

	now := time.Now()
	mockRepo.On("GetTotalStorageUsed", mock.Anything, userID).Return(int64(0), nil)
	mockRepo.On("GetMonthlyUploadCount", mock.Anything, userID, now.Year(), int(now.Month())).Return(0, nil)
	mockRepo.On("GetMonthlyDownloadCount", mock.Anything, userID, now.Year(), int(now.Month())).Return(0, nil)
	mockRepo.On("GetMonthlyTagsCreatedCount", mock.Anything, userID, now.Year(), int(now.Month())).Return(0, nil)

	result, err := uc.GetUserStats(ctx, userID)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, int64(0), result.Storage.Used)
	assert.Equal(t, float64(0), result.Storage.Percentage)
	assert.Equal(t, 0, result.MonthlyStats.Uploads)
	t.Logf("Success: zero storage handled correctly")

	mockRepo.AssertExpectations(t)
}

func TestGetUserStats_RepoError(t *testing.T) {
	t.Logf("TestGetUserStats_RepoError: verifying error propagation from repository")

	mockRepo := new(MockUserStatsRepository)
	uc := NewUserStatsCloudRepositoryUseCase(mockRepo, 5*time.Second)

	ctx := context.Background()
	userID := uint(1)

	mockRepo.On("GetTotalStorageUsed", mock.Anything, userID).Return(int64(0), errors.New("db error"))

	result, err := uc.GetUserStats(ctx, userID)

	assert.Error(t, err)
	assert.Nil(t, result)
	t.Logf("Expected: repository error propagated")

	mockRepo.AssertExpectations(t)
}
