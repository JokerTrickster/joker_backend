package response

import "time"

// CreateSessionResponse represents game session creation response
type CreateSessionResponse struct {
	SessionID    string `json:"sessionId"`
	RoomCode     string `json:"roomCode,omitempty"`
	WebSocketURL string `json:"webSocketUrl"`
}

// GameResultResponse represents game result response
type GameResultResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
}

// SessionInfoResponse represents session information
type SessionInfoResponse struct {
	SessionID    string     `json:"sessionId"`
	RoomCode     string     `json:"roomCode,omitempty"`
	Mode         string     `json:"mode"`
	Status       string     `json:"status"`
	PlayerCount  int        `json:"playerCount"`
	WebSocketURL string     `json:"webSocketUrl,omitempty"`
	StartedAt    *time.Time `json:"startedAt,omitempty"`
	Players      []PlayerInfo `json:"players,omitempty"`
}

// PlayerInfo represents player information in a session
type PlayerInfo struct {
	UserID   string `json:"userId"`
	Nickname string `json:"nickname"`
	AvatarID string `json:"avatarId,omitempty"`
	Level    int    `json:"level"`
}