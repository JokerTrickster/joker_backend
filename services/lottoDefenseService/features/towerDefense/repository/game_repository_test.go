package repository

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/JokerTrickster/joker_backend/services/lottoDefenseService/features/towerDefense/model/entity"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func setupGameTestDB(t *testing.T) *gorm.DB {
	dsn := os.Getenv("TEST_DB_DSN")
	if dsn == "" {
		dsn = "root:rootpassword@tcp(localhost:3307)/test_db?charset=utf8mb4&parseTime=True&loc=Local"
	}
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Skipf("Skipping: test database unavailable: %v", err)
	}
	if err := db.AutoMigrate(&entity.TDUser{}, &entity.TDUserStats{}, &entity.TDGameResult{}); err != nil {
		t.Skipf("Skipping: migration failed: %v", err)
	}
	return db
}

func createGameTestUser(t *testing.T, db *gorm.DB) *entity.TDUser {
	user := &entity.TDUser{
		Username:     "gameuser_" + time.Now().Format("20060102150405"),
		Email:        "game_" + time.Now().Format("20060102150405") + "@test.com",
		PasswordHash: "hash",
		IsActive:     true,
	}
	require.NoError(t, db.WithContext(context.Background()).Create(user).Error)
	return user
}

func TestTDGameRepository_Create(t *testing.T) {
	db := setupGameTestDB(t)
	requireTable(t, db, "td_users")
	requireTable(t, db, "td_game_results")
	ctx := context.Background()
	repo := NewTDGameRepository(db)
	user := createGameTestUser(t, db)

	result := &entity.TDGameResult{
		UserID:        user.ID,
		GameMode:      "single",
		RoundsReached: 10,
		MonstersKilled: 50,
		GoldEarned:    100,
		Result:        "victory",
	}
	err := repo.Create(ctx, result)
	require.NoError(t, err)
	require.NotZero(t, result.ID)
	t.Logf("Created game result ID: %d", result.ID)
}

func TestTDGameRepository_GetHistory(t *testing.T) {
	db := setupGameTestDB(t)
	requireTable(t, db, "td_users")
	requireTable(t, db, "td_game_results")
	ctx := context.Background()
	repo := NewTDGameRepository(db)
	user := createGameTestUser(t, db)

	for i := 0; i < 3; i++ {
		r := &entity.TDGameResult{
			UserID:        user.ID,
			GameMode:     "single",
			RoundsReached: uint(5 + i),
			MonstersKilled: 10,
			GoldEarned:   50,
			Result:       "victory",
		}
		require.NoError(t, repo.Create(ctx, r))
	}

	games, total, err := repo.GetHistory(ctx, user.ID, "single", 10, 0)
	require.NoError(t, err)
	require.GreaterOrEqual(t, len(games), 3)
	require.GreaterOrEqual(t, total, int64(3))
}

func TestTDGameRepository_GetHighestRound(t *testing.T) {
	db := setupGameTestDB(t)
	requireTable(t, db, "td_users")
	requireTable(t, db, "td_game_results")
	ctx := context.Background()
	repo := NewTDGameRepository(db)
	user := createGameTestUser(t, db)

	r := &entity.TDGameResult{
		UserID:        user.ID,
		GameMode:     "single",
		RoundsReached: 15,
		MonstersKilled: 100,
		GoldEarned:   200,
		Result:       "victory",
	}
	require.NoError(t, repo.Create(ctx, r))

	highest, err := repo.GetHighestRound(ctx, user.ID, "single")
	require.NoError(t, err)
	require.Equal(t, uint(15), highest)
}
