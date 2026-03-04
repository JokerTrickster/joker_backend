package repository

import (
	"context"
	"testing"
	"time"

	"github.com/JokerTrickster/joker_backend/services/cloudRepositoryService/features/cloudRepository/model/entity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFolderRepository_CreateFolder_Success(t *testing.T) {
	db := setupTestDB(t)
	requireTable(t, db, "folders")
	ctx := context.Background()

	repo := NewFolderRepository(db)

	folder := &entity.Folder{
		UserID:     989,
		FolderName: "test_folder_create",
	}

	err := repo.CreateFolder(ctx, folder)
	require.NoError(t, err)
	assert.Greater(t, folder.ID, uint(0))
	t.Logf("CreateFolder: created folder ID=%d", folder.ID)

	defer db.WithContext(ctx).Model(&entity.Folder{}).Where("id = ?", folder.ID).Update("deleted_at", time.Now())
}

func TestFolderRepository_GetFolderByID_Success(t *testing.T) {
	db := setupTestDB(t)
	requireTable(t, db, "folders")
	ctx := context.Background()

	folder := &entity.Folder{UserID: 988, FolderName: "test_get_folder"}
	err := db.WithContext(ctx).Create(folder).Error
	require.NoError(t, err)
	defer db.WithContext(ctx).Model(&entity.Folder{}).Where("id = ?", folder.ID).Update("deleted_at", time.Now())

	repo := NewFolderRepository(db)
	fetched, err := repo.GetFolderByID(ctx, folder.ID, int32(folder.UserID))
	require.NoError(t, err)
	assert.Equal(t, folder.ID, fetched.ID)
	assert.Equal(t, folder.FolderName, fetched.FolderName)
	t.Logf("GetFolderByID: retrieved folder successfully")
}

func TestFolderRepository_GetFoldersByUserID_Success(t *testing.T) {
	db := setupTestDB(t)
	requireTable(t, db, "folders")
	ctx := context.Background()

	userID := int32(987)
	folder := &entity.Folder{UserID: uint(userID), FolderName: "test_list_folder"}
	err := db.WithContext(ctx).Create(folder).Error
	require.NoError(t, err)
	defer db.WithContext(ctx).Model(&entity.Folder{}).Where("id = ?", folder.ID).Update("deleted_at", time.Now())

	repo := NewFolderRepository(db)
	folders, err := repo.GetFoldersByUserID(ctx, userID)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(folders), 1)
	t.Logf("GetFoldersByUserID: found %d folders", len(folders))
}

func TestFolderRepository_DeleteFolder_Success(t *testing.T) {
	db := setupTestDB(t)
	requireTable(t, db, "folders")
	ctx := context.Background()

	folder := &entity.Folder{UserID: 986, FolderName: "test_delete_folder"}
	err := db.WithContext(ctx).Create(folder).Error
	require.NoError(t, err)

	repo := NewFolderRepository(db)
	err = repo.DeleteFolder(ctx, folder.ID, int32(folder.UserID))
	require.NoError(t, err)
	t.Logf("DeleteFolder: soft deleted successfully")

	var deleted entity.Folder
	err = db.WithContext(ctx).Unscoped().Where("id = ?", folder.ID).First(&deleted).Error
	require.NoError(t, err)
	assert.NotNil(t, deleted.DeletedAt)
}
