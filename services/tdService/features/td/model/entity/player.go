package entity

import (
	"time"
)

// Player represents a game player
type Player struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	UserID     uint      `gorm:"not null;index" json:"userId"`
	Nickname   string    `gorm:"size:50;not null" json:"nickname"`
	AvatarID   string    `gorm:"size:100" json:"avatarId"`
	Level      int       `gorm:"default:1" json:"level"`
	Experience int       `gorm:"default:0" json:"experience"`
	CreatedAt  time.Time `json:"createdAt"`
	UpdatedAt  time.Time `json:"updatedAt"`

	// Stats
	Stats PlayerStats `gorm:"foreignKey:PlayerID" json:"stats,omitempty"`

	// Relations
	GameSessions []GameSession `gorm:"foreignKey:PlayerID" json:"-"`
}

// PlayerStats represents player statistics
type PlayerStats struct {
	ID             uint      `gorm:"primaryKey" json:"id"`
	PlayerID       uint      `gorm:"uniqueIndex;not null" json:"playerId"`
	GamesPlayed    int       `gorm:"default:0" json:"gamesPlayed"`
	Victories      int       `gorm:"default:0" json:"victories"`
	TotalScore     int64     `gorm:"default:0" json:"totalScore"`
	HighestScore   int64     `gorm:"default:0" json:"highestScore"`
	HighestWave    int       `gorm:"default:0" json:"highestWave"`
	TotalPlayTime  int       `gorm:"default:0" json:"totalPlayTime"` // in seconds
	UnitsPlaced    int       `gorm:"default:0" json:"unitsPlaced"`
	EnemiesKilled  int       `gorm:"default:0" json:"enemiesDefeated"`
	CreatedAt      time.Time `json:"createdAt"`
	UpdatedAt      time.Time `json:"updatedAt"`
}