package usecase

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"time"

	"github.com/JokerTrickster/joker_backend/services/lottoDefenseService/features/towerDefense/model/entity"
	_interface "github.com/JokerTrickster/joker_backend/services/lottoDefenseService/features/towerDefense/model/interface"
	"github.com/JokerTrickster/joker_backend/services/lottoDefenseService/features/towerDefense/model/request"
	"github.com/JokerTrickster/joker_backend/services/lottoDefenseService/features/towerDefense/model/response"
	"gorm.io/gorm"
)

type TDRoomUseCase struct {
	roomRepo _interface.ITDRoomRepository
	userRepo _interface.ITDUserRepository
	timeout  time.Duration
}

func NewTDRoomUseCase(roomRepo _interface.ITDRoomRepository, userRepo _interface.ITDUserRepository, timeout time.Duration) _interface.ITDRoomUseCase {
	return &TDRoomUseCase{
		roomRepo: roomRepo,
		userRepo: userRepo,
		timeout:  timeout,
	}
}

func (u *TDRoomUseCase) CreateRoom(ctx context.Context, userID uint, req *request.CreateRoomRequest) (*response.RoomResponse, error) {
	// Generate 4-digit code
	code := generateRoomCode()

	room := &entity.TDRoom{
		RoomCode:       code,
		HostUserID:     userID,
		RoomType:       req.RoomType,
		MaxPlayers:     2,
		CurrentPlayers: 1,
		Status:         "waiting",
		CurrentRound:   1,
		SharedGold:     100,
		ExpiresAt:      time.Now().Add(30 * time.Minute),
	}

	if err := u.roomRepo.Create(ctx, room); err != nil {
		return nil, err
	}

	// Add host as player
	player := &entity.TDRoomPlayer{
		RoomID:     room.ID,
		UserID:     userID,
		PlayerSlot: 0,
		IsReady:    false,
	}

	if err := u.roomRepo.AddPlayer(ctx, player); err != nil {
		return nil, err
	}

	return &response.RoomResponse{
		RoomID:     room.ID,
		RoomCode:   room.RoomCode,
		RoomType:   room.RoomType,
		Status:     room.Status,
		PlayerSlot: 0,
	}, nil
}

func (u *TDRoomUseCase) JoinRoom(ctx context.Context, userID uint, req *request.JoinRoomRequest) (*response.RoomResponse, error) {
	room, err := u.roomRepo.GetByCode(ctx, req.RoomCode)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("room not found")
		}
		return nil, err
	}

	if room.CurrentPlayers >= room.MaxPlayers {
		return nil, errors.New("room is full")
	}

	if room.Status != "waiting" {
		return nil, errors.New("room already started")
	}

	// Add player
	player := &entity.TDRoomPlayer{
		RoomID:     room.ID,
		UserID:     userID,
		PlayerSlot: 1,
		IsReady:    false,
	}

	if err := u.roomRepo.AddPlayer(ctx, player); err != nil {
		return nil, err
	}

	// Update room
	room.CurrentPlayers = 2
	if err := u.roomRepo.Update(ctx, room); err != nil {
		return nil, err
	}

	return &response.RoomResponse{
		RoomID:     room.ID,
		RoomCode:   room.RoomCode,
		RoomType:   room.RoomType,
		Status:     room.Status,
		PlayerSlot: 1,
		WSUrl:      fmt.Sprintf("ws://localhost:8080/ws/coop/%d", room.ID),
	}, nil
}

func (u *TDRoomUseCase) GetRoom(ctx context.Context, roomID uint) (*response.RoomDetailResponse, error) {
	room, err := u.roomRepo.GetByID(ctx, roomID)
	if err != nil {
		return nil, err
	}

	players, err := u.roomRepo.GetPlayers(ctx, roomID)
	if err != nil {
		return nil, err
	}

	playerInfos := make([]response.PlayerInfo, len(players))
	for i, p := range players {
		user, _ := u.userRepo.GetByID(ctx, p.UserID)
		username := ""
		if user != nil {
			username = user.Username
		}

		playerInfos[i] = response.PlayerInfo{
			UserID:          p.UserID,
			Username:        username,
			Slot:            p.PlayerSlot,
			IsReady:         p.IsReady,
			IsConnected:     p.IsConnected,
			Kills:           p.Kills,
			GoldContributed: p.GoldContributed,
		}
	}

	return &response.RoomDetailResponse{
		RoomID:         room.ID,
		RoomCode:       room.RoomCode,
		HostUserID:     room.HostUserID,
		RoomType:       room.RoomType,
		Status:         room.Status,
		CurrentPlayers: room.CurrentPlayers,
		MaxPlayers:     room.MaxPlayers,
		CurrentRound:   room.CurrentRound,
		SharedGold:     room.SharedGold,
		Players:        playerInfos,
		CreatedAt:      room.CreatedAt,
		StartedAt:      room.StartedAt,
	}, nil
}

func (u *TDRoomUseCase) LeaveRoom(ctx context.Context, userID uint, roomID uint) error {
	return u.roomRepo.RemovePlayer(ctx, roomID, userID)
}

func (u *TDRoomUseCase) SetReady(ctx context.Context, userID uint, roomID uint, isReady bool) (*response.RoomDetailResponse, error) {
	if err := u.roomRepo.UpdatePlayerReady(ctx, roomID, userID, isReady); err != nil {
		return nil, err
	}

	return u.GetRoom(ctx, roomID)
}

func (u *TDRoomUseCase) UpdatePlayerState(ctx context.Context, roomID, userID uint, req *request.UpdateGameStateRequest) error {
	// Convert request to JSON string
	stateJSON := fmt.Sprintf(`{"round":%d,"hp":%d,"gold":%d,"kills":%d,"timestamp":%d,"is_alive":%t}`,
		req.Round, req.HP, req.Gold, req.Kills, req.Timestamp, req.HP > 0)
	
	return u.roomRepo.UpdatePlayerState(ctx, roomID, userID, stateJSON)
}

func (u *TDRoomUseCase) GetOpponentState(ctx context.Context, roomID, userID uint) (*response.OpponentStateResponse, error) {
	stateJSON, opponentID, opponentName, err := u.roomRepo.GetOpponentState(ctx, roomID, userID)
	if err != nil {
		return nil, err
	}

	if stateJSON == "" {
		return nil, errors.New("opponent not found")
	}

	// Parse JSON state
	var state struct {
		Round     int   `json:"round"`
		HP        int   `json:"hp"`
		Gold      int   `json:"gold"`
		Kills     int   `json:"kills"`
		Timestamp int64 `json:"timestamp"`
		IsAlive   bool  `json:"is_alive"`
	}

	if err := json.Unmarshal([]byte(stateJSON), &state); err != nil {
		return nil, fmt.Errorf("failed to parse opponent state: %w", err)
	}

	return &response.OpponentStateResponse{
		OpponentID:   opponentID,
		OpponentName: opponentName,
		Round:        state.Round,
		HP:           state.HP,
		Gold:         state.Gold,
		Kills:        state.Kills,
		LastUpdate:   state.Timestamp,
		IsAlive:      state.IsAlive,
	}, nil
}

func generateRoomCode() string {
	const chars = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"
	b := make([]byte, 4)
	for i := range b {
		b[i] = chars[rand.Intn(len(chars))]
	}
	return string(b)
}
