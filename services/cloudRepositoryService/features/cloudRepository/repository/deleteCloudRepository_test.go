package repository

import (
	"context"
	"testing"
	"time"

	"github.com/JokerTrickster/joker_backend/services/cloudRepositoryService/features/cloudRepository/model/entity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDeleteCloudRepository_DeleteFromS3_SkipsWithoutS3(t *testing.T) {
	t.Skip("DeleteFromS3 requires S3 - skip in integration test without AWS config")
}

func TestDeleteCloudRepository_SoftDeleteFile_Success(t *testing.T) {
	db := setupTestDB(t)
	ctx := context.Background()

	file := &entity.CloudFile{
		UserID:      997,
		FileName:   "test_delete.jpg",
		S3Key:      "user997/test_delete.jpg",
		FileType:   entity.FileTypeImage,
		ContentType: "image/jpeg",
		FileSize:   512,
	}
	err := db.WithContext(ctx).Create(file).Error
	require.NoError(t, err)
	fileID := file.ID
	userID := uint(997)
	t.Logf("Setup: created file ID=%d", fileID)

	repo := NewDeleteCloudRepositoryRepository(db, "test-bucket")
	err = repo.SoftDeleteFile(ctx, fileID, userID)
	require.NoError(t, err)
	t.Logf("SoftDeleteFile: soft deleted successfully")

	// Verify file is soft deleted (deleted_at set)
	var deleted entity.CloudFile
	err = db.WithContext(ctx).Unscoped().Where("id = ?", fileID).First(&deleted).Error
	require.NoError(t, err)
	assert.NotNil(t, deleted.DeletedAt)
	t.Logf("Verified: deleted_at is set")
}

func TestDeleteCloudRepository_GetFileByID_Success(t *testing.T) {
	db := setupTestDB(t)
	ctx := context.Background()

	file := &entity.CloudFile{
		UserID:      996,
		FileName:   "test_get.jpg",
		S3Key:      "user996/test_get.jpg",
		FileType:   entity.FileTypeImage,
		ContentType: "image/jpeg",
		FileSize:   1024,
	}
	err := db.WithContext(ctx).Create(file).Error
	require.NoError(t, err)
	defer db.WithContext(ctx).Model(&entity.CloudFile{}).Where("id = ?", file.ID).Update("deleted_at", time.Now())

	repo := NewDeleteCloudRepositoryRepository(db, "test-bucket")
	fetched, err := repo.GetFileByID(ctx, file.ID)
	require.NoError(t, err)
	assert.Equal(t, file.ID, fetched.ID)
	t.Logf("GetFileByID: retrieved file successfully")
}
