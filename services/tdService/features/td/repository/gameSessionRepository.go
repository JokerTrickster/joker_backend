package repository

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/JokerTrickster/joker_backend/services/tdService/features/td/model/entity"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type GameSessionRepository struct {
	db *gorm.DB
}

func NewGameSessionRepository(db *gorm.DB) *GameSessionRepository {
	return &GameSessionRepository{db: db}
}

// CreateSession creates a new game session
func (r *GameSessionRepository) CreateSession(session *entity.GameSession) error {
	// Generate session ID
	session.SessionID = uuid.New().String()

	// Generate room code for coop mode
	if session.Mode == entity.GameModeCoop && session.RoomCode == "" {
		roomCode, err := r.generateUniqueRoomCode()
		if err != nil {
			return err
		}
		session.RoomCode = roomCode
	}

	// Set initial status
	session.Status = entity.GameStatusWaiting

	return r.db.Create(session).Error
}

// GetSessionByID retrieves a session by ID
func (r *GameSessionRepository) GetSessionByID(sessionID string) (*entity.GameSession, error) {
	var session entity.GameSession
	err := r.db.Preload("Player").Preload("Player2").Where("session_id = ?", sessionID).First(&session).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("session not found")
		}
		return nil, err
	}
	return &session, nil
}

// GetSessionByRoomCode retrieves a session by room code
func (r *GameSessionRepository) GetSessionByRoomCode(roomCode string) (*entity.GameSession, error) {
	var session entity.GameSession
	err := r.db.Preload("Player").Preload("Player2").
		Where("room_code = ? AND status = ?", roomCode, entity.GameStatusWaiting).
		First(&session).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("session not found or already started")
		}
		return nil, err
	}
	return &session, nil
}

// UpdateSession updates a game session
func (r *GameSessionRepository) UpdateSession(session *entity.GameSession) error {
	return r.db.Save(session).Error
}

// JoinSession adds a second player to a coop session
func (r *GameSessionRepository) JoinSession(sessionID string, playerID uint) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		var session entity.GameSession
		err := tx.Where("session_id = ?", sessionID).First(&session).Error
		if err != nil {
			return err
		}

		if session.Mode != entity.GameModeCoop {
			return fmt.Errorf("session is not in coop mode")
		}

		if session.Player2ID != nil {
			return fmt.Errorf("session is already full")
		}

		if session.Status != entity.GameStatusWaiting {
			return fmt.Errorf("session already started")
		}

		session.Player2ID = &playerID
		return tx.Save(&session).Error
	})
}

// StartSession starts a game session
func (r *GameSessionRepository) StartSession(sessionID string) error {
	now := time.Now()
	return r.db.Model(&entity.GameSession{}).
		Where("session_id = ?", sessionID).
		Updates(map[string]interface{}{
			"status":     entity.GameStatusPlaying,
			"started_at": now,
		}).Error
}

// EndSession ends a game session
func (r *GameSessionRepository) EndSession(sessionID string) error {
	now := time.Now()
	return r.db.Model(&entity.GameSession{}).
		Where("session_id = ?", sessionID).
		Updates(map[string]interface{}{
			"status":   entity.GameStatusFinished,
			"ended_at": now,
		}).Error
}

// GetActiveSessions retrieves active sessions for a player
func (r *GameSessionRepository) GetActiveSessions(playerID uint) ([]*entity.GameSession, error) {
	var sessions []*entity.GameSession
	err := r.db.Where("(player_id = ? OR player2_id = ?) AND status IN ?",
		playerID, playerID, []entity.GameStatus{entity.GameStatusWaiting, entity.GameStatusPlaying}).
		Find(&sessions).Error
	return sessions, err
}

// SaveGameResult saves game result for a session
func (r *GameSessionRepository) SaveGameResult(result *entity.GameResult) error {
	return r.db.Create(result).Error
}

// GetGameResults retrieves game results for a session
func (r *GameSessionRepository) GetGameResults(sessionID string) ([]*entity.GameResult, error) {
	var results []*entity.GameResult
	err := r.db.Preload("Player").Where("session_id = ?", sessionID).Find(&results).Error
	return results, err
}

// GetPlayerGameHistory retrieves game history for a player
func (r *GameSessionRepository) GetPlayerGameHistory(playerID uint, limit int) ([]*entity.GameResult, error) {
	var results []*entity.GameResult
	query := r.db.Preload("Session").Where("player_id = ?", playerID).Order("created_at DESC")
	if limit > 0 {
		query = query.Limit(limit)
	}
	err := query.Find(&results).Error
	return results, err
}

// generateUniqueRoomCode generates a unique room code
func (r *GameSessionRepository) generateUniqueRoomCode() (string, error) {
	for i := 0; i < 10; i++ { // Try up to 10 times
		code := r.generateRoomCode()

		// Check if code already exists
		var count int64
		err := r.db.Model(&entity.GameSession{}).
			Where("room_code = ? AND status = ?", code, entity.GameStatusWaiting).
			Count(&count).Error
		if err != nil {
			return "", err
		}

		if count == 0 {
			return code, nil
		}
	}
	return "", fmt.Errorf("failed to generate unique room code")
}

// generateRoomCode generates a random room code
func (r *GameSessionRepository) generateRoomCode() string {
	bytes := make([]byte, 3)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)[:6] // 6 character code
}