package response

import "time"

type GameResultResponse struct {
	GameID          uint   `json:"game_id"`
	NewHighestRound uint   `json:"new_highest_round"`
	Rewards         []Reward `json:"rewards"`
}

type Reward struct {
	Type   string `json:"type"`
	Amount *uint  `json:"amount,omitempty"`
	ItemID *string `json:"item_id,omitempty"`
}

type GameHistoryResponse struct {
	Total int64             `json:"total"`
	Games []GameHistoryItem `json:"games"`
}

type GameHistoryItem struct {
	GameID              uint      `json:"game_id"`
	GameMode            string    `json:"game_mode"`
	RoundsReached       uint      `json:"rounds_reached"`
	MonstersKilled      uint      `json:"monsters_killed"`
	GoldEarned          uint      `json:"gold_earned"`
	SurvivalTimeSeconds *uint     `json:"survival_time_seconds,omitempty"`
	Result              string    `json:"result"`
	PlayedAt            time.Time `json:"played_at"`
}
