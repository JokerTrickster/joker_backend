package repository

import (
	"context"
	"testing"

	"github.com/JokerTrickster/joker_backend/services/cloudRepositoryService/features/cloudRepository/model/entity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMultipartUploadRepository_CreateMultipartUpload_SkipsWithoutS3(t *testing.T) {
	t.Skip("CreateMultipartUpload requires S3 - skip in integration test without AWS config")
}

func TestMultipartUploadRepository_CreateMultipartUploadRecord_Success(t *testing.T) {
	db := setupTestDB(t)
	requireTable(t, db, "multipart_uploads")
	ctx := context.Background()

	repo := NewMultipartUploadRepository(db, "test-bucket")

	upload := &entity.MultipartUpload{
		UserID:      979,
		UploadID:   "test-upload-id-record",
		FileKey:    "user979/large_file.mp4",
		FileName:   "large_file.mp4",
		FileSize:   100 * 1024 * 1024,
		ContentType: "video/mp4",
		PartSize:   5 * 1024 * 1024,
		TotalParts: 20,
		Status:     entity.MultipartUploadStatusInitiated,
	}

	err := repo.CreateMultipartUploadRecord(ctx, upload)
	require.NoError(t, err)
	assert.Greater(t, upload.ID, uint(0))
	t.Logf("CreateMultipartUploadRecord: created record ID=%d", upload.ID)

	db.WithContext(ctx).Where("upload_id = ?", upload.UploadID).Delete(&entity.MultipartUpload{})
}

func TestMultipartUploadRepository_GetMultipartUpload_Success(t *testing.T) {
	db := setupTestDB(t)
	requireTable(t, db, "multipart_uploads")
	ctx := context.Background()

	upload := &entity.MultipartUpload{
		UserID:      978,
		UploadID:   "test-get-upload-id",
		FileKey:    "user978/get_test.mp4",
		FileName:   "get_test.mp4",
		FileSize:   50 * 1024 * 1024,
		ContentType: "video/mp4",
		PartSize:   5 * 1024 * 1024,
		TotalParts: 10,
		Status:     entity.MultipartUploadStatusInitiated,
	}
	err := db.WithContext(ctx).Create(upload).Error
	require.NoError(t, err)
	defer db.WithContext(ctx).Where("upload_id = ?", upload.UploadID).Delete(&entity.MultipartUpload{})

	repo := NewMultipartUploadRepository(db, "test-bucket")
	fetched, err := repo.GetMultipartUpload(ctx, upload.UploadID, upload.UserID)
	require.NoError(t, err)
	assert.Equal(t, upload.UploadID, fetched.UploadID)
	assert.Equal(t, upload.FileKey, fetched.FileKey)
	t.Logf("GetMultipartUpload: retrieved record successfully")
}

func TestMultipartUploadRepository_UpdateMultipartUploadStatus_Success(t *testing.T) {
	db := setupTestDB(t)
	requireTable(t, db, "multipart_uploads")
	ctx := context.Background()

	upload := &entity.MultipartUpload{
		UserID:      977,
		UploadID:   "test-status-upload",
		FileKey:    "user977/status_test.mp4",
		FileName:   "status_test.mp4",
		FileSize:   10 * 1024 * 1024,
		ContentType: "video/mp4",
		PartSize:   5 * 1024 * 1024,
		TotalParts: 2,
		Status:     entity.MultipartUploadStatusInitiated,
	}
	err := db.WithContext(ctx).Create(upload).Error
	require.NoError(t, err)
	defer db.WithContext(ctx).Where("upload_id = ?", upload.UploadID).Delete(&entity.MultipartUpload{})

	repo := NewMultipartUploadRepository(db, "test-bucket")
	err = repo.UpdateMultipartUploadStatus(ctx, upload.UploadID, entity.MultipartUploadStatusUploading)
	require.NoError(t, err)
	t.Logf("UpdateMultipartUploadStatus: updated successfully")

	fetched, err := repo.GetMultipartUpload(ctx, upload.UploadID, upload.UserID)
	require.NoError(t, err)
	assert.Equal(t, entity.MultipartUploadStatusUploading, fetched.Status)
}
