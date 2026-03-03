package repository

import (
	"context"
	"testing"
	"time"

	"github.com/JokerTrickster/joker_backend/services/cloudRepositoryService/features/cloudRepository/model/entity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFolderShareRepository_CreateFolderShare_Success(t *testing.T) {
	db := setupTestDB(t)
	ctx := context.Background()

	folder := &entity.Folder{UserID: 985, FolderName: "share_folder"}
	err := db.WithContext(ctx).Create(folder).Error
	require.NoError(t, err)
	defer db.WithContext(ctx).Model(&entity.Folder{}).Where("id = ?", folder.ID).Update("deleted_at", time.Now())

	repo := NewFolderShareRepository(db)
	share := &entity.FolderShare{
		FolderID:     folder.ID,
		OwnerID:      985,
		SharedWithID: 2,
		Permission:   entity.SharePermissionRead,
	}
	err = repo.CreateFolderShare(ctx, share)
	require.NoError(t, err)
	assert.Greater(t, share.ID, uint(0))
	t.Logf("CreateFolderShare: created share ID=%d", share.ID)

	db.WithContext(ctx).Where("folder_id = ? AND shared_with_id = ?", folder.ID, 2).Delete(&entity.FolderShare{})
}

func TestFolderShareRepository_GetFolderSharesByFolderID_Success(t *testing.T) {
	db := setupTestDB(t)
	ctx := context.Background()

	folder := &entity.Folder{UserID: 984, FolderName: "get_shares_folder"}
	err := db.WithContext(ctx).Create(folder).Error
	require.NoError(t, err)
	defer db.WithContext(ctx).Model(&entity.Folder{}).Where("id = ?", folder.ID).Update("deleted_at", time.Now())

	repo := NewFolderShareRepository(db)
	shares, err := repo.GetFolderSharesByFolderID(ctx, folder.ID)
	require.NoError(t, err)
	assert.NotNil(t, shares)
	t.Logf("GetFolderSharesByFolderID: returned %d shares", len(shares))
}

func TestFolderShareRepository_GetUsersByEmails_Success(t *testing.T) {
	db := setupTestDB(t)
	ctx := context.Background()

	repo := NewFolderShareRepository(db)
	users, err := repo.GetUsersByEmails(ctx, []string{"nonexistent@example.com"})
	require.NoError(t, err)
	assert.NotNil(t, users)
	t.Logf("GetUsersByEmails: returned %d users (may be 0 if no matching users)", len(users))
}

func TestFolderShareRepository_HasFolderAccess_Owner(t *testing.T) {
	db := setupTestDB(t)
	ctx := context.Background()

	folder := &entity.Folder{UserID: 983, FolderName: "access_folder"}
	err := db.WithContext(ctx).Create(folder).Error
	require.NoError(t, err)
	defer db.WithContext(ctx).Model(&entity.Folder{}).Where("id = ?", folder.ID).Update("deleted_at", time.Now())

	repo := NewFolderShareRepository(db)
	hasAccess, err := repo.HasFolderAccess(ctx, 983, folder.ID)
	require.NoError(t, err)
	assert.True(t, hasAccess, "owner should have access")
	t.Logf("HasFolderAccess: owner has access")
}
