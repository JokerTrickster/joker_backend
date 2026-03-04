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

func TestFavoriteRepository_AddFavorite_Success(t *testing.T) {
	db := setupTestDB(t)
	requireTable(t, db, "cloud_files")
	requireTable(t, db, "favorites")
	ctx := context.Background()

	userID := uint(993)
	file := &entity.CloudFile{
		UserID:      userID,
		FileName:   "fav_test.jpg",
		S3Key:      "user993/fav_test.jpg",
		FileType:   entity.FileTypeImage,
		ContentType: "image/jpeg",
		FileSize:   1024,
	}
	err := db.WithContext(ctx).Create(file).Error
	require.NoError(t, err)
	defer db.WithContext(ctx).Model(&entity.CloudFile{}).Where("id = ?", file.ID).Update("deleted_at", time.Now())

	repo := NewFavoriteRepository(db)
	fav, err := repo.AddFavorite(ctx, userID, file.ID)
	require.NoError(t, err)
	assert.NotNil(t, fav)
	assert.Equal(t, userID, fav.UserID)
	assert.Equal(t, file.ID, fav.FileID)
	assert.False(t, fav.FavoritedAt.IsZero())
	t.Logf("AddFavorite: created favorite for file ID=%d", file.ID)

	db.WithContext(ctx).Where("user_id = ? AND file_id = ?", userID, file.ID).Delete(&entity.Favorite{})
}

func TestFavoriteRepository_RemoveFavorite_Success(t *testing.T) {
	db := setupTestDB(t)
	requireTable(t, db, "cloud_files")
	requireTable(t, db, "favorites")
	ctx := context.Background()

	userID := uint(992)
	file := &entity.CloudFile{
		UserID:      userID,
		FileName:   "fav_remove.jpg",
		S3Key:      "user992/fav_remove.jpg",
		FileType:   entity.FileTypeImage,
		ContentType: "image/jpeg",
		FileSize:   1024,
	}
	err := db.WithContext(ctx).Create(file).Error
	require.NoError(t, err)
	defer db.WithContext(ctx).Model(&entity.CloudFile{}).Where("id = ?", file.ID).Update("deleted_at", time.Now())

	repo := NewFavoriteRepository(db)
	_, err = repo.AddFavorite(ctx, userID, file.ID)
	require.NoError(t, err)

	err = repo.RemoveFavorite(ctx, userID, file.ID)
	require.NoError(t, err)
	t.Logf("RemoveFavorite: removed favorite successfully")

	isFav, err := repo.CheckIsFavorited(ctx, userID, file.ID)
	require.NoError(t, err)
	assert.False(t, isFav, "file should no longer be favorited")
}

func TestFavoriteRepository_GetFavoritesByUserID_Success(t *testing.T) {
	db := setupTestDB(t)
	requireTable(t, db, "cloud_files")
	requireTable(t, db, "favorites")
	ctx := context.Background()

	userID := uint(991)
	file := &entity.CloudFile{
		UserID:      userID,
		FileName:   "fav_list.jpg",
		S3Key:      "user991/fav_list.jpg",
		FileType:   entity.FileTypeImage,
		ContentType: "image/jpeg",
		FileSize:   1024,
	}
	err := db.WithContext(ctx).Create(file).Error
	require.NoError(t, err)
	defer db.WithContext(ctx).Model(&entity.CloudFile{}).Where("id = ?", file.ID).Update("deleted_at", time.Now())

	repo := NewFavoriteRepository(db)
	_, err = repo.AddFavorite(ctx, userID, file.ID)
	require.NoError(t, err)
	defer db.WithContext(ctx).Where("user_id = ? AND file_id = ?", userID, file.ID).Delete(&entity.Favorite{})

	filter := request.ListFavoritesRequestDTO{Page: 1, Size: 20}
	files, total, err := repo.GetFavoritesByUserID(ctx, userID, filter)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, total, int64(1))
	assert.GreaterOrEqual(t, len(files), 1)
	t.Logf("GetFavoritesByUserID: found %d favorites", len(files))
}

func TestFavoriteRepository_CheckIsFavorited(t *testing.T) {
	db := setupTestDB(t)
	requireTable(t, db, "cloud_files")
	requireTable(t, db, "favorites")
	ctx := context.Background()

	userID := uint(990)
	file := &entity.CloudFile{
		UserID:      userID,
		FileName:   "fav_check.jpg",
		S3Key:      "user990/fav_check.jpg",
		FileType:   entity.FileTypeImage,
		ContentType: "image/jpeg",
		FileSize:   1024,
	}
	err := db.WithContext(ctx).Create(file).Error
	require.NoError(t, err)
	defer db.WithContext(ctx).Model(&entity.CloudFile{}).Where("id = ?", file.ID).Update("deleted_at", time.Now())

	repo := NewFavoriteRepository(db)
	_, err = repo.AddFavorite(ctx, userID, file.ID)
	require.NoError(t, err)
	defer db.WithContext(ctx).Where("user_id = ? AND file_id = ?", userID, file.ID).Delete(&entity.Favorite{})

	isFav, err := repo.CheckIsFavorited(ctx, userID, file.ID)
	require.NoError(t, err)
	assert.True(t, isFav)

	isFav, err = repo.CheckIsFavorited(ctx, userID, 999999)
	require.NoError(t, err)
	assert.False(t, isFav)
	t.Logf("CheckIsFavorited: verified behavior")
}
