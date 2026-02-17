package response

import "time"

type RoomResponse struct {
	RoomID     uint   `json:"room_id"`
	RoomCode   string `json:"room_code"`
	RoomType   string `json:"room_type"`
	Status     string `json:"status"`
	PlayerSlot uint   `json:"player_slot"`
	WSUrl      string `json:"ws_url,omitempty"`
}

type RoomDetailResponse struct {
	RoomID         uint            `json:"room_id"`
	RoomCode       string          `json:"room_code"`
	HostUserID     uint            `json:"host_user_id"`
	RoomType       string          `json:"room_type"`
	Status         string          `json:"status"`
	CurrentPlayers uint            `json:"current_players"`
	MaxPlayers     uint            `json:"max_players"`
	CurrentRound   uint            `json:"current_round"`
	SharedGold     uint            `json:"shared_gold"`
	Players        []PlayerInfo    `json:"players"`
	CreatedAt      time.Time       `json:"created_at"`
	StartedAt      *time.Time      `json:"started_at,omitempty"`
}

type PlayerInfo struct {
	UserID          uint   `json:"user_id"`
	Username        string `json:"username"`
	Slot            uint   `json:"slot"`
	IsReady         bool   `json:"is_ready"`
	IsConnected     bool   `json:"is_connected"`
	Kills           uint   `json:"kills"`
	GoldContributed uint   `json:"gold_contributed"`
}
