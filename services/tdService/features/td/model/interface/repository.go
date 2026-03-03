package iface

import "github.com/JokerTrickster/joker_backend/services/tdService/features/td/model/entity"

type PlayerRepository interface {
	CreatePlayer(player *entity.Player) error
	GetPlayerByUserID(userID uint) (*entity.Player, error)
	GetPlayerByID(playerID uint) (*entity.Player, error)
	UpdatePlayer(player *entity.Player) error
	CreateOrUpdateStats(stats *entity.PlayerStats) error
	UpdatePlayerStats(playerID uint, result *entity.GameResult) error
	GetOrCreatePlayer(userID uint, nickname string, avatarID string) (*entity.Player, error)
}

type GameSessionRepository interface {
	CreateSession(session *entity.GameSession) error
	GetSessionByID(sessionID string) (*entity.GameSession, error)
	GetSessionByRoomCode(roomCode string) (*entity.GameSession, error)
	UpdateSession(session *entity.GameSession) error
	JoinSession(sessionID string, playerID uint) error
	StartSession(sessionID string) error
	EndSession(sessionID string) error
	GetActiveSessions(playerID uint) ([]*entity.GameSession, error)
	SaveGameResult(result *entity.GameResult) error
	GetGameResults(sessionID string) ([]*entity.GameResult, error)
	GetPlayerGameHistory(playerID uint, limit int) ([]*entity.GameResult, error)
}
