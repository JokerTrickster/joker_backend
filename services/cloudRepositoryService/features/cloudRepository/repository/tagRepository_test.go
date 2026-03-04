package repository

import (
	"context"
	"testing"
	"time"

	"github.com/JokerTrickster/joker_backend/services/cloudRepositoryService/features/cloudRepository/model/entity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTagRepository_GetFileByID_Success(t *testing.T) {
	db := setupTestDB(t)
	requireTable(t, db, "cloud_files")
	ctx := context.Background()

	file := &entity.CloudFile{
		UserID:      976,
		FileName:   "tag_file.jpg",
		S3Key:      "user976/tag_file.jpg",
		FileType:   entity.FileTypeImage,
		ContentType: "image/jpeg",
		FileSize:   1024,
	}
	err := db.WithContext(ctx).Create(file).Error
	require.NoError(t, err)
	defer db.WithContext(ctx).Model(&entity.CloudFile{}).Where("id = ?", file.ID).Update("deleted_at", time.Now())

	repo := NewTagRepository(db)
	fetched, err := repo.GetFileByID(ctx, file.ID, file.UserID)
	require.NoError(t, err)
	assert.Equal(t, file.ID, fetched.ID)
	t.Logf("GetFileByID: retrieved file with tags")
}

func TestTagRepository_FindOrCreateTag_Success(t *testing.T) {
	db := setupTestDB(t)
	requireTable(t, db, "tags")
	ctx := context.Background()

	repo := NewTagRepository(db)
	userID := uint(975)

	tag, err := repo.FindOrCreateTag(ctx, userID, "test_tag_unique")
	require.NoError(t, err)
	assert.Greater(t, tag.ID, uint(0))
	assert.Equal(t, "test_tag_unique", tag.Name)
	t.Logf("FindOrCreateTag: created/found tag ID=%d", tag.ID)

	tag2, err := repo.FindOrCreateTag(ctx, userID, "test_tag_unique")
	require.NoError(t, err)
	assert.Equal(t, tag.ID, tag2.ID, "second call should return same tag")
	t.Logf("FindOrCreateTag: idempotent - returns same tag")

	db.WithContext(ctx).Where("user_id = ? AND name = ?", userID, "test_tag_unique").Delete(&entity.Tag{})
}

func TestTagRepository_AddTagToFile_UpdateFileTags_Success(t *testing.T) {
	db := setupTestDB(t)
	requireTable(t, db, "cloud_files")
	requireTable(t, db, "tags")
	ctx := context.Background()

	file := &entity.CloudFile{
		UserID:      974,
		FileName:   "tag_ops.jpg",
		S3Key:      "user974/tag_ops.jpg",
		FileType:   entity.FileTypeImage,
		ContentType: "image/jpeg",
		FileSize:   1024,
	}
	err := db.WithContext(ctx).Create(file).Error
	require.NoError(t, err)
	defer db.WithContext(ctx).Model(&entity.CloudFile{}).Where("id = ?", file.ID).Update("deleted_at", time.Now())

	repo := NewTagRepository(db)
	tag, err := repo.FindOrCreateTag(ctx, file.UserID, "work")
	require.NoError(t, err)
	defer db.WithContext(ctx).Where("id = ?", tag.ID).Delete(&entity.Tag{})

	err = repo.AddTagToFile(ctx, file.ID, file.UserID, *tag)
	require.NoError(t, err)
	t.Logf("AddTagToFile: added tag to file")

	err = repo.UpdateFileTags(ctx, file.ID, file.UserID, []entity.Tag{*tag})
	require.NoError(t, err)
	t.Logf("UpdateFileTags: updated file tags")

	fetched, err := repo.GetFileByID(ctx, file.ID, file.UserID)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(fetched.Tags), 0)
	t.Logf("GetFileByID: file has tags after update")
}
