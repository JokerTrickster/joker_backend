package repository

import (
	"context"

	"github.com/JokerTrickster/joker_backend/services/lottoDefenseService/features/lottoDefense/model/entity"
	_interface "github.com/JokerTrickster/joker_backend/services/lottoDefenseService/features/lottoDefense/model/interface"
	"gorm.io/gorm"
)

type GameRoundRepository struct {
	db *gorm.DB
}

func NewGameRoundRepository(db *gorm.DB) _interface.IGameRoundRepository {
	return &GameRoundRepository{db: db}
}

func (r *GameRoundRepository) Create(ctx context.Context, round *entity.GameRound) error {
	return r.db.WithContext(ctx).Create(round).Error
}

func (r *GameRoundRepository) GetByID(ctx context.Context, id uint) (*entity.GameRound, error) {
	var round entity.GameRound
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&round).Error
	if err != nil {
		return nil, err
	}
	return &round, nil
}

func (r *GameRoundRepository) GetByIDAndUser(ctx context.Context, id, userID uint) (*entity.GameRound, error) {
	var round entity.GameRound
	err := r.db.WithContext(ctx).Where("id = ? AND user_id = ?", id, userID).First(&round).Error
	if err != nil {
		return nil, err
	}
	return &round, nil
}

func (r *GameRoundRepository) Update(ctx context.Context, round *entity.GameRound) error {
	return r.db.WithContext(ctx).Save(round).Error
}

func (r *GameRoundRepository) ListByUserID(ctx context.Context, userID uint, limit int) ([]entity.GameRound, error) {
	if limit <= 0 {
		limit = 20
	}
	var rounds []entity.GameRound
	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Limit(limit).
		Find(&rounds).Error
	return rounds, err
}

func (r *GameRoundRepository) TopScores(ctx context.Context, limit int) ([]entity.GameRound, error) {
	if limit <= 0 {
		limit = 10
	}
	var rounds []entity.GameRound
	err := r.db.WithContext(ctx).
		Where("status = ? AND score IS NOT NULL", entity.RoundStatusCompleted).
		Order("score DESC").
		Limit(limit).
		Find(&rounds).Error
	return rounds, err
}

func (r *GameRoundRepository) Leaderboard(ctx context.Context, limit int) ([]_interface.LeaderboardRow, error) {
	if limit <= 0 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}
	var rows []_interface.LeaderboardRow
	err := r.db.WithContext(ctx).Raw(`
		SELECT u.id AS user_id, u.name, r.best_score AS score
		FROM (
			SELECT user_id, MAX(score) AS best_score
			FROM game_rounds
			WHERE status = ? AND score IS NOT NULL
			GROUP BY user_id
			ORDER BY best_score DESC
			LIMIT ?
		) r
		INNER JOIN users u ON u.id = r.user_id AND (u.deleted_at IS NULL)
		ORDER BY r.best_score DESC
	`, entity.RoundStatusCompleted, limit).Scan(&rows).Error
	return rows, err
}
