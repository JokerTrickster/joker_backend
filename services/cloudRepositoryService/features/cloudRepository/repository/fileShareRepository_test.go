package repository

import (
	"context"
	"testing"
	"time"

	"github.com/JokerTrickster/joker_backend/services/cloudRepositoryService/features/cloudRepository/model/entity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFileShareRepository_CreateFileShare_Success(t *testing.T) {
	db := setupTestDB(t)
	ctx := context.Background()

	file := &entity.CloudFile{
		UserID:      982,
		FileName:   "share_file.jpg",
		S3Key:      "user982/share_file.jpg",
		FileType:   entity.FileTypeImage,
		ContentType: "image/jpeg",
		FileSize:   1024,
	}
	err := db.WithContext(ctx).Create(file).Error
	require.NoError(t, err)
	defer db.WithContext(ctx).Model(&entity.CloudFile{}).Where("id = ?", file.ID).Update("deleted_at", time.Now())

	repo := NewFileShareRepository(db)
	share := &entity.FileShare{
		FileID:       file.ID,
		OwnerID:      982,
		SharedWithID: 3,
		Permission:   entity.SharePermissionRead,
	}
	err = repo.CreateFileShare(ctx, share)
	require.NoError(t, err)
	assert.Greater(t, share.ID, uint(0))
	t.Logf("CreateFileShare: created share ID=%d", share.ID)

	db.WithContext(ctx).Where("file_id = ? AND shared_with_id = ?", file.ID, 3).Delete(&entity.FileShare{})
}

func TestFileShareRepository_GetFileSharesByFileID_Success(t *testing.T) {
	db := setupTestDB(t)
	ctx := context.Background()

	repo := NewFileShareRepository(db)
	shares, err := repo.GetFileSharesByFileID(ctx, 999999)
	require.NoError(t, err)
	assert.NotNil(t, shares)
	assert.Len(t, shares, 0)
	t.Logf("GetFileSharesByFileID: returned empty for non-existent file")
}

func TestFileShareRepository_HasFileAccess_Owner(t *testing.T) {
	db := setupTestDB(t)
	ctx := context.Background()

	file := &entity.CloudFile{
		UserID:      981,
		FileName:   "access_file.jpg",
		S3Key:      "user981/access_file.jpg",
		FileType:   entity.FileTypeImage,
		ContentType: "image/jpeg",
		FileSize:   1024,
	}
	err := db.WithContext(ctx).Create(file).Error
	require.NoError(t, err)
	defer db.WithContext(ctx).Model(&entity.CloudFile{}).Where("id = ?", file.ID).Update("deleted_at", time.Now())

	repo := NewFileShareRepository(db)
	hasAccess, err := repo.HasFileAccess(ctx, 981, file.ID)
	require.NoError(t, err)
	assert.True(t, hasAccess, "owner should have access")
	t.Logf("HasFileAccess: owner has access")
}

func TestFileShareRepository_HasFileAccess_NoAccess(t *testing.T) {
	db := setupTestDB(t)
	ctx := context.Background()

	file := &entity.CloudFile{
		UserID:      980,
		FileName:   "noaccess_file.jpg",
		S3Key:      "user980/noaccess_file.jpg",
		FileType:   entity.FileTypeImage,
		ContentType: "image/jpeg",
		FileSize:   1024,
	}
	err := db.WithContext(ctx).Create(file).Error
	require.NoError(t, err)
	defer db.WithContext(ctx).Model(&entity.CloudFile{}).Where("id = ?", file.ID).Update("deleted_at", time.Now())

	repo := NewFileShareRepository(db)
	hasAccess, err := repo.HasFileAccess(ctx, 99, file.ID)
	require.NoError(t, err)
	assert.False(t, hasAccess, "other user should not have access")
	t.Logf("HasFileAccess: no access for non-owner without share")
}
