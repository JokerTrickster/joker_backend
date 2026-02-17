package entity

import "time"

// TDQuest represents a user quest
type TDQuest struct {
	ID               uint       `gorm:"primaryKey" json:"id"`
	UserID           uint       `gorm:"not null;index" json:"user_id"`
	QuestType        string     `gorm:"size:50;not null" json:"quest_type"`
	QuestName        string     `gorm:"size:100;not null" json:"quest_name"`
	QuestDescription string     `gorm:"type:text" json:"quest_description,omitempty"`
	TargetCount      uint       `gorm:"not null" json:"target_count"`
	CurrentCount     uint       `gorm:"default:0" json:"current_count"`
	RewardGold       uint       `gorm:"default:0" json:"reward_gold"`
	RewardItem       *string    `gorm:"size:50" json:"reward_item,omitempty"`
	Status           string     `gorm:"size:20;default:active" json:"status"` // 'active' | 'completed' | 'claimed'
	CreatedAt        time.Time  `gorm:"autoCreateTime" json:"created_at"`
	CompletedAt      *time.Time `json:"completed_at,omitempty"`
	ClaimedAt        *time.Time `json:"claimed_at,omitempty"`
}

func (TDQuest) TableName() string {
	return "td_quests"
}

// TDReward represents a reward record
type TDReward struct {
	ID             uint       `gorm:"primaryKey" json:"id"`
	UserID         uint       `gorm:"not null;index" json:"user_id"`
	RewardType     string     `gorm:"size:50;not null" json:"reward_type"`
	RewardSourceID *uint      `json:"reward_source_id,omitempty"`
	GoldAmount     uint       `gorm:"default:0" json:"gold_amount"`
	ItemID         *string    `gorm:"size:50" json:"item_id,omitempty"`
	ItemCount      uint       `gorm:"default:1" json:"item_count"`
	Claimed        bool       `gorm:"default:false" json:"claimed"`
	ClaimedAt      *time.Time `json:"claimed_at,omitempty"`
	CreatedAt      time.Time  `gorm:"autoCreateTime" json:"created_at"`
}

func (TDReward) TableName() string {
	return "td_rewards"
}
