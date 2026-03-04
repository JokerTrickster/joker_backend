package repository

import (
	"context"
	"testing"
	"time"

	"github.com/JokerTrickster/joker_backend/services/cloudRepositoryService/features/cloudRepository/model/entity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDownloadCloudRepository_PresignedURL_RequiresS3(t *testing.T) {
	t.Skip("GeneratePresignedDownloadURL and GeneratePresignedDownloadURLWithFilename require S3 - skip in integration test without AWS config")
}

func TestDownloadCloudRepository_GetFileByID_Found(t *testing.T) {
	t.Logf("TestDownloadCloudRepository_GetFileByID_Found: insert CloudFile, get by ID, verify returned")

	db := setupTestDB(t)
	requireTable(t, db, "cloud_files")
	ctx := context.Background()

	require.NoError(t, db.AutoMigrate(&entity.CloudFile{}))
	defer func() {
		_ = db.WithContext(ctx).Where("user_id = ?", uint(991)).Delete(&entity.CloudFile{}).Error
	}()

	file := &entity.CloudFile{
		UserID:      991,
		FileName:   "test_download_found.jpg",
		S3Key:      "user991/test_download_found.jpg",
		FileType:   entity.FileTypeImage,
		ContentType: "image/jpeg",
		FileSize:   2048,
	}
	err := db.WithContext(ctx).Create(file).Error
	require.NoError(t, err)
	fileID := file.ID
	t.Logf("Setup: created file ID=%d, S3Key=%s", fileID, file.S3Key)

	repo := NewDownloadCloudRepositoryRepository(db, "test-bucket")
	fetched, err := repo.GetFileByID(ctx, fileID)

	require.NoError(t, err)
	assert.NotNil(t, fetched)
	assert.Equal(t, fileID, fetched.ID)
	assert.Equal(t, file.FileName, fetched.FileName)
	assert.Equal(t, file.S3Key, fetched.S3Key)
	t.Logf("Success: GetFileByID returned file correctly")
}

func TestDownloadCloudRepository_GetFileByID_NotFound(t *testing.T) {
	t.Logf("TestDownloadCloudRepository_GetFileByID_NotFound: get non-existent ID, expect error")

	db := setupTestDB(t)
	requireTable(t, db, "cloud_files")
	ctx := context.Background()

	repo := NewDownloadCloudRepositoryRepository(db, "test-bucket")
	fetched, err := repo.GetFileByID(ctx, 99999999)

	assert.Error(t, err)
	assert.Nil(t, fetched)
	t.Logf("Success: GetFileByID returned error for non-existent ID")
}

func TestDownloadCloudRepository_GetFileByID_Deleted(t *testing.T) {
	t.Logf("TestDownloadCloudRepository_GetFileByID_Deleted: insert with deleted_at set, GetFileByID should not find it")

	db := setupTestDB(t)
	requireTable(t, db, "cloud_files")
	ctx := context.Background()

	require.NoError(t, db.AutoMigrate(&entity.CloudFile{}))
	now := time.Now()

	file := &entity.CloudFile{
		UserID:      992,
		FileName:   "test_download_deleted.jpg",
		S3Key:      "user992/test_download_deleted.jpg",
		FileType:   entity.FileTypeImage,
		ContentType: "image/jpeg",
		FileSize:   1024,
		DeletedAt:  &now,
	}
	err := db.WithContext(ctx).Create(file).Error
	require.NoError(t, err)
	fileID := file.ID
	t.Logf("Setup: created soft-deleted file ID=%d", fileID)

	repo := NewDownloadCloudRepositoryRepository(db, "test-bucket")
	fetched, err := repo.GetFileByID(ctx, fileID)

	assert.Error(t, err)
	assert.Nil(t, fetched)
	t.Logf("Success: GetFileByID correctly excludes soft-deleted files")
}
