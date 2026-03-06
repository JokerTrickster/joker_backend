package repository

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"testing"
	"time"

	"github.com/JokerTrickster/joker_backend/services/lottoDefenseService/features/towerDefense/model/entity"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func setupRoomTestDB(t *testing.T) *gorm.DB {
	dsn := os.Getenv("TEST_DB_DSN")
	if dsn == "" {
		dsn = "root:rootpassword@tcp(localhost:3307)/test_db?charset=utf8mb4&parseTime=True&loc=Local"
	}
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Skipf("Skipping: test database unavailable: %v", err)
	}
	if err := db.AutoMigrate(&entity.TDUser{}, &entity.TDRoom{}, &entity.TDRoomPlayer{}); err != nil {
		t.Skipf("Skipping: migration failed: %v", err)
	}
	return db
}

func createRoomTestUser(t *testing.T, db *gorm.DB) *entity.TDUser {
	suffix := fmt.Sprintf("%d_%d", time.Now().UnixNano(), rand.Intn(100000))
	user := &entity.TDUser{
		Username:     "roomuser_" + suffix,
		Email:        "room_" + suffix + "@test.com",
		PasswordHash: "hash",
		IsActive:     true,
	}
	require.NoError(t, db.WithContext(context.Background()).Create(user).Error)
	return user
}

func TestTDRoomRepository_Create(t *testing.T) {
	db := setupRoomTestDB(t)
	requireTable(t, db, "td_users")
	requireTable(t, db, "td_rooms")
	ctx := context.Background()
	repo := NewTDRoomRepository(db)
	user := createRoomTestUser(t, db)

	room := &entity.TDRoom{
		RoomCode:       fmt.Sprintf("T%03d", rand.Intn(999)),
		HostUserID:     user.ID,
		RoomType:       "random",
		MaxPlayers:     2,
		CurrentPlayers: 1,
		Status:         "waiting",
		SharedGold:     100,
		Player1State:   "{}",
		Player2State:   "{}",
		ExpiresAt:      time.Now().Add(30 * time.Minute),
	}
	err := repo.Create(ctx, room)
	require.NoError(t, err)
	require.NotZero(t, room.ID)
}

func TestTDRoomRepository_GetByID(t *testing.T) {
	db := setupRoomTestDB(t)
	requireTable(t, db, "td_users")
	requireTable(t, db, "td_rooms")
	ctx := context.Background()
	repo := NewTDRoomRepository(db)
	user := createRoomTestUser(t, db)

	room := &entity.TDRoom{
		RoomCode:       fmt.Sprintf("G%03d", rand.Intn(999)),
		HostUserID:     user.ID,
		RoomType:       "random",
		MaxPlayers:     2,
		CurrentPlayers: 1,
		Status:         "waiting",
		Player1State:   "{}",
		Player2State:   "{}",
		ExpiresAt:      time.Now().Add(30 * time.Minute),
	}
	require.NoError(t, repo.Create(ctx, room))

	got, err := repo.GetByID(ctx, room.ID)
	require.NoError(t, err)
	require.Equal(t, room.ID, got.ID)
	require.Equal(t, room.RoomCode, got.RoomCode)
}

func TestTDRoomRepository_GetByCode(t *testing.T) {
	db := setupRoomTestDB(t)
	requireTable(t, db, "td_users")
	requireTable(t, db, "td_rooms")
	ctx := context.Background()
	repo := NewTDRoomRepository(db)
	user := createRoomTestUser(t, db)

	code := fmt.Sprintf("C%03d", rand.Intn(999))
	room := &entity.TDRoom{
		RoomCode:       code,
		HostUserID:     user.ID,
		RoomType:       "random",
		MaxPlayers:     2,
		CurrentPlayers: 1,
		Status:         "waiting",
		Player1State:   "{}",
		Player2State:   "{}",
		ExpiresAt:      time.Now().Add(30 * time.Minute),
	}
	require.NoError(t, repo.Create(ctx, room))

	got, err := repo.GetByCode(ctx, code)
	require.NoError(t, err)
	require.Equal(t, room.ID, got.ID)
}

func TestTDRoomRepository_AddPlayer(t *testing.T) {
	db := setupRoomTestDB(t)
	ctx := context.Background()
	repo := NewTDRoomRepository(db)
	user := createRoomTestUser(t, db)

	room := &entity.TDRoom{
		RoomCode:       fmt.Sprintf("A%03d", rand.Intn(999)),
		HostUserID:     user.ID,
		RoomType:       "random",
		MaxPlayers:     2,
		CurrentPlayers: 1,
		Status:         "waiting",
		Player1State:   "{}",
		Player2State:   "{}",
		ExpiresAt:      time.Now().Add(30 * time.Minute),
	}
	require.NoError(t, repo.Create(ctx, room))

	player2 := createRoomTestUser(t, db)
	player := &entity.TDRoomPlayer{
		RoomID:     room.ID,
		UserID:     player2.ID,
		PlayerSlot: 1,
		IsReady:    false,
	}
	err := repo.AddPlayer(ctx, player)
	require.NoError(t, err)
	require.NotZero(t, player.ID)
}

func TestTDRoomRepository_RemovePlayer(t *testing.T) {
	db := setupRoomTestDB(t)
	requireTable(t, db, "td_users")
	requireTable(t, db, "td_rooms")
	requireTable(t, db, "td_room_players")
	ctx := context.Background()
	repo := NewTDRoomRepository(db)
	user := createRoomTestUser(t, db)

	room := &entity.TDRoom{
		RoomCode:       fmt.Sprintf("R%03d", rand.Intn(999)),
		HostUserID:     user.ID,
		RoomType:       "random",
		MaxPlayers:     2,
		CurrentPlayers: 1,
		Status:         "waiting",
		Player1State:   "{}",
		Player2State:   "{}",
		ExpiresAt:      time.Now().Add(30 * time.Minute),
	}
	require.NoError(t, repo.Create(ctx, room))

	player := &entity.TDRoomPlayer{RoomID: room.ID, UserID: user.ID, PlayerSlot: 0, IsReady: false}
	require.NoError(t, repo.AddPlayer(ctx, player))

	err := repo.RemovePlayer(ctx, room.ID, user.ID)
	require.NoError(t, err)

	players, _ := repo.GetPlayers(ctx, room.ID)
	require.Empty(t, players)
}

func TestTDRoomRepository_GetPlayers(t *testing.T) {
	db := setupRoomTestDB(t)
	ctx := context.Background()
	repo := NewTDRoomRepository(db)
	user := createRoomTestUser(t, db)

	roomCode := fmt.Sprintf("GP%02d", rand.Intn(99))
	room := &entity.TDRoom{
		RoomCode:       roomCode,
		HostUserID:     user.ID,
		RoomType:       "random",
		MaxPlayers:     2,
		CurrentPlayers: 1,
		Status:         "waiting",
		Player1State:   "{}",
		Player2State:   "{}",
		ExpiresAt:      time.Now().Add(30 * time.Minute),
	}
	require.NoError(t, repo.Create(ctx, room))

	player := &entity.TDRoomPlayer{RoomID: room.ID, UserID: user.ID, PlayerSlot: 0, IsReady: false}
	require.NoError(t, repo.AddPlayer(ctx, player))

	players, err := repo.GetPlayers(ctx, room.ID)
	require.NoError(t, err)
	require.GreaterOrEqual(t, len(players), 1)
}

func TestTDRoomRepository_UpdatePlayerReady(t *testing.T) {
	db := setupRoomTestDB(t)
	ctx := context.Background()
	repo := NewTDRoomRepository(db)
	user := createRoomTestUser(t, db)

	roomCode := fmt.Sprintf("RD%02d", rand.Intn(99))
	room := &entity.TDRoom{
		RoomCode:       roomCode,
		HostUserID:     user.ID,
		RoomType:       "random",
		MaxPlayers:     2,
		CurrentPlayers: 1,
		Status:         "waiting",
		Player1State:   "{}",
		Player2State:   "{}",
		ExpiresAt:      time.Now().Add(30 * time.Minute),
	}
	require.NoError(t, repo.Create(ctx, room))

	player := &entity.TDRoomPlayer{RoomID: room.ID, UserID: user.ID, PlayerSlot: 0, IsReady: false}
	require.NoError(t, repo.AddPlayer(ctx, player))

	err := repo.UpdatePlayerReady(ctx, room.ID, user.ID, true)
	require.NoError(t, err)

	players, _ := repo.GetPlayers(ctx, room.ID)
	require.NotEmpty(t, players)
	require.True(t, players[0].IsReady)
}
