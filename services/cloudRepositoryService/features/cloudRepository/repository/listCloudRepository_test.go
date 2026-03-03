package repository

import (
	"context"
	"testing"
	"time"

	"github.com/JokerTrickster/joker_backend/services/cloudRepositoryService/features/cloudRepository/model/entity"
	"github.com/JokerTrickster/joker_backend/services/cloudRepositoryService/features/cloudRepository/model/request"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestListCloudRepository_GeneratePresignedDownloadURL_SkipsWithoutS3(t *testing.T) {
	t.Skip("GeneratePresignedDownloadURL requires S3 - skip in integration test without AWS config")
}

func TestListCloudRepository_GetFilesByUserID_Success(t *testing.T) {
	db := setupTestDB(t)
	ctx := context.Background()

	userID := uint(995)
	file := &entity.CloudFile{
		UserID:      userID,
		FileName:    "test_list.jpg",
		S3Key:       "user995/test_list.jpg",
		FileType:    entity.FileTypeImage,
		ContentType: "image/jpeg",
		FileSize:    1024,
	}
	err := db.WithContext(ctx).Create(file).Error
	require.NoError(t, err)
	defer db.WithContext(ctx).Model(&entity.CloudFile{}).Where("id = ?", file.ID).Update("deleted_at", time.Now())

	repo := NewListCloudRepositoryRepository(db, "test-bucket")
	filter := request.ListFilesRequestDTO{Page: 1, PageSize: 20}

	files, total, err := repo.GetFilesByUserID(ctx, userID, filter)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, total, int64(1))
	assert.GreaterOrEqual(t, len(files), 1)
	found := false
	for _, f := range files {
		if f.ID == file.ID {
			found = true
			break
		}
	}
	assert.True(t, found, "created file should appear in list")
	t.Logf("GetFilesByUserID: found %d files, total=%d", len(files), total)
}

func TestListCloudRepository_GetFilesByUserID_Empty(t *testing.T) {
	db := setupTestDB(t)
	ctx := context.Background()

	repo := NewListCloudRepositoryRepository(db, "test-bucket")
	filter := request.ListFilesRequestDTO{Page: 1, PageSize: 20}

	files, total, err := repo.GetFilesByUserID(ctx, 999999, filter)
	require.NoError(t, err)
	assert.Equal(t, int64(0), total)
	assert.Len(t, files, 0)
	t.Logf("GetFilesByUserID: empty result for non-existent user")
}
