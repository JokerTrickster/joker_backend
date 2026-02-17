package entity

import "time"

// TDUser represents a tower defense game user
type TDUser struct {
	ID           uint       `gorm:"primaryKey" json:"id"`
	Username     string     `gorm:"size:50;uniqueIndex;not null" json:"username"`
	Email        string     `gorm:"size:100;uniqueIndex;not null" json:"email"`
	PasswordHash string     `gorm:"size:255;not null" json:"-"`
	CreatedAt    time.Time  `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt    time.Time  `gorm:"autoUpdateTime" json:"updated_at"`
	LastLogin    *time.Time `json:"last_login,omitempty"`
	IsActive     bool       `gorm:"default:true" json:"is_active"`
}

func (TDUser) TableName() string {
	return "td_users"
}

// TDUserStats represents user statistics
type TDUserStats struct {
	UserID             uint      `gorm:"primaryKey" json:"user_id"`
	SingleHighestRound uint      `gorm:"default:0" json:"single_highest_round"`
	SingleTotalGames   uint      `gorm:"default:0" json:"single_total_games"`
	SingleTotalKills   uint      `gorm:"default:0" json:"single_total_kills"`
	CoopHighestRound   uint      `gorm:"default:0" json:"coop_highest_round"`
	CoopTotalGames     uint      `gorm:"default:0" json:"coop_total_games"`
	CoopTotalKills     uint      `gorm:"default:0" json:"coop_total_kills"`
	CoopWins           uint      `gorm:"default:0" json:"coop_wins"`
	TotalGoldEarned    uint      `gorm:"default:0" json:"total_gold_earned"`
	CurrentGold        uint      `gorm:"default:0" json:"current_gold"`
	QuestsCompleted    uint      `gorm:"default:0" json:"quests_completed"`
	UpdatedAt          time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

func (TDUserStats) TableName() string {
	return "td_user_stats"
}
