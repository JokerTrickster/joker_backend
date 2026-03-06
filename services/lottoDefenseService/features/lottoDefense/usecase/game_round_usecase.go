package usecase

import (
	"context"
	"errors"
	"time"

	"github.com/JokerTrickster/joker_backend/services/lottoDefenseService/features/lottoDefense/model/entity"
	_interface "github.com/JokerTrickster/joker_backend/services/lottoDefenseService/features/lottoDefense/model/interface"
	"github.com/JokerTrickster/joker_backend/services/lottoDefenseService/features/lottoDefense/model/request"
	"github.com/JokerTrickster/joker_backend/services/lottoDefenseService/features/lottoDefense/model/response"
	"github.com/JokerTrickster/joker_backend/services/lottoDefenseService/pkg/lotto"
	"gorm.io/gorm"
)

var (
	ErrRoundNotFound  = errors.New("round not found or not owned by you")
	ErrRoundCompleted = errors.New("round is already completed")
)

type GameRoundUseCase struct {
	roundRepo _interface.IGameRoundRepository
	drawRepo  _interface.ILottoDrawRepository
	timeout   time.Duration
}

func NewGameRoundUseCase(roundRepo _interface.IGameRoundRepository, drawRepo _interface.ILottoDrawRepository, timeout time.Duration) _interface.IGameRoundUseCase {
	return &GameRoundUseCase{
		roundRepo: roundRepo,
		drawRepo:  drawRepo,
		timeout:   timeout,
	}
}

func (u *GameRoundUseCase) StartRound(ctx context.Context, userID uint) (*response.RoundResponse, error) {
	now := time.Now()
	round := &entity.GameRound{
		UserID:    userID,
		Status:    entity.RoundStatusActive,
		StartedAt: &now,
	}
	if err := u.roundRepo.Create(ctx, round); err != nil {
		return nil, err
	}
	return roundToResponse(round), nil
}

func (u *GameRoundUseCase) EndRound(ctx context.Context, userID uint, roundID uint, req *request.EndRoundRequest) (*response.RoundWithDrawResponse, error) {
	round, err := u.roundRepo.GetByIDAndUser(ctx, roundID, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrRoundNotFound
		}
		return nil, err
	}
	if round.Status != entity.RoundStatusActive {
		return nil, ErrRoundCompleted
	}

	now := time.Now()
	round.Status = entity.RoundStatusCompleted
	round.Score = &req.Score
	round.EndedAt = &now
	if err := u.roundRepo.Update(ctx, round); err != nil {
		return nil, err
	}

	numbers := lotto.Draw()
	draw := &entity.LottoDraw{
		RoundID: roundID,
		Numbers: entity.LottoNumbers(numbers),
	}
	if err := u.drawRepo.Create(ctx, draw); err != nil {
		return nil, err
	}

	resp := roundToResponse(round)
	return &response.RoundWithDrawResponse{
		RoundResponse: *resp,
		Numbers:       numbers[:],
	}, nil
}

func (u *GameRoundUseCase) GetMyRounds(ctx context.Context, userID uint, limit int) ([]response.RoundResponse, error) {
	rounds, err := u.roundRepo.ListByUserID(ctx, userID, limit)
	if err != nil {
		return nil, err
	}
	out := make([]response.RoundResponse, len(rounds))
	for i := range rounds {
		out[i] = *roundToResponse(&rounds[i])
	}
	return out, nil
}

func (u *GameRoundUseCase) GetRound(ctx context.Context, userID uint, roundID uint) (*response.RoundWithDrawResponse, error) {
	round, err := u.roundRepo.GetByIDAndUser(ctx, roundID, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrRoundNotFound
		}
		return nil, err
	}
	resp := roundToResponse(round)
	out := &response.RoundWithDrawResponse{RoundResponse: *resp}
	if round.Status == entity.RoundStatusCompleted {
		draw, err := u.drawRepo.GetByRoundID(ctx, roundID)
		if err == nil {
			out.Numbers = draw.Numbers[:]
		}
	}
	return out, nil
}

func roundToResponse(r *entity.GameRound) *response.RoundResponse {
	return &response.RoundResponse{
		ID:        r.ID,
		UserID:    r.UserID,
		Status:    string(r.Status),
		Score:     r.Score,
		StartedAt: r.StartedAt,
		EndedAt:   r.EndedAt,
		CreatedAt: r.CreatedAt,
	}
}
