package response

// RankingResponse represents weekly rankings
type RankingResponse struct {
	GameMode string        `json:"game_mode"`
	Rankings []RankingItem `json:"rankings"`
}

// RankingItem represents a single ranking entry
type RankingItem struct {
	Rank                int     `json:"rank"`
	UserID              uint    `json:"user_id"`
	Username            string  `json:"username"`
	RoundsReached       uint    `json:"rounds_reached"`
	SurvivalTimeSeconds *uint   `json:"survival_time_seconds,omitempty"`
	SurvivalMinutes     float64 `json:"survival_minutes,omitempty"` // Converted for UI
	PlayedAt            string  `json:"played_at"`
	
	// For co-op mode only
	Player2ID       *uint   `json:"player2_id,omitempty"`
	Player2Username *string `json:"player2_username,omitempty"`
}
