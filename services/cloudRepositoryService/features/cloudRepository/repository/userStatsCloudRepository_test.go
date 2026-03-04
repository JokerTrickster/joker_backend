package repository

import (
	"context"
	"testing"
	"time"

	"github.com/JokerTrickster/joker_backend/services/cloudRepositoryService/features/cloudRepository/model/entity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUserStatsCloudRepository_GetTotalStorageUsed_Success(t *testing.T) {
	db := setupTestDB(t)
	requireTable(t, db, "cloud_files")
	ctx := context.Background()

	userID := uint(973)
	file := &entity.CloudFile{
		UserID:      userID,
		FileName:   "stats_file.jpg",
		S3Key:      "user973/stats_file.jpg",
		FileType:   entity.FileTypeImage,
		ContentType: "image/jpeg",
		FileSize:   2048,
	}
	err := db.WithContext(ctx).Create(file).Error
	require.NoError(t, err)
	defer db.WithContext(ctx).Model(&entity.CloudFile{}).Where("id = ?", file.ID).Update("deleted_at", time.Now())

	repo := NewUserStatsCloudRepositoryRepository(db)
	used, err := repo.GetTotalStorageUsed(ctx, userID)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, used, int64(2048))
	t.Logf("GetTotalStorageUsed: used=%d bytes", used)
}

func TestUserStatsCloudRepository_GetTotalStorageUsed_Zero(t *testing.T) {
	db := setupTestDB(t)
	requireTable(t, db, "cloud_files")
	ctx := context.Background()

	repo := NewUserStatsCloudRepositoryRepository(db)
	used, err := repo.GetTotalStorageUsed(ctx, 999998)
	require.NoError(t, err)
	assert.Equal(t, int64(0), used)
	t.Logf("GetTotalStorageUsed: zero for user with no files")
}

func TestUserStatsCloudRepository_LogActivity_Success(t *testing.T) {
	db := setupTestDB(t)
	requireTable(t, db, "activity_logs")
	ctx := context.Background()

	repo := NewUserStatsCloudRepositoryRepository(db)
	activity := &entity.ActivityLog{
		UserID:       972,
		ActivityType: entity.ActivityTypeUpload,
	}

	err := repo.LogActivity(ctx, activity)
	require.NoError(t, err)
	assert.Greater(t, activity.ID, uint(0))
	t.Logf("LogActivity: logged upload activity ID=%d", activity.ID)

	db.WithContext(ctx).Where("id = ?", activity.ID).Delete(&entity.ActivityLog{})
}

func TestUserStatsCloudRepository_GetMonthlyUploadCount_Success(t *testing.T) {
	db := setupTestDB(t)
	requireTable(t, db, "activity_logs")
	ctx := context.Background()

	repo := NewUserStatsCloudRepositoryRepository(db)
	now := time.Now()
	count, err := repo.GetMonthlyUploadCount(ctx, 999997, now.Year(), int(now.Month()))
	require.NoError(t, err)
	assert.GreaterOrEqual(t, count, 0)
	t.Logf("GetMonthlyUploadCount: count=%d", count)
}
