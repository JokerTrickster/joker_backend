package entity

import "time"

// TDRoom represents a co-op game room
type TDRoom struct {
	ID             uint            `gorm:"primaryKey" json:"id"`
	RoomCode       string          `gorm:"size:4;uniqueIndex;not null" json:"room_code"`
	HostUserID     uint            `gorm:"not null;index" json:"host_user_id"`
	RoomType       string          `gorm:"size:20;not null" json:"room_type"` // 'random' | 'private'
	MaxPlayers     uint            `gorm:"default:2" json:"max_players"`
	CurrentPlayers uint            `gorm:"default:1" json:"current_players"`
	Status         string          `gorm:"size:20;default:waiting" json:"status"` // 'waiting' | 'playing' | 'finished'
	CurrentRound   uint            `gorm:"default:1" json:"current_round"`
	SharedGold     uint            `gorm:"default:100" json:"shared_gold"`
	Player1State   string          `gorm:"type:json" json:"player1_state,omitempty"` // JSON: {round, hp, gold, kills, timestamp, is_alive}
	Player2State   string          `gorm:"type:json" json:"player2_state,omitempty"` // JSON: {round, hp, gold, kills, timestamp, is_alive}
	LastP1Update   *time.Time      `json:"last_p1_update,omitempty"`
	LastP2Update   *time.Time      `json:"last_p2_update,omitempty"`
	Players        []TDRoomPlayer  `gorm:"foreignKey:RoomID" json:"players,omitempty"` // For Preload
	HostUser       *TDUser         `gorm:"foreignKey:HostUserID" json:"host_user,omitempty"` // For Preload
	CreatedAt      time.Time       `gorm:"autoCreateTime" json:"created_at"`
	StartedAt      *time.Time      `json:"started_at,omitempty"`
	FinishedAt     *time.Time      `json:"finished_at,omitempty"`
	ExpiresAt      time.Time       `json:"expires_at"`
}

func (TDRoom) TableName() string {
	return "td_rooms"
}

// TDRoomPlayer represents a player in a room
type TDRoomPlayer struct {
	ID              uint       `gorm:"primaryKey" json:"id"`
	RoomID          uint       `gorm:"not null;index:idx_room_slot" json:"room_id"`
	UserID          uint       `gorm:"not null" json:"user_id"`
	User            *TDUser    `gorm:"foreignKey:UserID" json:"user,omitempty"` // For Preload
	PlayerSlot      uint       `gorm:"not null;index:idx_room_slot" json:"player_slot"` // 0 | 1
	IsReady         bool       `gorm:"default:false" json:"is_ready"`
	IsConnected     bool       `gorm:"default:true" json:"is_connected"`
	Kills           uint       `gorm:"default:0" json:"kills"`
	GoldContributed uint       `gorm:"default:0" json:"gold_contributed"`
	JoinedAt        time.Time  `gorm:"autoCreateTime" json:"joined_at"`
	LeftAt          *time.Time `json:"left_at,omitempty"`
}

func (TDRoomPlayer) TableName() string {
	return "td_room_players"
}

// TDFriendship represents a friend relationship
type TDFriendship struct {
	ID         uint       `gorm:"primaryKey" json:"id"`
	UserID     uint       `gorm:"not null;index:idx_user_friend" json:"user_id"`
	FriendID   uint       `gorm:"not null;index:idx_user_friend" json:"friend_id"`
	Status     string     `gorm:"size:20;default:pending" json:"status"` // 'pending' | 'accepted' | 'blocked'
	CreatedAt  time.Time  `gorm:"autoCreateTime" json:"created_at"`
	AcceptedAt *time.Time `json:"accepted_at,omitempty"`
}

func (TDFriendship) TableName() string {
	return "td_friendships"
}
