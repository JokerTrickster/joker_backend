package _interface

import (
	"context"

	"github.com/JokerTrickster/joker_backend/services/lottoDefenseService/features/towerDefense/model/entity"
)

type ITDUserRepository interface {
	Create(ctx context.Context, user *entity.TDUser) error
	GetByID(ctx context.Context, id uint) (*entity.TDUser, error)
	GetByEmail(ctx context.Context, email string) (*entity.TDUser, error)
	GetByUsername(ctx context.Context, username string) (*entity.TDUser, error)
	Update(ctx context.Context, user *entity.TDUser) error
	UpdateLastLogin(ctx context.Context, id uint) error
	GetStats(ctx context.Context, userID uint) (*entity.TDUserStats, error)
	CreateStats(ctx context.Context, stats *entity.TDUserStats) error
	UpdateStats(ctx context.Context, stats *entity.TDUserStats) error
}

type ITDGameRepository interface {
	Create(ctx context.Context, result *entity.TDGameResult) error
	GetByID(ctx context.Context, id uint) (*entity.TDGameResult, error)
	GetHistory(ctx context.Context, userID uint, gameMode string, limit, offset int) ([]entity.TDGameResult, int64, error)
	GetHighestRound(ctx context.Context, userID uint, gameMode string) (uint, error)
}

type ITDQuestRepository interface {
	Create(ctx context.Context, quest *entity.TDQuest) error
	GetByID(ctx context.Context, id uint) (*entity.TDQuest, error)
	GetActiveQuests(ctx context.Context, userID uint) ([]entity.TDQuest, error)
	UpdateProgress(ctx context.Context, questID uint, increment uint) error
	CompleteQuest(ctx context.Context, questID uint) error
	ClaimQuest(ctx context.Context, questID uint) error
}

type ITDRoomRepository interface {
	Create(ctx context.Context, room *entity.TDRoom) error
	GetByID(ctx context.Context, id uint) (*entity.TDRoom, error)
	GetByCode(ctx context.Context, code string) (*entity.TDRoom, error)
	Update(ctx context.Context, room *entity.TDRoom) error
	Delete(ctx context.Context, id uint) error
	AddPlayer(ctx context.Context, player *entity.TDRoomPlayer) error
	RemovePlayer(ctx context.Context, roomID, userID uint) error
	GetPlayers(ctx context.Context, roomID uint) ([]entity.TDRoomPlayer, error)
	UpdatePlayerReady(ctx context.Context, roomID, userID uint, isReady bool) error
}
