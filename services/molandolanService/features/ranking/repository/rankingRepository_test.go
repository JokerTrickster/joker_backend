package repository

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"testing"
	"time"

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
		t.Skipf("Integration test: requires test database: %v", err)
	}
	if err := db.AutoMigrate(&entity.Ranking{}); err != nil {
		t.Skipf("Integration test: migration failed: %v", err)
	}
	return db
}

func uniqueID() string {
	return fmt.Sprintf("%d_%d", time.Now().UnixNano(), rand.Intn(100000))
}

func shortGameType() string {
	return fmt.Sprintf("g%d", rand.Intn(99999))
}

func TestRankingRepository_Create(t *testing.T) {
	db := setupRankingTestDB(t)
	ctx := context.Background()
	repo := NewRankingRepository(db)

	gameType := shortGameType()
	userID := uint(rand.Intn(100000) + 1)
	r := &entity.Ranking{
		UserID:      userID,
		GameType:    gameType,
		Nickname:    "CreateTest_" + uniqueID(),
		ClearTimeMs: 5000,
	}
	err := repo.Create(ctx, r)
	require.NoError(t, err)
	assert.Greater(t, r.ID, uint(0))
	t.Logf("Create: id=%d userID=%d gameType=%s", r.ID, r.UserID, r.GameType)

	defer func() {
		db.WithContext(ctx).Delete(r)
	}()
}

func TestRankingRepository_List(t *testing.T) {
	db := setupRankingTestDB(t)
	ctx := context.Background()
	repo := NewRankingRepository(db)

	gameType := shortGameType()
	items, total, err := repo.List(ctx, gameType, 1, 20)
	require.NoError(t, err)
	assert.NotNil(t, items)
	assert.Equal(t, int64(0), total)
	assert.Len(t, items, 0)
	t.Logf("List empty gameType: total=%d", total)
}

func TestRankingRepository_List_WithData(t *testing.T) {
	db := setupRankingTestDB(t)
	ctx := context.Background()
	repo := NewRankingRepository(db)

	gameType := shortGameType()
	r1 := &entity.Ranking{UserID: 1001, GameType: gameType, Nickname: "Player1", ClearTimeMs: 3000}
	r2 := &entity.Ranking{UserID: 1002, GameType: gameType, Nickname: "Player2", ClearTimeMs: 5000}
	r3 := &entity.Ranking{UserID: 1003, GameType: gameType, Nickname: "Player3", ClearTimeMs: 7000}
	require.NoError(t, db.WithContext(ctx).Create(r1).Error)
	require.NoError(t, db.WithContext(ctx).Create(r2).Error)
	require.NoError(t, db.WithContext(ctx).Create(r3).Error)
	defer func() {
		db.WithContext(ctx).Delete(r1)
		db.WithContext(ctx).Delete(r2)
		db.WithContext(ctx).Delete(r3)
	}()

	items, total, err := repo.List(ctx, gameType, 1, 10)
	require.NoError(t, err)
	assert.NotNil(t, items)
	assert.Equal(t, int64(3), total)
	assert.Len(t, items, 3)
	assert.Equal(t, uint(3000), items[0].ClearTimeMs, "List should be ordered by clear_time_ms ASC")
	assert.Equal(t, uint(5000), items[1].ClearTimeMs)
	assert.Equal(t, uint(7000), items[2].ClearTimeMs)
	t.Logf("List with data: total=%d", total)
}

func TestRankingRepository_List_Pagination(t *testing.T) {
	db := setupRankingTestDB(t)
	ctx := context.Background()
	repo := NewRankingRepository(db)

	gameType := shortGameType()
	for i := 0; i < 5; i++ {
		r := &entity.Ranking{UserID: uint(2000 + i), GameType: gameType, Nickname: fmt.Sprintf("P%d", i), ClearTimeMs: uint(1000 * (i + 1))}
		require.NoError(t, db.WithContext(ctx).Create(r).Error)
		defer db.WithContext(ctx).Delete(r)
	}

	items, total, err := repo.List(ctx, gameType, 1, 2)
	require.NoError(t, err)
	assert.NotNil(t, items)
	assert.Equal(t, int64(5), total)
	assert.Len(t, items, 2)
	t.Logf("List pagination: page=1 limit=2 returned %d of %d", len(items), total)
}

func TestRankingRepository_FindByUserAndGame(t *testing.T) {
	db := setupRankingTestDB(t)
	ctx := context.Background()
	repo := NewRankingRepository(db)

	userID := uint(3000 + rand.Intn(1000))
	gameType := shortGameType()
	r := &entity.Ranking{UserID: userID, GameType: gameType, Nickname: "FindTest", ClearTimeMs: 6000}
	require.NoError(t, db.WithContext(ctx).Create(r).Error)
	defer func() { db.WithContext(ctx).Delete(r) }()

	found, err := repo.FindByUserAndGame(ctx, userID, gameType)
	require.NoError(t, err)
	require.NotNil(t, found)
	assert.Equal(t, userID, found.UserID)
	assert.Equal(t, gameType, found.GameType)
	assert.Equal(t, uint(6000), found.ClearTimeMs)
	t.Logf("FindByUserAndGame: id=%d userID=%d", found.ID, found.UserID)
}

func TestRankingRepository_FindByUserAndGame_NotFound(t *testing.T) {
	db := setupRankingTestDB(t)
	ctx := context.Background()
	repo := NewRankingRepository(db)

	found, err := repo.FindByUserAndGame(ctx, 999999, "nonexistent_game")
	require.Error(t, err)
	assert.Nil(t, found)
	assert.ErrorIs(t, err, gorm.ErrRecordNotFound)
	t.Log("FindByUserAndGame correctly returns error for non-existent record")
}

func TestRankingRepository_Update(t *testing.T) {
	db := setupRankingTestDB(t)
	ctx := context.Background()
	repo := NewRankingRepository(db)

	gameType := shortGameType()
	userID := uint(4000)
	r := &entity.Ranking{UserID: userID, GameType: gameType, Nickname: "UpdateTest", ClearTimeMs: 7000}
	require.NoError(t, db.WithContext(ctx).Create(r).Error)
	defer func() { db.WithContext(ctx).Delete(r) }()

	r.ClearTimeMs = 6500
	err := repo.Update(ctx, r)
	require.NoError(t, err)

	found, err := repo.FindByUserAndGame(ctx, userID, gameType)
	require.NoError(t, err)
	require.NotNil(t, found)
	assert.Equal(t, uint(6500), found.ClearTimeMs)
	t.Logf("Update: ClearTimeMs changed to %d", found.ClearTimeMs)
}

func TestRankingRepository_Delete(t *testing.T) {
	db := setupRankingTestDB(t)
	ctx := context.Background()
	repo := NewRankingRepository(db)

	gameType := shortGameType()
	userID := uint(5000)
	r := &entity.Ranking{UserID: userID, GameType: gameType, Nickname: "DeleteTest", ClearTimeMs: 8000}
	require.NoError(t, db.WithContext(ctx).Create(r).Error)

	err := repo.Delete(ctx, r.ID)
	require.NoError(t, err)

	_, err = repo.FindByUserAndGame(ctx, userID, gameType)
	require.Error(t, err)
	assert.ErrorIs(t, err, gorm.ErrRecordNotFound)
	t.Log("Delete: record removed successfully")
}

func TestRankingRepository_Delete_NotFound(t *testing.T) {
	db := setupRankingTestDB(t)
	ctx := context.Background()
	repo := NewRankingRepository(db)

	err := repo.Delete(ctx, 999999)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "ranking not found")
	t.Log("Delete correctly returns error for non-existent id")
}

func TestRankingRepository_GetRank(t *testing.T) {
	db := setupRankingTestDB(t)
	ctx := context.Background()
	repo := NewRankingRepository(db)

	gameType := shortGameType()
	r1 := &entity.Ranking{UserID: 6001, GameType: gameType, Nickname: "R1", ClearTimeMs: 2000}
	r2 := &entity.Ranking{UserID: 6002, GameType: gameType, Nickname: "R2", ClearTimeMs: 4000}
	r3 := &entity.Ranking{UserID: 6003, GameType: gameType, Nickname: "R3", ClearTimeMs: 6000}
	require.NoError(t, db.WithContext(ctx).Create(r1).Error)
	require.NoError(t, db.WithContext(ctx).Create(r2).Error)
	require.NoError(t, db.WithContext(ctx).Create(r3).Error)
	defer func() {
		db.WithContext(ctx).Delete(r1)
		db.WithContext(ctx).Delete(r2)
		db.WithContext(ctx).Delete(r3)
	}()

	rank1000, err := repo.GetRank(ctx, gameType, 1000)
	require.NoError(t, err)
	assert.Equal(t, 1, rank1000, "1000ms: no one faster, rank 1")

	rank3000, err := repo.GetRank(ctx, gameType, 3000)
	require.NoError(t, err)
	assert.Equal(t, 2, rank3000, "3000ms: 1 person faster (2000), rank 2")

	rank5000, err := repo.GetRank(ctx, gameType, 5000)
	require.NoError(t, err)
	assert.Equal(t, 3, rank5000, "5000ms: 2 faster, rank 3")

	rank7000, err := repo.GetRank(ctx, gameType, 7000)
	require.NoError(t, err)
	assert.Equal(t, 4, rank7000, "7000ms: 3 faster, rank 4")

	rank9999, err := repo.GetRank(ctx, gameType, 9999)
	require.NoError(t, err)
	assert.Equal(t, 4, rank9999, "9999ms: same 3 faster, rank 4")

	t.Logf("GetRank: verified rank calculation for gameType=%s", gameType)
}

func TestRankingRepository_GetRank_EmptyGameType(t *testing.T) {
	db := setupRankingTestDB(t)
	ctx := context.Background()
	repo := NewRankingRepository(db)

	gameType := shortGameType()
	rank, err := repo.GetRank(ctx, gameType, 5000)
	require.NoError(t, err)
	assert.Equal(t, 1, rank, "no rankings: rank should be 1")
	t.Log("GetRank for empty gameType returns 1")
}
