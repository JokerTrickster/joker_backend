package repository

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"testing"
	"time"

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
		t.Skipf("Integration test: requires test database: %v", err)
	}
	if err := db.AutoMigrate(&entity.News{}); err != nil {
		t.Skipf("Integration test: migration failed: %v", err)
	}
	return db
}

func uniqueID() string {
	return fmt.Sprintf("%d_%d", time.Now().UnixNano(), rand.Intn(100000))
}

func TestNewsRepository_Create(t *testing.T) {
	db := setupTestDB(t)
	ctx := context.Background()
	repo := NewNewsRepository(db)

	title := "Create_" + uniqueID()
	news := &entity.News{
		Title:    title,
		Summary:  "Summary " + uniqueID(),
		Content:  "Content " + uniqueID(),
		Category: "test",
		Date:     "2024-01-15",
	}
	created, err := repo.Create(ctx, news)
	require.NoError(t, err)
	require.NotNil(t, created)
	assert.Greater(t, created.ID, uint(0))
	assert.Equal(t, title, created.Title)
	t.Logf("Create: id=%d title=%s", created.ID, created.Title)

	defer func() {
		db.WithContext(ctx).Unscoped().Delete(created)
	}()
}

func TestNewsRepository_List(t *testing.T) {
	db := setupTestDB(t)
	ctx := context.Background()
	repo := NewNewsRepository(db)

	items, total, err := repo.List(ctx, 1, 20, "")
	require.NoError(t, err)
	assert.NotNil(t, items)
	assert.GreaterOrEqual(t, total, int64(0))
	t.Logf("List: total=%d items=%d", total, len(items))
}

func TestNewsRepository_List_WithCategory(t *testing.T) {
	db := setupTestDB(t)
	ctx := context.Background()
	repo := NewNewsRepository(db)

	category := "cat_" + uniqueID()
	news1 := &entity.News{
		Title: "ListCat1_" + uniqueID(), Summary: "S1", Content: "C1", Category: category, Date: "2024-01-15",
	}
	news2 := &entity.News{
		Title: "ListCat2_" + uniqueID(), Summary: "S2", Content: "C2", Category: category, Date: "2024-01-16",
	}
	newsOther := &entity.News{
		Title: "OtherCat_" + uniqueID(), Summary: "S3", Content: "C3", Category: "other", Date: "2024-01-17",
	}
	require.NoError(t, db.WithContext(ctx).Create(news1).Error)
	require.NoError(t, db.WithContext(ctx).Create(news2).Error)
	require.NoError(t, db.WithContext(ctx).Create(newsOther).Error)
	defer func() {
		db.WithContext(ctx).Unscoped().Delete(news1)
		db.WithContext(ctx).Unscoped().Delete(news2)
		db.WithContext(ctx).Unscoped().Delete(newsOther)
	}()

	items, total, err := repo.List(ctx, 1, 10, category)
	require.NoError(t, err)
	assert.NotNil(t, items)
	assert.Equal(t, int64(2), total)
	assert.Len(t, items, 2)
	for _, item := range items {
		assert.Equal(t, category, item.Category)
	}
	t.Logf("List with category %s: total=%d", category, total)
}

func TestNewsRepository_List_Pagination(t *testing.T) {
	db := setupTestDB(t)
	ctx := context.Background()
	repo := NewNewsRepository(db)

	items, total, err := repo.List(ctx, 1, 2, "")
	require.NoError(t, err)
	assert.NotNil(t, items)
	assert.GreaterOrEqual(t, total, int64(0))
	assert.LessOrEqual(t, len(items), 2)
	t.Logf("List page=1 limit=2: items=%d total=%d", len(items), total)
}

func TestNewsRepository_FindByID(t *testing.T) {
	db := setupTestDB(t)
	ctx := context.Background()
	repo := NewNewsRepository(db)

	title := "FindByID_" + uniqueID()
	news := &entity.News{
		Title: title, Summary: "S", Content: "C", Category: "test", Date: "2024-01-15",
	}
	require.NoError(t, db.WithContext(ctx).Create(news).Error)
	defer func() { db.WithContext(ctx).Unscoped().Delete(news) }()

	found, err := repo.FindByID(ctx, news.ID)
	require.NoError(t, err)
	require.NotNil(t, found)
	assert.Equal(t, news.ID, found.ID)
	assert.Equal(t, title, found.Title)
	t.Logf("FindByID: id=%d title=%s", found.ID, found.Title)
}

func TestNewsRepository_FindByID_NotFound(t *testing.T) {
	db := setupTestDB(t)
	ctx := context.Background()
	repo := NewNewsRepository(db)

	found, err := repo.FindByID(ctx, 999999)
	require.Error(t, err)
	assert.Nil(t, found)
	assert.Contains(t, err.Error(), "news not found")
	t.Log("FindByID correctly returns error for non-existent id")
}

func TestNewsRepository_Update(t *testing.T) {
	db := setupTestDB(t)
	ctx := context.Background()
	repo := NewNewsRepository(db)

	news := &entity.News{
		Title: "Update_" + uniqueID(), Summary: "S", Content: "C", Category: "test", Date: "2024-01-15",
	}
	require.NoError(t, db.WithContext(ctx).Create(news).Error)
	defer func() { db.WithContext(ctx).Unscoped().Delete(news) }()

	newTitle := "Updated_" + uniqueID()
	updated, err := repo.Update(ctx, news.ID, map[string]interface{}{"title": newTitle})
	require.NoError(t, err)
	require.NotNil(t, updated)
	assert.Equal(t, newTitle, updated.Title)
	t.Logf("Update: id=%d new title=%s", updated.ID, updated.Title)
}

func TestNewsRepository_Update_NotFound(t *testing.T) {
	db := setupTestDB(t)
	ctx := context.Background()
	repo := NewNewsRepository(db)

	updated, err := repo.Update(ctx, 999999, map[string]interface{}{"title": "x"})
	require.Error(t, err)
	assert.Nil(t, updated)
	assert.Contains(t, err.Error(), "news not found")
	t.Log("Update correctly returns error for non-existent id")
}

func TestNewsRepository_Delete(t *testing.T) {
	db := setupTestDB(t)
	ctx := context.Background()
	repo := NewNewsRepository(db)

	news := &entity.News{
		Title: "Delete_" + uniqueID(), Summary: "S", Content: "C", Category: "test", Date: "2024-01-15",
	}
	require.NoError(t, db.WithContext(ctx).Create(news).Error)

	err := repo.Delete(ctx, news.ID)
	require.NoError(t, err)

	_, err = repo.FindByID(ctx, news.ID)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "news not found")
	t.Log("Delete: record removed successfully")
}

func TestNewsRepository_Delete_NotFound(t *testing.T) {
	db := setupTestDB(t)
	ctx := context.Background()
	repo := NewNewsRepository(db)

	err := repo.Delete(ctx, 999999)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "news not found")
	t.Log("Delete correctly returns error for non-existent id")
}
