package usecase

import (
	"context"
	"errors"
	"time"

	_interface "github.com/JokerTrickster/joker_backend/services/lottoDefenseService/features/towerDefense/model/interface"
	"github.com/JokerTrickster/joker_backend/services/lottoDefenseService/features/towerDefense/model/request"
	"github.com/JokerTrickster/joker_backend/services/lottoDefenseService/features/towerDefense/model/response"
	"gorm.io/gorm"
)

type TDQuestUseCase struct {
	questRepo _interface.ITDQuestRepository
	userRepo  _interface.ITDUserRepository
	timeout   time.Duration
}

func NewTDQuestUseCase(questRepo _interface.ITDQuestRepository, userRepo _interface.ITDUserRepository, timeout time.Duration) _interface.ITDQuestUseCase {
	return &TDQuestUseCase{
		questRepo: questRepo,
		userRepo:  userRepo,
		timeout:   timeout,
	}
}

func (u *TDQuestUseCase) GetActiveQuests(ctx context.Context, userID uint) ([]response.QuestResponse, error) {
	quests, err := u.questRepo.GetActiveQuests(ctx, userID)
	if err != nil {
		return nil, err
	}

	resp := make([]response.QuestResponse, len(quests))
	for i, q := range quests {
		progress := float64(0)
		if q.TargetCount > 0 {
			progress = float64(q.CurrentCount) / float64(q.TargetCount)
		}

		resp[i] = response.QuestResponse{
			QuestID:          q.ID,
			QuestType:        q.QuestType,
			QuestName:        q.QuestName,
			QuestDescription: q.QuestDescription,
			TargetCount:      q.TargetCount,
			CurrentCount:     q.CurrentCount,
			Progress:         progress,
			RewardGold:       q.RewardGold,
			RewardItem:       q.RewardItem,
			Status:           q.Status,
			CreatedAt:        q.CreatedAt,
			CompletedAt:      q.CompletedAt,
		}
	}

	return resp, nil
}

func (u *TDQuestUseCase) UpdateQuestProgress(ctx context.Context, userID uint, questID uint, req *request.UpdateQuestProgressRequest) (*response.QuestResponse, error) {
	quest, err := u.questRepo.GetByID(ctx, questID)
	if err != nil {
		return nil, err
	}

	if quest.UserID != userID {
		return nil, errors.New("quest not found")
	}

	if err := u.questRepo.UpdateProgress(ctx, questID, req.Increment); err != nil {
		return nil, err
	}

	// Refetch
	quest, err = u.questRepo.GetByID(ctx, questID)
	if err != nil {
		return nil, err
	}

	// Check completion
	if quest.CurrentCount >= quest.TargetCount && quest.Status == "active" {
		if err := u.questRepo.CompleteQuest(ctx, questID); err != nil {
			return nil, err
		}
		quest.Status = "completed"
		now := time.Now()
		quest.CompletedAt = &now
	}

	progress := float64(quest.CurrentCount) / float64(quest.TargetCount)

	return &response.QuestResponse{
		QuestID:          quest.ID,
		QuestType:        quest.QuestType,
		QuestName:        quest.QuestName,
		QuestDescription: quest.QuestDescription,
		TargetCount:      quest.TargetCount,
		CurrentCount:     quest.CurrentCount,
		Progress:         progress,
		RewardGold:       quest.RewardGold,
		RewardItem:       quest.RewardItem,
		Status:           quest.Status,
		CreatedAt:        quest.CreatedAt,
		CompletedAt:      quest.CompletedAt,
	}, nil
}

func (u *TDQuestUseCase) ClaimReward(ctx context.Context, userID uint, questID uint) (*response.ClaimRewardResponse, error) {
	quest, err := u.questRepo.GetByID(ctx, questID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("quest not found")
		}
		return nil, err
	}

	if quest.UserID != userID {
		return nil, errors.New("quest not found")
	}

	if quest.Status != "completed" {
		return nil, errors.New("quest not completed")
	}

	// Claim quest
	if err := u.questRepo.ClaimQuest(ctx, questID); err != nil {
		return nil, err
	}

	// Update user gold
	stats, err := u.userRepo.GetStats(ctx, userID)
	if err != nil {
		return nil, err
	}

	stats.CurrentGold += quest.RewardGold
	stats.QuestsCompleted++

	if err := u.userRepo.UpdateStats(ctx, stats); err != nil {
		return nil, err
	}

	rewards := []response.Reward{
		{
			Type:   "gold",
			Amount: &quest.RewardGold,
		},
	}

	now := time.Now()
	quest.ClaimedAt = &now
	quest.Status = "claimed"

	progress := float64(quest.CurrentCount) / float64(quest.TargetCount)

	return &response.ClaimRewardResponse{
		Quest: &response.QuestResponse{
			QuestID:          quest.ID,
			QuestType:        quest.QuestType,
			QuestName:        quest.QuestName,
			QuestDescription: quest.QuestDescription,
			TargetCount:      quest.TargetCount,
			CurrentCount:     quest.CurrentCount,
			Progress:         progress,
			RewardGold:       quest.RewardGold,
			RewardItem:       quest.RewardItem,
			Status:           quest.Status,
			CreatedAt:        quest.CreatedAt,
			CompletedAt:      quest.CompletedAt,
		},
		Rewards:   rewards,
		NewGold:   stats.CurrentGold,
		ClaimedAt: now,
	}, nil
}
