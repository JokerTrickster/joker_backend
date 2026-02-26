package entity

import (
	"database/sql/driver"
	"time"
)

// GameMode represents the game mode
type GameMode string

const (
	GameModeSingle GameMode = "single"
	GameModeCoop   GameMode = "coop"
)

// GameStatus represents the game session status
type GameStatus string

const (
	GameStatusWaiting  GameStatus = "waiting"
	GameStatusPlaying  GameStatus = "playing"
	GameStatusFinished GameStatus = "finished"
	GameStatusAborted  GameStatus = "aborted"
)

// Scan implements the Scanner interface
func (g *GameMode) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	if v, ok := value.([]byte); ok {
		*g = GameMode(v)
	} else if v, ok := value.(string); ok {
		*g = GameMode(v)
	}
	return nil
}

// Value implements the driver Valuer interface
func (g GameMode) Value() (driver.Value, error) {
	return string(g), nil
}

// Scan implements the Scanner interface
func (s *GameStatus) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	if v, ok := value.([]byte); ok {
		*s = GameStatus(v)
	} else if v, ok := value.(string); ok {
		*s = GameStatus(v)
	}
	return nil
}

// Value implements the driver Valuer interface
func (s GameStatus) Value() (driver.Value, error) {
	return string(s), nil
}

// GameSession represents a game session
type GameSession struct {
	ID           uint       `gorm:"primaryKey" json:"id"`
	SessionID    string     `gorm:"uniqueIndex;size:100;not null" json:"sessionId"`
	RoomCode     string     `gorm:"index;size:20" json:"roomCode,omitempty"`
	Mode         GameMode   `gorm:"type:varchar(20);not null" json:"mode"`
	Status       GameStatus `gorm:"type:varchar(20);not null;default:'waiting'" json:"status"`
	PlayerCount  int        `gorm:"not null" json:"playerCount"`
	PlayerID     uint       `gorm:"not null;index" json:"playerId"` // Host player
	Player2ID    *uint      `gorm:"index" json:"player2Id,omitempty"` // For coop mode
	WebSocketURL string     `gorm:"size:255" json:"webSocketUrl,omitempty"`
	StartedAt    *time.Time `json:"startedAt,omitempty"`
	EndedAt      *time.Time `json:"endedAt,omitempty"`
	CreatedAt    time.Time  `json:"createdAt"`
	UpdatedAt    time.Time  `json:"updatedAt"`

	// Relations
	Player  *Player      `gorm:"foreignKey:PlayerID" json:"player,omitempty"`
	Player2 *Player      `gorm:"foreignKey:Player2ID" json:"player2,omitempty"`
	Results []GameResult `gorm:"foreignKey:SessionID;references:SessionID" json:"results,omitempty"`
}

// GameResult represents game results for a session
type GameResult struct {
	ID              uint       `gorm:"primaryKey" json:"id"`
	SessionID       string     `gorm:"index;not null" json:"sessionId"`
	PlayerID        uint       `gorm:"index;not null" json:"playerId"`
	Score           int64      `gorm:"not null" json:"score"`
	WavesCompleted  int        `gorm:"not null" json:"wavesCompleted"`
	UnitsPlaced     int        `gorm:"not null" json:"unitsPlaced"`
	EnemiesDefeated int        `gorm:"not null" json:"enemiesDefeated"`
	Victory         bool       `gorm:"not null" json:"victory"`
	PlayTime        int        `gorm:"not null" json:"playTime"` // in seconds
	CreatedAt       time.Time  `json:"createdAt"`

	// Relations
	Session *GameSession `gorm:"foreignKey:SessionID;references:SessionID" json:"session,omitempty"`
	Player  *Player      `gorm:"foreignKey:PlayerID" json:"player,omitempty"`
}