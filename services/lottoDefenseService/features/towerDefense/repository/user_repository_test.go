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

func setupTestDB(t *testing.T) *gorm.DB {
	dsn := os.Getenv("TEST_DB_DSN")
	if dsn == "" {
		dsn = "root:rootpassword@tcp(localhost:3307)/test_db?charset=utf8mb4&parseTime=True&loc=Local"
	}
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Skipf("Skipping: test database unavailable: %v", err)
	}
	if err := db.AutoMigrate(&entity.TDUser{}, &entity.TDUserStats{}); err != nil {
		t.Skip("Skipping: migration failed:", err)
	}
	return db
}

func TestTDUserRepository_Create(t *testing.T) {
	db := setupTestDB(t)
	ctx := context.Background()
	repo := NewTDUserRepository(db)

	user := &entity.TDUser{
		Username:     "testuser_" + time.Now().Format("20060102150405"),
		Email:        "test_" + time.Now().Format("20060102150405") + "@test.com",
		PasswordHash: "hashed",
		IsActive:     true,
	}
	err := repo.Create(ctx, user)
	require.NoError(t, err)
	require.NotZero(t, user.ID)
	t.Logf("Created user ID: %d", user.ID)
}

func TestTDUserRepository_GetByID(t *testing.T) {
	db := setupTestDB(t)
	ctx := context.Background()
	repo := NewTDUserRepository(db)

	user := &entity.TDUser{
		Username:     "getbyid_" + time.Now().Format("20060102150405"),
		Email:        "getbyid_" + time.Now().Format("20060102150405") + "@test.com",
		PasswordHash: "hash",
		IsActive:     true,
	}
	require.NoError(t, repo.Create(ctx, user))

	got, err := repo.GetByID(ctx, user.ID)
	require.NoError(t, err)
	require.Equal(t, user.ID, got.ID)
	require.Equal(t, user.Username, got.Username)
}

func TestTDUserRepository_GetByEmail(t *testing.T) {
	db := setupTestDB(t)
	ctx := context.Background()
	repo := NewTDUserRepository(db)

	email := "getbyemail_" + time.Now().Format("20060102150405") + "@test.com"
	user := &entity.TDUser{
		Username:     "getbyemail_" + time.Now().Format("20060102150405"),
		Email:        email,
		PasswordHash: "hash",
		IsActive:     true,
	}
	require.NoError(t, repo.Create(ctx, user))

	got, err := repo.GetByEmail(ctx, email)
	require.NoError(t, err)
	require.Equal(t, user.ID, got.ID)
	require.Equal(t, email, got.Email)
}

func TestTDUserRepository_GetByUsername(t *testing.T) {
	db := setupTestDB(t)
	ctx := context.Background()
	repo := NewTDUserRepository(db)

	username := "getbyuname_" + time.Now().Format("20060102150405")
	user := &entity.TDUser{
		Username:     username,
		Email:        "uname_" + time.Now().Format("20060102150405") + "@test.com",
		PasswordHash: "hash",
		IsActive:     true,
	}
	require.NoError(t, repo.Create(ctx, user))

	got, err := repo.GetByUsername(ctx, username)
	require.NoError(t, err)
	require.Equal(t, user.ID, got.ID)
	require.Equal(t, username, got.Username)
}

func TestTDUserRepository_GetStats(t *testing.T) {
	db := setupTestDB(t)
	ctx := context.Background()
	repo := NewTDUserRepository(db)

	user := &entity.TDUser{
		Username:     "stats_" + time.Now().Format("20060102150405"),
		Email:        "stats_" + time.Now().Format("20060102150405") + "@test.com",
		PasswordHash: "hash",
		IsActive:     true,
	}
	require.NoError(t, repo.Create(ctx, user))

	stats := &entity.TDUserStats{UserID: user.ID}
	err := repo.CreateStats(ctx, stats)
	require.NoError(t, err)

	got, err := repo.GetStats(ctx, user.ID)
	require.NoError(t, err)
	require.Equal(t, user.ID, got.UserID)
}

func TestTDUserRepository_CreateStats(t *testing.T) {
	db := setupTestDB(t)
	ctx := context.Background()
	repo := NewTDUserRepository(db)

	user := &entity.TDUser{
		Username:     "createstats_" + time.Now().Format("20060102150405"),
		Email:        "createstats_" + time.Now().Format("20060102150405") + "@test.com",
		PasswordHash: "hash",
		IsActive:     true,
	}
	require.NoError(t, repo.Create(ctx, user))

	stats := &entity.TDUserStats{UserID: user.ID, SingleHighestRound: 5, CurrentGold: 100}
	err := repo.CreateStats(ctx, stats)
	require.NoError(t, err)

	got, err := repo.GetStats(ctx, user.ID)
	require.NoError(t, err)
	require.Equal(t, uint(5), got.SingleHighestRound)
	require.Equal(t, uint(100), got.CurrentGold)
}
