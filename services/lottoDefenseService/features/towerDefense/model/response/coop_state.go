package response

// OpponentStateResponse represents opponent's current game state
type OpponentStateResponse struct {
	OpponentID   uint   `json:"opponent_id"`
	OpponentName string `json:"opponent_name"`
	Round        int    `json:"round"`
	HP           int    `json:"hp"`
	Gold         int    `json:"gold"`
	Kills        int    `json:"kills"`
	LastUpdate   int64  `json:"last_update"`
	IsAlive      bool   `json:"is_alive"`
}
