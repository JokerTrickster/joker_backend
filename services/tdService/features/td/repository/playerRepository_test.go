package repository

import (
	"testing"

	"github.com/JokerTrickster/joker_backend/services/tdService/features/td/model/entity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPlayerRepository_CreatePlayer(t *testing.T) {
	db := setupTestDB(t)
	repo := NewPlayerRepository(db)

	player := &entity.Player{
		UserID:     1001,
		Nickname:   "testplayer",
		AvatarID:   "avatar-1",
		Level:      1,
		Experience: 0,
	}

	err := repo.CreatePlayer(player)
	require.NoError(t, err)
	assert.NotZero(t, player.ID, "Player ID should be set after create")

	t.Logf("Created player with ID=%d, UserID=%d", player.ID, player.UserID)

	var count int64
	db.Model(&entity.Player{}).Where("user_id = ?", 1001).Count(&count)
	assert.Equal(t, int64(1), count, "Player should exist in database")
}

func TestPlayerRepository_GetPlayerByUserID_Found(t *testing.T) {
	db := setupTestDB(t)
	repo := NewPlayerRepository(db)

	player := &entity.Player{
		UserID:   2001,
		Nickname: "foundplayer",
		AvatarID: "avatar-2",
	}
	require.NoError(t, db.Create(player).Error)

	got, err := repo.GetPlayerByUserID(2001)
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, player.ID, got.ID)
	assert.Equal(t, uint(2001), got.UserID)
	assert.Equal(t, "foundplayer", got.Nickname)

	t.Logf("GetPlayerByUserID found player ID=%d", got.ID)
}

func TestPlayerRepository_GetPlayerByUserID_NotFound(t *testing.T) {
	db := setupTestDB(t)
	repo := NewPlayerRepository(db)

	got, err := repo.GetPlayerByUserID(99999)
	require.NoError(t, err)
	assert.Nil(t, got, "Expected nil when player not found")

	t.Log("GetPlayerByUserID correctly returns nil, nil for non-existent user")
}

func TestPlayerRepository_GetPlayerByID_Found(t *testing.T) {
	db := setupTestDB(t)
	repo := NewPlayerRepository(db)

	player := &entity.Player{
		UserID:   3001,
		Nickname: "idlookup",
		AvatarID: "avatar-3",
	}
	require.NoError(t, db.Create(player).Error)

	got, err := repo.GetPlayerByID(player.ID)
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, player.ID, got.ID)
	assert.Equal(t, uint(3001), got.UserID)

	t.Logf("GetPlayerByID found player by ID=%d", got.ID)
}

func TestPlayerRepository_GetPlayerByID_NotFound(t *testing.T) {
	db := setupTestDB(t)
	repo := NewPlayerRepository(db)

	got, err := repo.GetPlayerByID(99999)
	require.Error(t, err)
	assert.Nil(t, got)
	assert.Contains(t, err.Error(), "player not found")

	t.Log("GetPlayerByID correctly returns error for non-existent player ID")
}

func TestPlayerRepository_UpdatePlayer(t *testing.T) {
	db := setupTestDB(t)
	repo := NewPlayerRepository(db)

	player := &entity.Player{
		UserID:    4001,
		Nickname:  "original",
		AvatarID:  "avatar-orig",
		Level:     1,
		Experience: 0,
	}
	require.NoError(t, repo.CreatePlayer(player))

	player.Nickname = "updated_nick"
	player.Level = 5
	player.Experience = 100

	err := repo.UpdatePlayer(player)
	require.NoError(t, err)

	got, err := repo.GetPlayerByID(player.ID)
	require.NoError(t, err)
	assert.Equal(t, "updated_nick", got.Nickname)
	assert.Equal(t, 5, got.Level)
	assert.Equal(t, 100, got.Experience)

	t.Logf("UpdatePlayer verified: Nickname=%s, Level=%d", got.Nickname, got.Level)
}

func TestPlayerRepository_CreateOrUpdateStats(t *testing.T) {
	db := setupTestDB(t)
	repo := NewPlayerRepository(db)

	player := &entity.Player{
		UserID:   5001,
		Nickname: "statplayer",
		AvatarID: "avatar-5",
	}
	require.NoError(t, db.Create(player).Error)

	stats := &entity.PlayerStats{
		PlayerID:      player.ID,
		GamesPlayed:   5,
		Victories:     2,
		TotalScore:    1000,
		HighestScore:  500,
		HighestWave:   10,
	}

	err := repo.CreateOrUpdateStats(stats)
	require.NoError(t, err)
	assert.NotZero(t, stats.ID)

	var dbStats entity.PlayerStats
	require.NoError(t, db.Where("player_id = ?", player.ID).First(&dbStats).Error)
	assert.Equal(t, 5, dbStats.GamesPlayed)
	assert.Equal(t, 2, dbStats.Victories)
	assert.Equal(t, int64(1000), dbStats.TotalScore)

	t.Logf("CreateOrUpdateStats created stats ID=%d", stats.ID)
}

func TestPlayerRepository_UpdatePlayerStats(t *testing.T) {
	db := setupTestDB(t)
	repo := NewPlayerRepository(db)

	player := &entity.Player{
		UserID:   6001,
		Nickname: "statsupdate",
		AvatarID: "avatar-6",
	}
	require.NoError(t, db.Create(player).Error)

	stats := &entity.PlayerStats{
		PlayerID:   player.ID,
		GamesPlayed: 2,
		Victories:   1,
		TotalScore:  200,
		HighestScore: 150,
		HighestWave:  5,
		UnitsPlaced:  10,
		EnemiesKilled: 50,
		TotalPlayTime: 300,
	}
	require.NoError(t, db.Create(stats).Error)

	result := &entity.GameResult{
		SessionID:      "session-abc",
		PlayerID:       player.ID,
		Score:          300,
		WavesCompleted: 8,
		UnitsPlaced:    15,
		EnemiesDefeated: 80,
		Victory:        true,
		PlayTime:       120,
	}

	err := repo.UpdatePlayerStats(player.ID, result)
	require.NoError(t, err)

	var dbStats entity.PlayerStats
	require.NoError(t, db.Where("player_id = ?", player.ID).First(&dbStats).Error)
	assert.Equal(t, 3, dbStats.GamesPlayed, "GamesPlayed should increment")
	assert.Equal(t, 2, dbStats.Victories, "Victories should increment on victory")
	assert.Equal(t, int64(500), dbStats.TotalScore, "TotalScore should add 300")
	assert.Equal(t, int64(300), dbStats.HighestScore, "HighestScore should update from 300 > 150")
	assert.Equal(t, 8, dbStats.HighestWave, "HighestWave should update from 8 > 5")
	assert.Equal(t, 25, dbStats.UnitsPlaced, "UnitsPlaced should add 15")
	assert.Equal(t, 130, dbStats.EnemiesKilled, "EnemiesDefeated should add 80")
	assert.Equal(t, 420, dbStats.TotalPlayTime, "PlayTime should add 120")

	t.Log("UpdatePlayerStats verified all stat increments")
}

func TestPlayerRepository_GetOrCreatePlayer_New(t *testing.T) {
	db := setupTestDB(t)
	repo := NewPlayerRepository(db)

	player, err := repo.GetOrCreatePlayer(7001, "newplayer", "avatar-7")
	require.NoError(t, err)
	require.NotNil(t, player)
	assert.NotZero(t, player.ID)
	assert.Equal(t, uint(7001), player.UserID)
	assert.Equal(t, "newplayer", player.Nickname)
	assert.Equal(t, "avatar-7", player.AvatarID)
	assert.Equal(t, 1, player.Level)
	assert.Equal(t, 0, player.Experience)

	var stats entity.PlayerStats
	err = db.Where("player_id = ?", player.ID).First(&stats).Error
	require.NoError(t, err)
	assert.NotZero(t, stats.ID, "Initial stats should be created")

	t.Logf("GetOrCreatePlayer created new player ID=%d with initial stats", player.ID)
}

func TestPlayerRepository_GetOrCreatePlayer_Existing(t *testing.T) {
	db := setupTestDB(t)
	repo := NewPlayerRepository(db)

	existing := &entity.Player{
		UserID:   8001,
		Nickname: "existing",
		AvatarID: "avatar-8",
		Level:    3,
	}
	require.NoError(t, db.Create(existing).Error)

	got, err := repo.GetOrCreatePlayer(8001, "ignored_nick", "ignored_avatar")
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, existing.ID, got.ID)
	assert.Equal(t, "existing", got.Nickname, "Should return existing player, not overwrite with new nickname")
	assert.Equal(t, "avatar-8", got.AvatarID)
	assert.Equal(t, 3, got.Level)

	t.Log("GetOrCreatePlayer correctly returns existing player without creating duplicate")
}
