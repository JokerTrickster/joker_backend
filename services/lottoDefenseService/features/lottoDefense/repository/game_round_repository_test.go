package repository

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/JokerTrickster/joker_backend/services/lottoDefenseService/features/lottoDefense/model/entity"
	"github.com/JokerTrickster/joker_backend/shared/db/mysql"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func initTestDB(t *testing.T) *gorm.DB {
	if os.Getenv("MYSQL_HOST") == "" {
		t.Skip("MYSQL_HOST not set, skipping repository integration test")
	}
	if err := mysql.InitMySQL(); err != nil {
		t.Skipf("MySQL init failed (is DB running?): %v", err)
	}
	db := mysql.GormMysqlDB
	if err := db.AutoMigrate(&entity.GameRound{}, &entity.LottoDraw{}); err != nil {
		t.Fatalf("migrate game tables: %v", err)
	}
	// Ensure test users exist for FK (game_rounds.user_id -> users.id)
	for _, id := range []uint{1, 42, 99} {
		db.Exec("INSERT IGNORE INTO users (id, name, email, password, provider, created_at, updated_at) VALUES (?, 'test', ?, '', 'local', NOW(), NOW())", id, fmt.Sprintf("testuser%d@test.com", id))
	}
	return db
}

func TestGameRoundRepository_CreateAndGetByID(t *testing.T) {
	db := initTestDB(t)
	ctx := context.Background()
	repo := NewGameRoundRepository(db)

	now := time.Now()
	round := &entity.GameRound{
		UserID:    1,
		Status:    entity.RoundStatusActive,
		StartedAt: &now,
	}
	err := repo.Create(ctx, round)
	require.NoError(t, err)
	require.NotZero(t, round.ID, "Create should set ID")

	got, err := repo.GetByID(ctx, round.ID)
	require.NoError(t, err)
	require.Equal(t, round.ID, got.ID)
	require.Equal(t, uint(1), got.UserID)
	require.Equal(t, entity.RoundStatusActive, got.Status)
}

func TestGameRoundRepository_GetByIDAndUser(t *testing.T) {
	db := initTestDB(t)
	ctx := context.Background()
	repo := NewGameRoundRepository(db)

	now := time.Now()
	round := &entity.GameRound{UserID: 99, Status: entity.RoundStatusActive, StartedAt: &now}
	err := repo.Create(ctx, round)
	require.NoError(t, err)

	got, err := repo.GetByIDAndUser(ctx, round.ID, 99)
	require.NoError(t, err)
	require.Equal(t, round.ID, got.ID)

	_, err = repo.GetByIDAndUser(ctx, round.ID, 1)
	require.Error(t, err)
	require.Equal(t, gorm.ErrRecordNotFound, err)
}

func TestGameRoundRepository_Update(t *testing.T) {
	db := initTestDB(t)
	ctx := context.Background()
	repo := NewGameRoundRepository(db)

	now := time.Now()
	round := &entity.GameRound{UserID: 1, Status: entity.RoundStatusActive, StartedAt: &now}
	err := repo.Create(ctx, round)
	require.NoError(t, err)

	score := uint(1000)
	ended := now.Add(5 * time.Minute)
	round.Status = entity.RoundStatusCompleted
	round.Score = &score
	round.EndedAt = &ended
	err = repo.Update(ctx, round)
	require.NoError(t, err)

	got, err := repo.GetByID(ctx, round.ID)
	require.NoError(t, err)
	require.Equal(t, entity.RoundStatusCompleted, got.Status)
	require.NotNil(t, got.Score)
	require.Equal(t, uint(1000), *got.Score)
}

func TestGameRoundRepository_ListByUserID(t *testing.T) {
	db := initTestDB(t)
	ctx := context.Background()
	repo := NewGameRoundRepository(db)

	now := time.Now()
	for i := 0; i < 3; i++ {
		r := &entity.GameRound{UserID: 42, Status: entity.RoundStatusActive, StartedAt: &now}
		err := repo.Create(ctx, r)
		require.NoError(t, err)
	}

	rounds, err := repo.ListByUserID(ctx, 42, 10)
	require.NoError(t, err)
	require.GreaterOrEqual(t, len(rounds), 3)
}
