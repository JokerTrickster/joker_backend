package usecase

import (
	"fmt"
	"strconv"
	"time"

	"github.com/JokerTrickster/joker_backend/services/tdService/features/td/model/entity"
	"github.com/JokerTrickster/joker_backend/services/tdService/features/td/model/request"
	"github.com/JokerTrickster/joker_backend/services/tdService/features/td/model/response"
	"github.com/JokerTrickster/joker_backend/services/tdService/features/td/repository"
)

type GameUseCase struct {
	sessionRepo *repository.GameSessionRepository
	playerRepo  *repository.PlayerRepository
	wsBaseURL   string
}

func NewGameUseCase(sessionRepo *repository.GameSessionRepository, playerRepo *repository.PlayerRepository, wsBaseURL string) *GameUseCase {
	return &GameUseCase{
		sessionRepo: sessionRepo,
		playerRepo:  playerRepo,
		wsBaseURL:   wsBaseURL,
	}
}

// CreateSession creates a new game session
func (uc *GameUseCase) CreateSession(userID uint, req *request.CreateSessionRequest) (*response.CreateSessionResponse, error) {
	// Get player
	player, err := uc.playerRepo.GetPlayerByUserID(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get player: %w", err)
	}
	if player == nil {
		return nil, fmt.Errorf("player not found")
	}

	// Create session
	session := &entity.GameSession{
		Mode:        entity.GameMode(req.Mode),
		PlayerCount: req.PlayerCount,
		PlayerID:    player.ID,
	}

	if err := uc.sessionRepo.CreateSession(session); err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	// Generate WebSocket URL
	wsURL := fmt.Sprintf("%s/game/%s", uc.wsBaseURL, session.SessionID)

	return &response.CreateSessionResponse{
		SessionID:    session.SessionID,
		RoomCode:     session.RoomCode,
		WebSocketURL: wsURL,
	}, nil
}

// JoinSession joins an existing game session
func (uc *GameUseCase) JoinSession(userID uint, req *request.JoinSessionRequest) (*response.CreateSessionResponse, error) {
	// Get player
	player, err := uc.playerRepo.GetPlayerByUserID(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get player: %w", err)
	}
	if player == nil {
		return nil, fmt.Errorf("player not found")
	}

	// Get session by room code
	session, err := uc.sessionRepo.GetSessionByRoomCode(req.RoomCode)
	if err != nil {
		return nil, err
	}

	// Join session
	if err := uc.sessionRepo.JoinSession(session.SessionID, player.ID); err != nil {
		return nil, fmt.Errorf("failed to join session: %w", err)
	}

	// Generate WebSocket URL
	wsURL := fmt.Sprintf("%s/game/%s", uc.wsBaseURL, session.SessionID)

	return &response.CreateSessionResponse{
		SessionID:    session.SessionID,
		RoomCode:     session.RoomCode,
		WebSocketURL: wsURL,
	}, nil
}

// SaveResult saves game result
func (uc *GameUseCase) SaveResult(userID uint, req *request.SaveResultRequest) (*response.GameResultResponse, error) {
	// Get player
	player, err := uc.playerRepo.GetPlayerByUserID(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get player: %w", err)
	}
	if player == nil {
		return nil, fmt.Errorf("player not found")
	}

	// Get session
	session, err := uc.sessionRepo.GetSessionByID(req.SessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get session: %w", err)
	}

	// Calculate play time
	playTime := 0
	if session.StartedAt != nil {
		playTime = int(time.Since(*session.StartedAt).Seconds())
	}

	// Save game result
	result := &entity.GameResult{
		SessionID:       req.SessionID,
		PlayerID:        player.ID,
		Score:           req.Score,
		WavesCompleted:  req.WavesCompleted,
		UnitsPlaced:     req.UnitsPlaced,
		EnemiesDefeated: req.EnemiesDefeated,
		Victory:         req.Victory,
		PlayTime:        playTime,
	}

	if err := uc.sessionRepo.SaveGameResult(result); err != nil {
		return nil, fmt.Errorf("failed to save game result: %w", err)
	}

	// Update player stats
	if err := uc.playerRepo.UpdatePlayerStats(player.ID, result); err != nil {
		return nil, fmt.Errorf("failed to update player stats: %w", err)
	}

	// Update player experience
	expGained := uc.calculateExperience(result)
	player.Experience += expGained

	// Level up logic
	for player.Experience >= uc.getRequiredExperience(player.Level) {
		player.Experience -= uc.getRequiredExperience(player.Level)
		player.Level++
	}

	if err := uc.playerRepo.UpdatePlayer(player); err != nil {
		return nil, fmt.Errorf("failed to update player: %w", err)
	}

	// End session if all players have submitted results
	results, _ := uc.sessionRepo.GetGameResults(req.SessionID)
	expectedResults := 1
	if session.Mode == entity.GameModeCoop && session.Player2ID != nil {
		expectedResults = 2
	}

	if len(results) >= expectedResults {
		uc.sessionRepo.EndSession(req.SessionID)
	}

	return &response.GameResultResponse{
		Success: true,
		Message: "Game result saved successfully",
	}, nil
}

// GetProfile gets player profile
func (uc *GameUseCase) GetProfile(userIDStr string) (*response.ProfileResponse, error) {
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID")
	}

	player, err := uc.playerRepo.GetPlayerByUserID(uint(userID))
	if err != nil {
		return nil, fmt.Errorf("failed to get player: %w", err)
	}
	if player == nil {
		return nil, fmt.Errorf("player not found")
	}

	resp := &response.ProfileResponse{
		Nickname:   player.Nickname,
		AvatarID:   player.AvatarID,
		Level:      player.Level,
		Experience: player.Experience,
	}

	// Add stats if available
	if player.Stats.ID != 0 {
		resp.Stats = &response.StatsResponse{
			GamesPlayed:     player.Stats.GamesPlayed,
			Victories:       player.Stats.Victories,
			TotalScore:      player.Stats.TotalScore,
			HighestScore:    player.Stats.HighestScore,
			HighestWave:     player.Stats.HighestWave,
			UnitsPlaced:     player.Stats.UnitsPlaced,
			EnemiesDefeated: player.Stats.EnemiesKilled,
		}
	}

	return resp, nil
}

// GetSessionInfo gets session information
func (uc *GameUseCase) GetSessionInfo(sessionID string) (*response.SessionInfoResponse, error) {
	session, err := uc.sessionRepo.GetSessionByID(sessionID)
	if err != nil {
		return nil, err
	}

	resp := &response.SessionInfoResponse{
		SessionID:    session.SessionID,
		RoomCode:     session.RoomCode,
		Mode:         string(session.Mode),
		Status:       string(session.Status),
		PlayerCount:  session.PlayerCount,
		WebSocketURL: fmt.Sprintf("%s/game/%s", uc.wsBaseURL, session.SessionID),
		StartedAt:    session.StartedAt,
		Players:      []response.PlayerInfo{},
	}

	// Add player info
	if session.Player != nil {
		resp.Players = append(resp.Players, response.PlayerInfo{
			UserID:   strconv.FormatUint(uint64(session.Player.UserID), 10),
			Nickname: session.Player.Nickname,
			AvatarID: session.Player.AvatarID,
			Level:    session.Player.Level,
		})
	}

	if session.Player2 != nil {
		resp.Players = append(resp.Players, response.PlayerInfo{
			UserID:   strconv.FormatUint(uint64(session.Player2.UserID), 10),
			Nickname: session.Player2.Nickname,
			AvatarID: session.Player2.AvatarID,
			Level:    session.Player2.Level,
		})
	}

	return resp, nil
}

// calculateExperience calculates experience gained from game result
func (uc *GameUseCase) calculateExperience(result *entity.GameResult) int {
	exp := int(result.Score / 100) // Base exp from score
	exp += result.WavesCompleted * 10 // Bonus for waves
	exp += result.EnemiesDefeated     // Bonus for enemies
	if result.Victory {
		exp = int(float64(exp) * 1.5) // 50% bonus for victory
	}
	return exp
}

// getRequiredExperience gets required experience for next level
func (uc *GameUseCase) getRequiredExperience(level int) int {
	return level * 1000 // Simple linear progression
}