package usecase

import (
	"context"
	"errors"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/JokerTrickster/joker_backend/services/lottoDefenseService/features/lottoDefense/model/entity"
	_interface "github.com/JokerTrickster/joker_backend/services/lottoDefenseService/features/lottoDefense/model/interface"
	"github.com/JokerTrickster/joker_backend/services/lottoDefenseService/features/lottoDefense/model/request"
	"github.com/JokerTrickster/joker_backend/services/lottoDefenseService/features/lottoDefense/repository"
	"github.com/JokerTrickster/joker_backend/shared/db/mysql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

type mockRoundRepoForGameRound struct {
	mock.Mock
}

func (m *mockRoundRepoForGameRound) Create(ctx context.Context, round *entity.GameRound) error {
	args := m.Called(ctx, round)
	return args.Error(0)
}

func (m *mockRoundRepoForGameRound) GetByID(ctx context.Context, id uint) (*entity.GameRound, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.GameRound), args.Error(1)
}

func (m *mockRoundRepoForGameRound) GetByIDAndUser(ctx context.Context, id, userID uint) (*entity.GameRound, error) {
	args := m.Called(ctx, id, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.GameRound), args.Error(1)
}

func (m *mockRoundRepoForGameRound) Update(ctx context.Context, round *entity.GameRound) error {
	args := m.Called(ctx, round)
	return args.Error(0)
}

func (m *mockRoundRepoForGameRound) ListByUserID(ctx context.Context, userID uint, limit int) ([]entity.GameRound, error) {
	args := m.Called(ctx, userID, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]entity.GameRound), args.Error(1)
}

func (m *mockRoundRepoForGameRound) TopScores(ctx context.Context, limit int) ([]entity.GameRound, error) {
	args := m.Called(ctx, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]entity.GameRound), args.Error(1)
}

func (m *mockRoundRepoForGameRound) Leaderboard(ctx context.Context, limit int) ([]_interface.LeaderboardRow, error) {
	args := m.Called(ctx, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]_interface.LeaderboardRow), args.Error(1)
}

type mockLottoDrawRepository struct {
	mock.Mock
}

func (m *mockLottoDrawRepository) Create(ctx context.Context, draw *entity.LottoDraw) error {
	args := m.Called(ctx, draw)
	return args.Error(0)
}

func (m *mockLottoDrawRepository) GetByRoundID(ctx context.Context, roundID uint) (*entity.LottoDraw, error) {
	args := m.Called(ctx, roundID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.LottoDraw), args.Error(1)
}

func TestGameRoundUseCase_StartRound_Unit(t *testing.T) {
	t.Log("StartRound: success with mocks")
	mockRoundRepo := new(mockRoundRepoForGameRound)
	mockDrawRepo := new(mockLottoDrawRepository)
	uc := NewGameRoundUseCase(mockRoundRepo, mockDrawRepo, 10*time.Second)
	ctx := context.Background()
	userID := uint(1)

	mockRoundRepo.On("Create", ctx, mock.AnythingOfType("*entity.GameRound")).Run(func(args mock.Arguments) {
		r := args.Get(1).(*entity.GameRound)
		r.ID = 1
		r.CreatedAt = time.Now()
	}).Return(nil)

	resp, err := uc.StartRound(ctx, userID)
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, uint(1), resp.ID)
	assert.Equal(t, userID, resp.UserID)
	assert.Equal(t, "active", resp.Status)
	assert.NotNil(t, resp.StartedAt)
	mockRoundRepo.AssertExpectations(t)
}

func TestGameRoundUseCase_StartRound_CreateError(t *testing.T) {
	t.Log("StartRound: create error")
	mockRoundRepo := new(mockRoundRepoForGameRound)
	mockDrawRepo := new(mockLottoDrawRepository)
	uc := NewGameRoundUseCase(mockRoundRepo, mockDrawRepo, 10*time.Second)
	ctx := context.Background()

	mockRoundRepo.On("Create", ctx, mock.AnythingOfType("*entity.GameRound")).Return(errors.New("db error"))

	resp, err := uc.StartRound(ctx, 1)
	assert.Error(t, err)
	assert.Nil(t, resp)
	mockRoundRepo.AssertExpectations(t)
}

func TestGameRoundUseCase_EndRound_Unit(t *testing.T) {
	t.Log("EndRound: success with mocks")
	mockRoundRepo := new(mockRoundRepoForGameRound)
	mockDrawRepo := new(mockLottoDrawRepository)
	uc := NewGameRoundUseCase(mockRoundRepo, mockDrawRepo, 10*time.Second)
	ctx := context.Background()
	userID := uint(1)
	roundID := uint(1)
	req := &request.EndRoundRequest{Score: 500}

	now := time.Now()
	round := &entity.GameRound{ID: roundID, UserID: userID, Status: entity.RoundStatusActive, StartedAt: &now, CreatedAt: now}
	mockRoundRepo.On("GetByIDAndUser", ctx, roundID, userID).Return(round, nil)
	mockRoundRepo.On("Update", ctx, mock.AnythingOfType("*entity.GameRound")).Return(nil)
	mockDrawRepo.On("Create", ctx, mock.AnythingOfType("*entity.LottoDraw")).Return(nil)

	resp, err := uc.EndRound(ctx, userID, roundID, req)
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, "completed", resp.Status)
	require.NotNil(t, resp.Score)
	assert.Equal(t, uint(500), *resp.Score)
	assert.Len(t, resp.Numbers, 6)
	mockRoundRepo.AssertExpectations(t)
	mockDrawRepo.AssertExpectations(t)
}

func TestGameRoundUseCase_EndRound_NotFound_Unit(t *testing.T) {
	t.Log("EndRound: round not found (unit)")
	mockRoundRepo := new(mockRoundRepoForGameRound)
	mockDrawRepo := new(mockLottoDrawRepository)
	uc := NewGameRoundUseCase(mockRoundRepo, mockDrawRepo, 10*time.Second)
	ctx := context.Background()

	mockRoundRepo.On("GetByIDAndUser", ctx, uint(999), uint(1)).Return(nil, gorm.ErrRecordNotFound)

	resp, err := uc.EndRound(ctx, 1, 999, &request.EndRoundRequest{Score: 100})
	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.ErrorIs(t, err, ErrRoundNotFound)
	mockRoundRepo.AssertExpectations(t)
	mockDrawRepo.AssertNotCalled(t, "Create")
}

func TestGameRoundUseCase_EndRound_AlreadyCompleted(t *testing.T) {
	t.Log("EndRound: round already completed")
	mockRoundRepo := new(mockRoundRepoForGameRound)
	mockDrawRepo := new(mockLottoDrawRepository)
	uc := NewGameRoundUseCase(mockRoundRepo, mockDrawRepo, 10*time.Second)
	ctx := context.Background()
	userID := uint(1)
	roundID := uint(1)

	now := time.Now()
	round := &entity.GameRound{ID: roundID, UserID: userID, Status: entity.RoundStatusCompleted, StartedAt: &now, EndedAt: &now, CreatedAt: now}
	mockRoundRepo.On("GetByIDAndUser", ctx, roundID, userID).Return(round, nil)

	resp, err := uc.EndRound(ctx, userID, roundID, &request.EndRoundRequest{Score: 500})
	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.ErrorIs(t, err, ErrRoundCompleted)
	mockRoundRepo.AssertExpectations(t)
	mockRoundRepo.AssertNotCalled(t, "Update")
	mockDrawRepo.AssertNotCalled(t, "Create")
}

func TestGameRoundUseCase_GetMyRounds_Unit(t *testing.T) {
	t.Log("GetMyRounds: success with mocks")
	mockRoundRepo := new(mockRoundRepoForGameRound)
	mockDrawRepo := new(mockLottoDrawRepository)
	uc := NewGameRoundUseCase(mockRoundRepo, mockDrawRepo, 10*time.Second)
	ctx := context.Background()
	userID := uint(1)
	limit := 10

	now := time.Now()
	rounds := []entity.GameRound{
		{ID: 1, UserID: userID, Status: entity.RoundStatusActive, StartedAt: &now, CreatedAt: now},
	}
	mockRoundRepo.On("ListByUserID", ctx, userID, limit).Return(rounds, nil)

	resp, err := uc.GetMyRounds(ctx, userID, limit)
	require.NoError(t, err)
	require.Len(t, resp, 1)
	assert.Equal(t, uint(1), resp[0].ID)
	assert.Equal(t, "active", resp[0].Status)
	mockRoundRepo.AssertExpectations(t)
}

func TestGameRoundUseCase_GetMyRounds_Empty(t *testing.T) {
	t.Log("GetMyRounds: empty result")
	mockRoundRepo := new(mockRoundRepoForGameRound)
	mockDrawRepo := new(mockLottoDrawRepository)
	uc := NewGameRoundUseCase(mockRoundRepo, mockDrawRepo, 10*time.Second)
	ctx := context.Background()

	mockRoundRepo.On("ListByUserID", ctx, uint(1), 20).Return([]entity.GameRound{}, nil)

	resp, err := uc.GetMyRounds(ctx, 1, 20)
	require.NoError(t, err)
	require.Empty(t, resp)
	mockRoundRepo.AssertExpectations(t)
}

func TestGameRoundUseCase_GetRound_Unit(t *testing.T) {
	t.Log("GetRound: success - active round without draw")
	mockRoundRepo := new(mockRoundRepoForGameRound)
	mockDrawRepo := new(mockLottoDrawRepository)
	uc := NewGameRoundUseCase(mockRoundRepo, mockDrawRepo, 10*time.Second)
	ctx := context.Background()
	userID := uint(1)
	roundID := uint(1)

	now := time.Now()
	round := &entity.GameRound{ID: roundID, UserID: userID, Status: entity.RoundStatusActive, StartedAt: &now, CreatedAt: now}
	mockRoundRepo.On("GetByIDAndUser", ctx, roundID, userID).Return(round, nil)

	resp, err := uc.GetRound(ctx, userID, roundID)
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, uint(1), resp.ID)
	assert.Nil(t, resp.Numbers)
	mockRoundRepo.AssertExpectations(t)
}

func TestGameRoundUseCase_GetRound_CompletedWithDraw(t *testing.T) {
	t.Log("GetRound: completed round includes draw numbers")
	mockRoundRepo := new(mockRoundRepoForGameRound)
	mockDrawRepo := new(mockLottoDrawRepository)
	uc := NewGameRoundUseCase(mockRoundRepo, mockDrawRepo, 10*time.Second)
	ctx := context.Background()
	userID := uint(1)
	roundID := uint(1)

	now := time.Now()
	round := &entity.GameRound{ID: roundID, UserID: userID, Status: entity.RoundStatusCompleted, StartedAt: &now, EndedAt: &now, CreatedAt: now}
	draw := &entity.LottoDraw{RoundID: roundID, Numbers: entity.LottoNumbers{1, 2, 3, 4, 5, 6}}
	mockRoundRepo.On("GetByIDAndUser", ctx, roundID, userID).Return(round, nil)
	mockDrawRepo.On("GetByRoundID", ctx, roundID).Return(draw, nil)

	resp, err := uc.GetRound(ctx, userID, roundID)
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, uint(1), resp.ID)
	require.Len(t, resp.Numbers, 6)
	assert.Equal(t, []int{1, 2, 3, 4, 5, 6}, resp.Numbers)
	mockRoundRepo.AssertExpectations(t)
	mockDrawRepo.AssertExpectations(t)
}

func TestGameRoundUseCase_GetRound_NotFound(t *testing.T) {
	t.Log("GetRound: round not found")
	mockRoundRepo := new(mockRoundRepoForGameRound)
	mockDrawRepo := new(mockLottoDrawRepository)
	uc := NewGameRoundUseCase(mockRoundRepo, mockDrawRepo, 10*time.Second)
	ctx := context.Background()

	mockRoundRepo.On("GetByIDAndUser", ctx, uint(999), uint(1)).Return(nil, gorm.ErrRecordNotFound)

	resp, err := uc.GetRound(ctx, 1, 999)
	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.ErrorIs(t, err, ErrRoundNotFound)
	mockRoundRepo.AssertExpectations(t)
}

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
