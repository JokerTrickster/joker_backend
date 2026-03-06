package repository

import (
	"context"
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/JokerTrickster/joker_backend/services/cloudRepositoryService/features/cloudRepository/model/entity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUploadCloudRepository_PresignedURL_RequiresS3(t *testing.T) {
	t.Skip("GeneratePresignedUploadURL requires S3 - skip in integration test without AWS config")
}

func TestUploadCloudRepository_CreateFile_Success(t *testing.T) {
	t.Logf("TestUploadCloudRepository_CreateFile_Success: create file, verify in DB")

	db := setupTestDB(t)
	ctx := context.Background()

	require.NoError(t, db.AutoMigrate(&entity.CloudFile{}))
	defer func() {
		_ = db.WithContext(ctx).Where("user_id = ?", uint(993)).Delete(&entity.CloudFile{}).Error
	}()

	uid := fmt.Sprintf("%d_%d", time.Now().UnixNano(), rand.Intn(100000))
	file := &entity.CloudFile{
		UserID:      993,
		FileName:   "test_upload_create.jpg",
		S3Key:      "user993/test_upload_create_" + uid + ".jpg",
		FileType:   entity.FileTypeImage,
		ContentType: "image/jpeg",
		FileSize:   4096,
	}

	repo := NewUploadCloudRepositoryRepository(db, "test-bucket")
	err := repo.CreateFile(ctx, file)
	require.NoError(t, err)
	t.Logf("CreateFile: created successfully")

	assert.NotZero(t, file.ID)
	assert.False(t, file.CreatedAt.IsZero())
	assert.False(t, file.UpdatedAt.IsZero())

	var fetched entity.CloudFile
	err = db.WithContext(ctx).Where("id = ?", file.ID).First(&fetched).Error
	require.NoError(t, err)
	assert.Equal(t, file.FileName, fetched.FileName)
	assert.Equal(t, file.S3Key, fetched.S3Key)
	assert.Equal(t, file.FileSize, fetched.FileSize)
	t.Logf("Success: file verified in DB")
}

func TestUploadCloudRepository_CreateFile_DuplicateID(t *testing.T) {
	t.Logf("TestUploadCloudRepository_CreateFile_DuplicateID: attempt to create with same primary key ID, expect error")

	db := setupTestDB(t)
	ctx := context.Background()

	require.NoError(t, db.AutoMigrate(&entity.CloudFile{}))
	defer func() {
		_ = db.WithContext(ctx).Where("user_id = ?", uint(994)).Delete(&entity.CloudFile{}).Error
	}()

	uid2 := fmt.Sprintf("%d_%d", time.Now().UnixNano(), rand.Intn(100000))
	file1 := &entity.CloudFile{
		UserID:      994,
		FileName:   "first.jpg",
		S3Key:      "user994/first_dup_" + uid2 + ".jpg",
		FileType:   entity.FileTypeImage,
		ContentType: "image/jpeg",
		FileSize:   100,
	}
	err := db.WithContext(ctx).Create(file1).Error
	require.NoError(t, err)
	fileID := file1.ID
	t.Logf("Setup: created first file ID=%d", fileID)

	// Attempt to create second file with same ID (explicit primary key)
	file2 := &entity.CloudFile{
		ID:          fileID,
		UserID:      994,
		FileName:   "second.jpg",
		S3Key:      "user994/second_dup_" + uid2 + ".jpg",
		FileType:   entity.FileTypeImage,
		ContentType: "image/jpeg",
		FileSize:   200,
	}

	repo := NewUploadCloudRepositoryRepository(db, "test-bucket")
	err = repo.CreateFile(ctx, file2)

	assert.Error(t, err)
	t.Logf("Success: CreateFile rejected duplicate primary key")
}
