package _interface

import (
	"context"

	"github.com/JokerTrickster/joker_backend/services/lottoDefenseService/features/towerDefense/model/request"
	"github.com/JokerTrickster/joker_backend/services/lottoDefenseService/features/towerDefense/model/response"
)

type ITDAuthUseCase interface {
	Register(ctx context.Context, req *request.RegisterRequest) (*response.AuthResponse, error)
	Login(ctx context.Context, req *request.LoginRequest) (*response.AuthResponse, error)
	GetUserInfo(ctx context.Context, userID uint) (*response.UserInfoResponse, error)
}

type ITDGameUseCase interface {
	SaveSingleResult(ctx context.Context, userID uint, req *request.SaveGameResultRequest) (*response.GameResultResponse, error)
	GetGameHistory(ctx context.Context, userID uint, req *request.GameHistoryRequest) (*response.GameHistoryResponse, error)
	GetUserStats(ctx context.Context, userID uint) (*response.UserStatsResponse, error)
	GetWeeklyRankings(ctx context.Context, gameMode string) (*response.RankingResponse, error)
}

type ITDQuestUseCase interface {
	GetActiveQuests(ctx context.Context, userID uint) ([]response.QuestResponse, error)
	UpdateQuestProgress(ctx context.Context, userID uint, questID uint, req *request.UpdateQuestProgressRequest) (*response.QuestResponse, error)
	ClaimReward(ctx context.Context, userID uint, questID uint) (*response.ClaimRewardResponse, error)
}

type ITDRoomUseCase interface {
	CreateRoom(ctx context.Context, userID uint, req *request.CreateRoomRequest) (*response.RoomResponse, error)
	JoinRoom(ctx context.Context, userID uint, req *request.JoinRoomRequest) (*response.RoomResponse, error)
	GetRoom(ctx context.Context, roomID uint) (*response.RoomDetailResponse, error)
	LeaveRoom(ctx context.Context, userID uint, roomID uint) error
	SetReady(ctx context.Context, userID uint, roomID uint, isReady bool) (*response.RoomDetailResponse, error)
}
