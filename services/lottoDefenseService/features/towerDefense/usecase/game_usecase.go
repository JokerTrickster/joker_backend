package usecase

import (
	"context"
	"time"

	"github.com/JokerTrickster/joker_backend/services/lottoDefenseService/features/towerDefense/model/entity"
	_interface "github.com/JokerTrickster/joker_backend/services/lottoDefenseService/features/towerDefense/model/interface"
	"github.com/JokerTrickster/joker_backend/services/lottoDefenseService/features/towerDefense/model/request"
	"github.com/JokerTrickster/joker_backend/services/lottoDefenseService/features/towerDefense/model/response"
)

type TDGameUseCase struct {
	gameRepo _interface.ITDGameRepository
	userRepo _interface.ITDUserRepository
	timeout  time.Duration
}

func NewTDGameUseCase(gameRepo _interface.ITDGameRepository, userRepo _interface.ITDUserRepository, timeout time.Duration) _interface.ITDGameUseCase {
	return &TDGameUseCase{
		gameRepo: gameRepo,
		userRepo: userRepo,
		timeout:  timeout,
	}
}

func (u *TDGameUseCase) SaveSingleResult(ctx context.Context, userID uint, req *request.SaveGameResultRequest) (*response.GameResultResponse, error) {
	// Save game result
	result := &entity.TDGameResult{
		UserID:              userID,
		GameMode:            req.GameMode,
		RoundsReached:       req.RoundsReached,
		MonstersKilled:      req.MonstersKilled,
		GoldEarned:          req.GoldEarned,
		SurvivalTimeSeconds: req.SurvivalTimeSeconds,
		FinalArmyValue:      req.FinalArmyValue,
		Result:              req.Result,
	}

	if err := u.gameRepo.Create(ctx, result); err != nil {
		return nil, err
	}

	// Update user stats
	stats, err := u.userRepo.GetStats(ctx, userID)
	if err != nil {
		return nil, err
	}

	if req.GameMode == "single" {
		stats.SingleTotalGames++
		stats.SingleTotalKills += req.MonstersKilled
		if req.RoundsReached > stats.SingleHighestRound {
			stats.SingleHighestRound = req.RoundsReached
		}
	} else {
		stats.CoopTotalGames++
		stats.CoopTotalKills += req.MonstersKilled
		if req.RoundsReached > stats.CoopHighestRound {
			stats.CoopHighestRound = req.RoundsReached
		}
		if req.Result == "victory" {
			stats.CoopWins++
		}
	}

	stats.TotalGoldEarned += req.GoldEarned
	stats.CurrentGold += req.GoldEarned

	if err := u.userRepo.UpdateStats(ctx, stats); err != nil {
		return nil, err
	}

	rewards := []response.Reward{
		{
			Type:   "gold",
			Amount: &req.GoldEarned,
		},
	}

	newHighestRound := stats.SingleHighestRound
	if req.GameMode == "coop" {
		newHighestRound = stats.CoopHighestRound
	}

	return &response.GameResultResponse{
		GameID:          result.ID,
		NewHighestRound: newHighestRound,
		Rewards:         rewards,
	}, nil
}

func (u *TDGameUseCase) GetGameHistory(ctx context.Context, userID uint, req *request.GameHistoryRequest) (*response.GameHistoryResponse, error) {
	if req.Limit == 0 {
		req.Limit = 10
	}
	if req.Limit > 50 {
		req.Limit = 50
	}

	results, total, err := u.gameRepo.GetHistory(ctx, userID, req.GameMode, req.Limit, req.Offset)
	if err != nil {
		return nil, err
	}

	games := make([]response.GameHistoryItem, len(results))
	for i, r := range results {
		games[i] = response.GameHistoryItem{
			GameID:              r.ID,
			GameMode:            r.GameMode,
			RoundsReached:       r.RoundsReached,
			MonstersKilled:      r.MonstersKilled,
			GoldEarned:          r.GoldEarned,
			SurvivalTimeSeconds: r.SurvivalTimeSeconds,
			Result:              r.Result,
			PlayedAt:            r.PlayedAt,
		}
	}

	return &response.GameHistoryResponse{
		Total: total,
		Games: games,
	}, nil
}

func (u *TDGameUseCase) GetUserStats(ctx context.Context, userID uint) (*response.UserStatsResponse, error) {
	stats, err := u.userRepo.GetStats(ctx, userID)
	if err != nil {
		return nil, err
	}

	var singleAvg float64
	if stats.SingleTotalGames > 0 {
		singleAvg = float64(stats.SingleTotalKills) / float64(stats.SingleTotalGames)
	}

	var coopWinRate float64
	if stats.CoopTotalGames > 0 {
		coopWinRate = float64(stats.CoopWins) / float64(stats.CoopTotalGames)
	}

	return &response.UserStatsResponse{
		Single: response.SingleStats{
			HighestRound: stats.SingleHighestRound,
			TotalGames:   stats.SingleTotalGames,
			TotalKills:   stats.SingleTotalKills,
			AverageRound: singleAvg,
		},
		Coop: response.CoopStats{
			HighestRound: stats.CoopHighestRound,
			TotalGames:   stats.CoopTotalGames,
			TotalKills:   stats.CoopTotalKills,
			Wins:         stats.CoopWins,
			WinRate:      coopWinRate,
		},
		Gold: response.GoldStats{
			TotalEarned: stats.TotalGoldEarned,
			Current:     stats.CurrentGold,
		},
	}, nil
}

func (u *TDGameUseCase) GetWeeklyRankings(ctx context.Context, gameMode string) (*response.RankingResponse, error) {
	// Get top 10 results from last 7 days
	results, err := u.gameRepo.GetWeeklyRankings(ctx, gameMode, 10)
	if err != nil {
		return nil, err
	}

	rankings := make([]response.RankingItem, len(results))
	for i, result := range results {
		item := response.RankingItem{
			Rank:                i + 1,
			UserID:              result.UserID,
			RoundsReached:       result.RoundsReached,
			SurvivalTimeSeconds: result.SurvivalTimeSeconds,
			PlayedAt:            result.PlayedAt.Format("2006-01-02 15:04:05"),
		}

		// Calculate survival minutes
		if result.SurvivalTimeSeconds != nil {
			minutes := float64(*result.SurvivalTimeSeconds) / 60.0
			item.SurvivalMinutes = minutes
		}

		// Get username from User relation
		if result.User != nil {
			item.Username = result.User.Username
		}

		// For co-op mode, get player 2 info from room
		if gameMode == "coop" && result.Room != nil {
			// Find the other player in the room
			// This would require a more complex query or additional repository method
			// For now, we'll handle it simply
			// TODO: Implement proper co-op player 2 lookup
		}

		rankings[i] = item
	}

	return &response.RankingResponse{
		GameMode: gameMode,
		Rankings: rankings,
	}, nil
}
