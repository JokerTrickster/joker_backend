package repository

import (
	"context"
	"os"
	"testing"

	"github.com/JokerTrickster/joker_backend/services/molandolanService/features/ranking/model/entity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func setupRankingTestDB(t *testing.T) *gorm.DB {
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

func TestRankingRepository_List(t *testing.T) {
	t.Skip("Integration test: requires test database")
	db := setupRankingTestDB(t)
	ctx := context.Background()
	repo := NewRankingRepository(db)

	items, total, err := repo.List(ctx, "puzzle", 1, 5)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, total, int64(0))
	assert.GreaterOrEqual(t, len(items), 0)
	t.Logf("List: total=%d", total)
}

func TestRankingRepository_FindByUserAndGame(t *testing.T) {
	t.Skip("Integration test: requires test database")
	db := setupRankingTestDB(t)
	ctx := context.Background()
	repo := NewRankingRepository(db)

	r := &entity.Ranking{UserID: 1, GameType: "puzzle", Nickname: "Test", ClearTimeMs: 5000}
	require.NoError(t, db.WithContext(ctx).Create(r).Error)
	defer db.WithContext(ctx).Delete(r)

	found, err := repo.FindByUserAndGame(ctx, 1, "puzzle")
	require.NoError(t, err)
	require.NotNil(t, found)
	assert.Equal(t, uint(1), found.UserID)
	t.Logf("FindByUserAndGame: %+v", found)
}

func TestRankingRepository_Create(t *testing.T) {
	t.Skip("Integration test: requires test database")
	db := setupRankingTestDB(t)
	ctx := context.Background()
	repo := NewRankingRepository(db)

	r := &entity.Ranking{UserID: 1, GameType: "puzzle", Nickname: "CreateTest", ClearTimeMs: 6000}
	err := repo.Create(ctx, r)
	require.NoError(t, err)
	assert.Greater(t, r.ID, uint(0))
	db.WithContext(ctx).Delete(r)
}

func TestRankingRepository_Update(t *testing.T) {
	t.Skip("Integration test: requires test database")
	db := setupRankingTestDB(t)
	ctx := context.Background()
	repo := NewRankingRepository(db)

	r := &entity.Ranking{UserID: 1, GameType: "puzzle", Nickname: "UpdateTest", ClearTimeMs: 7000}
	require.NoError(t, db.WithContext(ctx).Create(r).Error)
	defer db.WithContext(ctx).Delete(r)

	r.ClearTimeMs = 6500
	err := repo.Update(ctx, r)
	require.NoError(t, err)

	found, _ := repo.FindByUserAndGame(ctx, 1, "puzzle")
	assert.Equal(t, uint(6500), found.ClearTimeMs)
}

func TestRankingRepository_Delete(t *testing.T) {
	t.Skip("Integration test: requires test database")
	db := setupRankingTestDB(t)
	ctx := context.Background()
	repo := NewRankingRepository(db)

	r := &entity.Ranking{UserID: 1, GameType: "puzzle", Nickname: "DeleteTest", ClearTimeMs: 8000}
	require.NoError(t, db.WithContext(ctx).Create(r).Error)

	require.NoError(t, repo.Delete(ctx, r.ID))
	_, err := repo.FindByUserAndGame(ctx, 1, "puzzle")
	assert.Error(t, err)
}

func TestRankingRepository_GetRank(t *testing.T) {
	t.Skip("Integration test: requires test database")
	db := setupRankingTestDB(t)
	ctx := context.Background()
	repo := NewRankingRepository(db)

	rank, err := repo.GetRank(ctx, "puzzle", 5000)
	require.NoError(t, err)
	assert.Greater(t, rank, 0)
	t.Logf("GetRank: %d", rank)
}
