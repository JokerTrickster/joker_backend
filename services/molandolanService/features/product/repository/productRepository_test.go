package repository

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"testing"
	"time"

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
		t.Skipf("Integration test: requires test database: %v", err)
	}
	if err := db.AutoMigrate(&entity.Product{}); err != nil {
		t.Skipf("Integration test: migration failed: %v", err)
	}
	return db
}

func uniqueID() string {
	return fmt.Sprintf("%d_%d", time.Now().UnixNano(), rand.Intn(100000))
}

func TestProductRepository_Create(t *testing.T) {
	db := setupProductTestDB(t)
	ctx := context.Background()
	repo := NewProductRepository(db)

	name := "Create_" + uniqueID()
	p := &entity.Product{
		Name:        name,
		Price:       100,
		Description: "Desc " + uniqueID(),
		Image:       "https://example.com/img.png",
		Category:    "test",
		InStock:     true,
	}
	created, err := repo.Create(ctx, p)
	require.NoError(t, err)
	require.NotNil(t, created)
	assert.Greater(t, created.ID, uint(0))
	assert.Equal(t, name, created.Name)
	t.Logf("Create: id=%d name=%s", created.ID, created.Name)

	defer func() {
		db.WithContext(ctx).Unscoped().Delete(created)
	}()
}

func TestProductRepository_List(t *testing.T) {
	db := setupProductTestDB(t)
	ctx := context.Background()
	repo := NewProductRepository(db)

	items, total, err := repo.List(ctx, 1, 20, "", nil)
	require.NoError(t, err)
	assert.NotNil(t, items)
	assert.GreaterOrEqual(t, total, int64(0))
	t.Logf("List: total=%d items=%d", total, len(items))
}

func TestProductRepository_List_WithCategory(t *testing.T) {
	db := setupProductTestDB(t)
	ctx := context.Background()
	repo := NewProductRepository(db)

	category := "cat_" + uniqueID()
	p1 := &entity.Product{
		Name: "ListCat1_" + uniqueID(), Price: 100, Description: "D1", Image: "i1", Category: category, InStock: true,
	}
	p2 := &entity.Product{
		Name: "ListCat2_" + uniqueID(), Price: 200, Description: "D2", Image: "i2", Category: category, InStock: true,
	}
	pOther := &entity.Product{
		Name: "OtherCat_" + uniqueID(), Price: 300, Description: "D3", Image: "i3", Category: "other", InStock: true,
	}
	require.NoError(t, db.WithContext(ctx).Create(p1).Error)
	require.NoError(t, db.WithContext(ctx).Create(p2).Error)
	require.NoError(t, db.WithContext(ctx).Create(pOther).Error)
	defer func() {
		db.WithContext(ctx).Unscoped().Delete(p1)
		db.WithContext(ctx).Unscoped().Delete(p2)
		db.WithContext(ctx).Unscoped().Delete(pOther)
	}()

	items, total, err := repo.List(ctx, 1, 10, category, nil)
	require.NoError(t, err)
	assert.NotNil(t, items)
	assert.Equal(t, int64(2), total)
	assert.Len(t, items, 2)
	for _, item := range items {
		assert.Equal(t, category, item.Category)
	}
	t.Logf("List with category %s: total=%d", category, total)
}

func TestProductRepository_List_WithInStock(t *testing.T) {
	db := setupProductTestDB(t)
	ctx := context.Background()
	repo := NewProductRepository(db)

	inStockTrue := true
	inStockFalse := false

	itemsInStock, totalInStock, err := repo.List(ctx, 1, 20, "", &inStockTrue)
	require.NoError(t, err)
	assert.NotNil(t, itemsInStock)
	assert.GreaterOrEqual(t, totalInStock, int64(0))
	for _, item := range itemsInStock {
		assert.True(t, item.InStock)
	}

	itemsOutOfStock, totalOutOfStock, err := repo.List(ctx, 1, 20, "", &inStockFalse)
	require.NoError(t, err)
	assert.NotNil(t, itemsOutOfStock)
	assert.GreaterOrEqual(t, totalOutOfStock, int64(0))
	for _, item := range itemsOutOfStock {
		assert.False(t, item.InStock)
	}
	t.Logf("List inStock=true: %d, inStock=false: %d", totalInStock, totalOutOfStock)
}

func TestProductRepository_List_Pagination(t *testing.T) {
	db := setupProductTestDB(t)
	ctx := context.Background()
	repo := NewProductRepository(db)

	items, total, err := repo.List(ctx, 1, 2, "", nil)
	require.NoError(t, err)
	assert.NotNil(t, items)
	assert.GreaterOrEqual(t, total, int64(0))
	assert.LessOrEqual(t, len(items), 2)
	t.Logf("List page=1 limit=2: items=%d total=%d", len(items), total)
}

func TestProductRepository_FindByID(t *testing.T) {
	db := setupProductTestDB(t)
	ctx := context.Background()
	repo := NewProductRepository(db)

	name := "FindByID_" + uniqueID()
	p := &entity.Product{
		Name: name, Price: 100, Description: "D", Image: "i", Category: "test", InStock: true,
	}
	require.NoError(t, db.WithContext(ctx).Create(p).Error)
	defer func() { db.WithContext(ctx).Unscoped().Delete(p) }()

	found, err := repo.FindByID(ctx, p.ID)
	require.NoError(t, err)
	require.NotNil(t, found)
	assert.Equal(t, p.ID, found.ID)
	assert.Equal(t, name, found.Name)
	t.Logf("FindByID: id=%d name=%s", found.ID, found.Name)
}

func TestProductRepository_FindByID_NotFound(t *testing.T) {
	db := setupProductTestDB(t)
	ctx := context.Background()
	repo := NewProductRepository(db)

	found, err := repo.FindByID(ctx, 999999)
	require.Error(t, err)
	assert.Nil(t, found)
	assert.Contains(t, err.Error(), "product not found")
	t.Log("FindByID correctly returns error for non-existent id")
}

func TestProductRepository_Update(t *testing.T) {
	db := setupProductTestDB(t)
	ctx := context.Background()
	repo := NewProductRepository(db)

	p := &entity.Product{
		Name: "Update_" + uniqueID(), Price: 100, Description: "D", Image: "i", Category: "test", InStock: true,
	}
	require.NoError(t, db.WithContext(ctx).Create(p).Error)
	defer func() { db.WithContext(ctx).Unscoped().Delete(p) }()

	newName := "Updated_" + uniqueID()
	updated, err := repo.Update(ctx, p.ID, map[string]interface{}{"name": newName})
	require.NoError(t, err)
	require.NotNil(t, updated)
	assert.Equal(t, newName, updated.Name)
	t.Logf("Update: id=%d new name=%s", updated.ID, updated.Name)
}

func TestProductRepository_Update_NotFound(t *testing.T) {
	db := setupProductTestDB(t)
	ctx := context.Background()
	repo := NewProductRepository(db)

	updated, err := repo.Update(ctx, 999999, map[string]interface{}{"name": "x"})
	require.Error(t, err)
	assert.Nil(t, updated)
	assert.Contains(t, err.Error(), "product not found")
	t.Log("Update correctly returns error for non-existent id")
}

func TestProductRepository_Delete(t *testing.T) {
	db := setupProductTestDB(t)
	ctx := context.Background()
	repo := NewProductRepository(db)

	p := &entity.Product{
		Name: "Delete_" + uniqueID(), Price: 100, Description: "D", Image: "i", Category: "test", InStock: true,
	}
	require.NoError(t, db.WithContext(ctx).Create(p).Error)

	err := repo.Delete(ctx, p.ID)
	require.NoError(t, err)

	_, err = repo.FindByID(ctx, p.ID)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "product not found")
	t.Log("Delete: record removed successfully")
}

func TestProductRepository_Delete_NotFound(t *testing.T) {
	db := setupProductTestDB(t)
	ctx := context.Background()
	repo := NewProductRepository(db)

	err := repo.Delete(ctx, 999999)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "product not found")
	t.Log("Delete correctly returns error for non-existent id")
}
