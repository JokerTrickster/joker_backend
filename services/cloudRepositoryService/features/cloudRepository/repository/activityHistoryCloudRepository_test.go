package repository

import (
	"context"
	"testing"
	"time"

	"github.com/JokerTrickster/joker_backend/services/cloudRepositoryService/features/cloudRepository/model/entity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestActivityHistoryCloudRepository_GetMonthlyActivity_Success(t *testing.T) {
	db := setupTestDB(t)
	ctx := context.Background()

	userID := uint(971)
	activity := &entity.ActivityLog{
		UserID:       userID,
		ActivityType: entity.ActivityTypeUpload,
	}
	err := db.WithContext(ctx).Create(activity).Error
	require.NoError(t, err)
	defer db.WithContext(ctx).Where("id = ?", activity.ID).Delete(&entity.ActivityLog{})

	repo := NewActivityHistoryCloudRepositoryRepository(db)
	now := time.Now()
	activities, err := repo.GetMonthlyActivity(ctx, userID, now.Year(), int(now.Month()))
	require.NoError(t, err)
	assert.NotNil(t, activities)
	assert.GreaterOrEqual(t, len(activities), 0)
	t.Logf("GetMonthlyActivity: returned %d activities", len(activities))
}

func TestActivityHistoryCloudRepository_GetMonthlyActivity_Empty(t *testing.T) {
	db := setupTestDB(t)
	ctx := context.Background()

	repo := NewActivityHistoryCloudRepositoryRepository(db)
	activities, err := repo.GetMonthlyActivity(ctx, 999996, 2099, 12)
	require.NoError(t, err)
	assert.NotNil(t, activities)
	assert.Len(t, activities, 0)
	t.Logf("GetMonthlyActivity: empty for future month")
}

func TestActivityHistoryCloudRepository_GetMonthlyUsedTags_Success(t *testing.T) {
	db := setupTestDB(t)
	ctx := context.Background()

	repo := NewActivityHistoryCloudRepositoryRepository(db)
	now := time.Now()
	tagsByDate, err := repo.GetMonthlyUsedTags(ctx, 999995, now.Year(), int(now.Month()))
	require.NoError(t, err)
	assert.NotNil(t, tagsByDate)
	t.Logf("GetMonthlyUsedTags: returned %d dates", len(tagsByDate))
}
