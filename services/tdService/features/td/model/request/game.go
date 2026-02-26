package request

// CreateSessionRequest represents game session creation request
type CreateSessionRequest struct {
	Mode        string `json:"mode" validate:"required,oneof=single coop"`
	PlayerCount int    `json:"playerCount" validate:"required,min=1,max=2"`
}

// JoinSessionRequest represents join session request
type JoinSessionRequest struct {
	RoomCode string `json:"roomCode" validate:"required"`
}

// SaveResultRequest represents game result save request
type SaveResultRequest struct {
	SessionID       string `json:"sessionId" validate:"required"`
	Score           int64  `json:"score" validate:"min=0"`
	WavesCompleted  int    `json:"wavesCompleted" validate:"min=0"`
	UnitsPlaced     int    `json:"unitsPlaced" validate:"min=0"`
	EnemiesDefeated int    `json:"enemiesDefeated" validate:"min=0"`
	Victory         bool   `json:"victory"`
}