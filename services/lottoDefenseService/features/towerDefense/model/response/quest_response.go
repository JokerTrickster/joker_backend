package response

import "time"

type QuestResponse struct {
	QuestID          uint       `json:"quest_id"`
	QuestType        string     `json:"quest_type"`
	QuestName        string     `json:"quest_name"`
	QuestDescription string     `json:"quest_description,omitempty"`
	TargetCount      uint       `json:"target_count"`
	CurrentCount     uint       `json:"current_count"`
	Progress         float64    `json:"progress"`
	RewardGold       uint       `json:"reward_gold"`
	RewardItem       *string    `json:"reward_item,omitempty"`
	Status           string     `json:"status"`
	CreatedAt        time.Time  `json:"created_at"`
	CompletedAt      *time.Time `json:"completed_at,omitempty"`
}

type ClaimRewardResponse struct {
	Quest       *QuestResponse `json:"quest"`
	Rewards     []Reward       `json:"rewards"`
	NewGold     uint           `json:"new_gold"`
	ClaimedAt   time.Time      `json:"claimed_at"`
}
