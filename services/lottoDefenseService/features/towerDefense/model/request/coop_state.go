package request

// UpdateGameStateRequest updates player's current game state during co-op play
type UpdateGameStateRequest struct {
	Round     int `json:"round" validate:"required,min=1,max=100"`
	HP        int `json:"hp" validate:"required,min=0"`
	Gold      int `json:"gold" validate:"min=0"`
	Kills     int `json:"kills" validate:"min=0"`
	Timestamp int64 `json:"timestamp"`
}
