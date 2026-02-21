package repository

import (
	"context"
	"os"
	"testing"

	"github.com/JokerTrickster/joker_backend/services/lottoDefenseService/features/lottoDefense/model/entity"
	"github.com/JokerTrickster/joker_backend/shared/db/mysql"
	"github.com/stretchr/testify/require"
)

func TestLottoDrawRepository_CreateAndGetByRoundID(t *testing.T) {
	if os.Getenv("MYSQL_HOST") == "" {
		t.Skip("MYSQL_HOST not set, skipping integration test")
	}
	if err := mysql.InitMySQL(); err != nil {
		t.Skipf("MySQL init failed: %v", err)
	}
	db := mysql.GormMysqlDB
	if err := db.AutoMigrate(&entity.GameRound{}, &entity.LottoDraw{}); err != nil {
		t.Fatalf("migrate game tables: %v", err)
	}
	db.Exec("INSERT IGNORE INTO users (id, name, email, password, provider, created_at, updated_at) VALUES (1, 'test', ?, '', 'local', NOW(), NOW())", "testuser1@test.com")
	ctx := context.Background()

	roundRepo := NewGameRoundRepository(db)
	drawRepo := NewLottoDrawRepository(db)

	round := &entity.GameRound{UserID: 1, Status: entity.RoundStatusActive}
	err := roundRepo.Create(ctx, round)
	require.NoError(t, err)

	draw := &entity.LottoDraw{RoundID: round.ID, Numbers: entity.LottoNumbers([6]int{1, 2, 3, 4, 5, 6})}
	err = drawRepo.Create(ctx, draw)
	require.NoError(t, err)
	require.NotZero(t, draw.ID)

	got, err := drawRepo.GetByRoundID(ctx, round.ID)
	require.NoError(t, err)
	require.Equal(t, draw.RoundID, got.RoundID)
	require.Equal(t, [6]int{1, 2, 3, 4, 5, 6}, got.Numbers)
}
