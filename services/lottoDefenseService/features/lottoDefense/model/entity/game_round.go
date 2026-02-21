package entity

import "time"

// RoundStatus represents the status of a game round
type RoundStatus string

const (
	RoundStatusActive    RoundStatus = "active"
	RoundStatusCompleted RoundStatus = "completed"
)

// GameRound represents a single game round for a user
type GameRound struct {
	ID        uint        `gorm:"primaryKey" json:"id"`
	UserID    uint        `gorm:"not null;index" json:"user_id"`
	Status    RoundStatus `gorm:"size:20;not null;default:active;index" json:"status"`
	Score     *uint       `gorm:"index" json:"score,omitempty"`
	StartedAt *time.Time  `json:"started_at,omitempty"`
	EndedAt   *time.Time  `json:"ended_at,omitempty"`
	CreatedAt time.Time   `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time   `gorm:"autoUpdateTime" json:"updated_at"`
}

// TableName specifies the table name for GameRound
func (GameRound) TableName() string {
	return "game_rounds"
}
