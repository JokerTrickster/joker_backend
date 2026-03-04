package repository

import (
	"context"
	"os"
	"testing"

	"github.com/JokerTrickster/joker_backend/services/molandolanService/features/product/model/entity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func setupProductTestDB(t *testing.T) *gorm.DB {
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

func TestProductRepository_List(t *testing.T) {
	t.Skip("Integration test: requires test database")
	db := setupProductTestDB(t)
	ctx := context.Background()
	repo := NewProductRepository(db)

	items, total, err := repo.List(ctx, 1, 20, "", nil)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, total, int64(0))
	assert.GreaterOrEqual(t, len(items), 0)
	t.Logf("List: total=%d", total)
}

func TestProductRepository_FindByID(t *testing.T) {
	t.Skip("Integration test: requires test database")
	db := setupProductTestDB(t)
	ctx := context.Background()
	repo := NewProductRepository(db)

	p := &entity.Product{
		Name: "Repo Test", Price: 100, Description: "D", Image: "i", Category: "test", InStock: true,
	}
	err := db.WithContext(ctx).Create(p).Error
	require.NoError(t, err)
	defer db.WithContext(ctx).Delete(p)

	found, err := repo.FindByID(ctx, p.ID)
	require.NoError(t, err)
	require.NotNil(t, found)
	assert.Equal(t, p.ID, found.ID)
	t.Logf("FindByID: %+v", found)
}

func TestProductRepository_Create(t *testing.T) {
	t.Skip("Integration test: requires test database")
	db := setupProductTestDB(t)
	ctx := context.Background()
	repo := NewProductRepository(db)

	p := &entity.Product{
		Name: "Create Test", Price: 200, Description: "D", Image: "i", Category: "test", InStock: true,
	}
	created, err := repo.Create(ctx, p)
	require.NoError(t, err)
	require.NotNil(t, created)
	assert.Greater(t, created.ID, uint(0))
	db.WithContext(ctx).Delete(created)
}

func TestProductRepository_Update(t *testing.T) {
	t.Skip("Integration test: requires test database")
	db := setupProductTestDB(t)
	ctx := context.Background()
	repo := NewProductRepository(db)

	p := &entity.Product{Name: "Update Test", Price: 100, Description: "D", Image: "i", Category: "test", InStock: true}
	require.NoError(t, db.WithContext(ctx).Create(p).Error)
	defer db.WithContext(ctx).Delete(p)

	updated, err := repo.Update(ctx, p.ID, map[string]interface{}{"name": "Updated"})
	require.NoError(t, err)
	assert.Equal(t, "Updated", updated.Name)
}

func TestProductRepository_Delete(t *testing.T) {
	t.Skip("Integration test: requires test database")
	db := setupProductTestDB(t)
	ctx := context.Background()
	repo := NewProductRepository(db)

	p := &entity.Product{Name: "Delete Test", Price: 100, Description: "D", Image: "i", Category: "test", InStock: true}
	require.NoError(t, db.WithContext(ctx).Create(p).Error)

	require.NoError(t, repo.Delete(ctx, p.ID))
	_, err := repo.FindByID(ctx, p.ID)
	assert.Error(t, err)
}
