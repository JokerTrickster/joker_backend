package usecase

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/JokerTrickster/joker_backend/services/lottoDefenseService/features/lottoDefense/model/entity"
	_interface "github.com/JokerTrickster/joker_backend/services/lottoDefenseService/features/lottoDefense/model/interface"
	"github.com/JokerTrickster/joker_backend/services/lottoDefenseService/features/lottoDefense/model/request"
	"github.com/JokerTrickster/joker_backend/services/lottoDefenseService/features/lottoDefense/repository"
	"github.com/JokerTrickster/joker_backend/shared/db/mysql"
	"github.com/stretchr/testify/require"
)

func initTestDB(t *testing.T) (_interface.IGameRoundRepository, _interface.ILottoDrawRepository) {
	if os.Getenv("MYSQL_HOST") == "" {
		t.Skip("MYSQL_HOST not set, skipping usecase integration test")
	}
	if err := mysql.InitMySQL(); err != nil {
		t.Skipf("MySQL init failed: %v", err)
	}
	db := mysql.GormMysqlDB
	if err := db.AutoMigrate(&entity.GameRound{}, &entity.LottoDraw{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	for _, id := range []uint{1, 2} {
		db.Exec("INSERT IGNORE INTO users (id, name, email, password, provider, created_at, updated_at) VALUES (?, 'test', ?, '', 'local', NOW(), NOW())", id, fmt.Sprintf("uc_test_%d@test.com", id))
	}
	return repository.NewGameRoundRepository(db), repository.NewLottoDrawRepository(db)
}

func TestGameRoundUseCase_StartRound(t *testing.T) {
	roundRepo, drawRepo := initTestDB(t)
	uc := NewGameRoundUseCase(roundRepo, drawRepo, 10*time.Second)
	ctx := context.Background()

	resp, err := uc.StartRound(ctx, 1)
	require.NoError(t, err)
	require.NotZero(t, resp.ID)
	require.Equal(t, "active", resp.Status)
	require.Equal(t, uint(1), resp.UserID)
	require.NotNil(t, resp.StartedAt)
}

func TestGameRoundUseCase_EndRound(t *testing.T) {
	roundRepo, drawRepo := initTestDB(t)
	uc := NewGameRoundUseCase(roundRepo, drawRepo, 10*time.Second)
	ctx := context.Background()

	start, err := uc.StartRound(ctx, 1)
	require.NoError(t, err)

	endReq := &request.EndRoundRequest{Score: 500}
	resp, err := uc.EndRound(ctx, 1, start.ID, endReq)
	require.NoError(t, err)
	require.Equal(t, "completed", resp.Status)
	require.NotNil(t, resp.Score)
	require.Equal(t, uint(500), *resp.Score)
	require.Len(t, resp.Numbers, 6, "completed round must include lotto draw")
}

func TestGameRoundUseCase_EndRound_NotFound(t *testing.T) {
	roundRepo, drawRepo := initTestDB(t)
	uc := NewGameRoundUseCase(roundRepo, drawRepo, 10*time.Second)
	ctx := context.Background()

	_, err := uc.EndRound(ctx, 1, 99999, &request.EndRoundRequest{Score: 100})
	require.Error(t, err)
	require.Contains(t, err.Error(), "not found")
}

func TestGameRoundUseCase_GetMyRounds(t *testing.T) {
	roundRepo, drawRepo := initTestDB(t)
	uc := NewGameRoundUseCase(roundRepo, drawRepo, 10*time.Second)
	ctx := context.Background()

	_, err := uc.StartRound(ctx, 2)
	require.NoError(t, err)

	rounds, err := uc.GetMyRounds(ctx, 2, 10)
	require.NoError(t, err)
	require.GreaterOrEqual(t, len(rounds), 1)
}
