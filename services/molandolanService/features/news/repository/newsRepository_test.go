package repository

import (
	"context"
	"os"
	"testing"

	"github.com/JokerTrickster/joker_backend/services/molandolanService/features/news/model/entity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	dsn := os.Getenv("TEST_DB_DSN")
	if dsn == "" {
		dsn = "root:rootpassword@tcp(localhost:3307)/test_db?charset=utf8mb4&parseTime=True&loc=Local"
	}
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Skip("Skipping: test database unavailable:", err)
	}
	return db
}

func TestNewsRepository_List(t *testing.T) {
	t.Skip("Integration test: requires test database")
	db := setupTestDB(t)
	ctx := context.Background()
	repo := NewNewsRepository(db)

	items, total, err := repo.List(ctx, 1, 20, "")
	require.NoError(t, err)
	assert.GreaterOrEqual(t, total, int64(0))
	assert.GreaterOrEqual(t, len(items), 0)
	t.Logf("List: total=%d, items=%d", total, len(items))
}

func TestNewsRepository_FindByID(t *testing.T) {
	t.Skip("Integration test: requires test database")
	db := setupTestDB(t)
	ctx := context.Background()
	repo := NewNewsRepository(db)

	news := &entity.News{
		Title: "Repo Test", Summary: "Sum", Content: "Body", Category: "test", Date: "2024-01-15",
	}
	err := db.WithContext(ctx).Create(news).Error
	require.NoError(t, err)
	defer db.WithContext(ctx).Delete(news)

	found, err := repo.FindByID(ctx, news.ID)
	require.NoError(t, err)
	require.NotNil(t, found)
	assert.Equal(t, news.ID, found.ID)
	assert.Equal(t, "Repo Test", found.Title)
	t.Logf("FindByID: %+v", found)
}

func TestNewsRepository_Create(t *testing.T) {
	t.Skip("Integration test: requires test database")
	db := setupTestDB(t)
	ctx := context.Background()
	repo := NewNewsRepository(db)

	news := &entity.News{
		Title: "Create Test", Summary: "S", Content: "C", Category: "test", Date: "2024-01-15",
	}
	created, err := repo.Create(ctx, news)
	require.NoError(t, err)
	require.NotNil(t, created)
	assert.Greater(t, created.ID, uint(0))
	t.Logf("Create: id=%d", created.ID)
	db.WithContext(ctx).Delete(created)
}

func TestNewsRepository_Update(t *testing.T) {
	t.Skip("Integration test: requires test database")
	db := setupTestDB(t)
	ctx := context.Background()
	repo := NewNewsRepository(db)

	news := &entity.News{
		Title: "Update Test", Summary: "S", Content: "C", Category: "test", Date: "2024-01-15",
	}
	err := db.WithContext(ctx).Create(news).Error
	require.NoError(t, err)
	defer db.WithContext(ctx).Delete(news)

	updated, err := repo.Update(ctx, news.ID, map[string]interface{}{"title": "Updated Title"})
	require.NoError(t, err)
	require.NotNil(t, updated)
	assert.Equal(t, "Updated Title", updated.Title)
	t.Logf("Update: %+v", updated)
}

func TestNewsRepository_Delete(t *testing.T) {
	t.Skip("Integration test: requires test database")
	db := setupTestDB(t)
	ctx := context.Background()
	repo := NewNewsRepository(db)

	news := &entity.News{
		Title: "Delete Test", Summary: "S", Content: "C", Category: "test", Date: "2024-01-15",
	}
	err := db.WithContext(ctx).Create(news).Error
	require.NoError(t, err)

	err = repo.Delete(ctx, news.ID)
	require.NoError(t, err)

	_, err = repo.FindByID(ctx, news.ID)
	assert.Error(t, err)
	t.Log("Delete: record removed")
}
