package entity

import (
	"encoding/json"
	"time"
)

// WSMessageType represents WebSocket message types
type WSMessageType string

const (
	// Client to Server
	WSTypeAuth         WSMessageType = "auth"
	WSTypeUnitPlaced   WSMessageType = "unit_placed"
	WSTypeUnitRemoved  WSMessageType = "unit_removed"
	WSTypeHeartbeat    WSMessageType = "heartbeat"
	WSTypeGameStart    WSMessageType = "game_start"
	WSTypeGamePause    WSMessageType = "game_pause"
	WSTypeGameResume   WSMessageType = "game_resume"

	// Server to Client
	WSTypeGameState    WSMessageType = "game_state"
	WSTypeWaveStart    WSMessageType = "wave_start"
	WSTypeWaveComplete WSMessageType = "wave_complete"
	WSTypeGameOver     WSMessageType = "game_over"
	WSTypePlayerJoined WSMessageType = "player_joined"
	WSTypePlayerLeft   WSMessageType = "player_left"
	WSTypeError        WSMessageType = "error"
	WSTypeAck          WSMessageType = "ack"
)

// WSMessage represents a WebSocket message
type WSMessage struct {
	Type      WSMessageType   `json:"type"`
	Payload   json.RawMessage `json:"payload"`
	Timestamp int64           `json:"timestamp"`
}

// NewWSMessage creates a new WebSocket message
func NewWSMessage(msgType WSMessageType, payload interface{}) (*WSMessage, error) {
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	return &WSMessage{
		Type:      msgType,
		Payload:   payloadBytes,
		Timestamp: time.Now().UnixMilli(),
	}, nil
}

// Auth message payloads
type AuthPayload struct {
	Token     string `json:"token"`
	SessionID string `json:"sessionId"`
}

// Unit action payloads
type UnitActionPayload struct {
	PlayerNumber int      `json:"playerNumber"`
	Position     Position `json:"position"`
	UnitType     string   `json:"unitType,omitempty"`
}

// Position represents a grid position
type Position struct {
	X int `json:"x"`
	Y int `json:"y"`
}

// Game state payloads
type GameStatePayload struct {
	Wave     int           `json:"wave"`
	Enemies  []EnemyState  `json:"enemies"`
	Players  []PlayerState `json:"players"`
	GameTime int           `json:"gameTime"` // in seconds
}

// EnemyState represents enemy state in game
type EnemyState struct {
	ID       string   `json:"id"`
	Type     string   `json:"type"`
	Position Position `json:"position"`
	Health   int      `json:"health"`
	MaxHealth int     `json:"maxHealth"`
}

// PlayerState represents player state in game
type PlayerState struct {
	PlayerNumber int    `json:"playerNumber"`
	UserID       string `json:"userId"`
	Nickname     string `json:"nickname"`
	Gold         int    `json:"gold"`
	Lives        int    `json:"lives"`
	Score        int64  `json:"score"`
}

// Wave payloads
type WaveStartPayload struct {
	WaveNumber int `json:"waveNumber"`
	EnemyCount int `json:"enemyCount"`
	Duration   int `json:"duration"` // in seconds
}

type WaveCompletePayload struct {
	WaveNumber    int `json:"waveNumber"`
	BonusGold     int `json:"bonusGold"`
	NextWaveDelay int `json:"nextWaveDelay"` // in seconds
}

// Game over payload
type GameOverPayload struct {
	Victory bool                   `json:"victory"`
	Scores  map[string]PlayerScore `json:"scores"`
}

type PlayerScore struct {
	UserID          string `json:"userId"`
	Score           int64  `json:"score"`
	WavesCompleted  int    `json:"wavesCompleted"`
	EnemiesDefeated int    `json:"enemiesDefeated"`
}

// Error payload
type ErrorPayload struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}