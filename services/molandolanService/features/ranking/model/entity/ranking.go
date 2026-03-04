package entity

import "time"

type Ranking struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	UserID      uint      `gorm:"not null;uniqueIndex:idx_rankings_user_game" json:"userId"`
	GameType    string    `gorm:"size:30;not null;uniqueIndex:idx_rankings_user_game;index:idx_rankings_game_time" json:"gameType"`
	Nickname    string    `gorm:"size:50;not null" json:"nickname"`
	ClearTimeMs uint      `gorm:"not null;index:idx_rankings_game_time" json:"clearTimeMs"`
	CreatedAt   time.Time `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt   time.Time `gorm:"autoUpdateTime" json:"updatedAt"`
}

func (Ranking) TableName() string {
	return "rankings"
}
