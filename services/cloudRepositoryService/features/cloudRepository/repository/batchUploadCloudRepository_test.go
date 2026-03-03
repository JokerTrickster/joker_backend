package repository

import (
	"context"
	"testing"
	"time"

	"github.com/JokerTrickster/joker_backend/services/cloudRepositoryService/features/cloudRepository/model/entity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBatchUploadCloudRepository_GeneratePresignedUploadURL_SkipsWithoutS3(t *testing.T) {
	t.Skip("GeneratePresignedUploadURL requires S3 - skip in integration test without AWS config")
}

func TestBatchUploadCloudRepository_CreateFile_Success(t *testing.T) {
	db := setupTestDB(t)
	ctx := context.Background()

	repo := NewBatchUploadCloudRepositoryRepository(db, "test-bucket")

	file := &entity.CloudFile{
		UserID:      994,
		FileName:   "batch_upload_test.jpg",
		S3Key:      "user994/batch_upload_test.jpg",
		FileType:   entity.FileTypeImage,
		ContentType: "image/jpeg",
		FileSize:   512,
	}

	err := repo.CreateFile(ctx, file)
	require.NoError(t, err)
	assert.Greater(t, file.ID, uint(0))
	t.Logf("CreateFile: created file ID=%d", file.ID)

	db.WithContext(ctx).Model(&entity.CloudFile{}).Where("id = ?", file.ID).Update("deleted_at", time.Now())
}
