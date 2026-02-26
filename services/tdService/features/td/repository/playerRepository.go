package repository

import (
	"errors"
	"fmt"

	"github.com/JokerTrickster/joker_backend/services/tdService/features/td/model/entity"
	"gorm.io/gorm"
)

type PlayerRepository struct {
	db *gorm.DB
}

func NewPlayerRepository(db *gorm.DB) *PlayerRepository {
	return &PlayerRepository{db: db}
}

// CreatePlayer creates a new player
func (r *PlayerRepository) CreatePlayer(player *entity.Player) error {
	return r.db.Create(player).Error
}

// GetPlayerByUserID retrieves a player by user ID
func (r *PlayerRepository) GetPlayerByUserID(userID uint) (*entity.Player, error) {
	var player entity.Player
	err := r.db.Preload("Stats").Where("user_id = ?", userID).First(&player).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &player, nil
}

// GetPlayerByID retrieves a player by ID
func (r *PlayerRepository) GetPlayerByID(playerID uint) (*entity.Player, error) {
	var player entity.Player
	err := r.db.Preload("Stats").First(&player, playerID).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("player not found")
		}
		return nil, err
	}
	return &player, nil
}

// UpdatePlayer updates player information
func (r *PlayerRepository) UpdatePlayer(player *entity.Player) error {
	return r.db.Save(player).Error
}

// CreateOrUpdateStats creates or updates player statistics
func (r *PlayerRepository) CreateOrUpdateStats(stats *entity.PlayerStats) error {
	return r.db.Save(stats).Error
}

// UpdatePlayerStats updates player statistics after a game
func (r *PlayerRepository) UpdatePlayerStats(playerID uint, result *entity.GameResult) error {
	var stats entity.PlayerStats

	// Find or create stats
	err := r.db.Where("player_id = ?", playerID).First(&stats).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			stats = entity.PlayerStats{
				PlayerID: playerID,
			}
		} else {
			return err
		}
	}

	// Update stats
	stats.GamesPlayed++
	if result.Victory {
		stats.Victories++
	}
	stats.TotalScore += result.Score
	if result.Score > stats.HighestScore {
		stats.HighestScore = result.Score
	}
	if result.WavesCompleted > stats.HighestWave {
		stats.HighestWave = result.WavesCompleted
	}
	stats.UnitsPlaced += result.UnitsPlaced
	stats.EnemiesKilled += result.EnemiesDefeated
	stats.TotalPlayTime += result.PlayTime

	return r.db.Save(&stats).Error
}

// GetOrCreatePlayer gets existing player or creates a new one
func (r *PlayerRepository) GetOrCreatePlayer(userID uint, nickname string, avatarID string) (*entity.Player, error) {
	player, err := r.GetPlayerByUserID(userID)
	if err != nil {
		return nil, err
	}

	if player == nil {
		// Create new player
		player = &entity.Player{
			UserID:   userID,
			Nickname: nickname,
			AvatarID: avatarID,
			Level:    1,
			Experience: 0,
		}
		if err := r.CreatePlayer(player); err != nil {
			return nil, err
		}

		// Create initial stats
		stats := &entity.PlayerStats{
			PlayerID: player.ID,
		}
		if err := r.CreateOrUpdateStats(stats); err != nil {
			return nil, err
		}
		player.Stats = *stats
	}

	return player, nil
}