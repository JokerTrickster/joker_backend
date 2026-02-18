package entity

import "time"

// TDGameResult represents a game result record
type TDGameResult struct {
	ID                  uint       `gorm:"primaryKey" json:"id"`
	UserID              uint       `gorm:"not null;index" json:"user_id"`
	User                *TDUser    `gorm:"foreignKey:UserID" json:"user,omitempty"` // For Preload
	GameMode            string     `gorm:"size:20;not null" json:"game_mode"` // 'single' | 'coop'
	RoomID              *uint      `json:"room_id,omitempty"`
	Room                *TDRoom    `gorm:"foreignKey:RoomID" json:"room,omitempty"` // For co-op rankings
	RoundsReached       uint       `gorm:"not null" json:"rounds_reached"`
	MonstersKilled      uint       `gorm:"not null" json:"monsters_killed"`
	GoldEarned          uint       `gorm:"not null" json:"gold_earned"`
	SurvivalTimeSeconds *uint      `json:"survival_time_seconds,omitempty"`
	FinalArmyValue      *uint      `json:"final_army_value,omitempty"`
	Result              string     `gorm:"size:20" json:"result"` // 'victory' | 'defeat' | 'disconnect'
	PlayedAt            time.Time  `gorm:"autoCreateTime" json:"played_at"`
}

func (TDGameResult) TableName() string {
	return "td_game_results"
}
