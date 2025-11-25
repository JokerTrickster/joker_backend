package usecase

import (
	"context"
	"time"

	_interface "github.com/JokerTrickster/joker_backend/services/cloudRepositoryService/features/cloudRepository/model/interface"
	"github.com/JokerTrickster/joker_backend/services/cloudRepositoryService/features/cloudRepository/model/response"
)

type UserStatsCloudRepositoryUseCase struct {
	Repo           _interface.IUserStatsCloudRepositoryRepository
	ContextTimeout time.Duration
}

func NewUserStatsCloudRepositoryUseCase(repo _interface.IUserStatsCloudRepositoryRepository, timeout time.Duration) _interface.IUserStatsCloudRepositoryUseCase {
	return &UserStatsCloudRepositoryUseCase{
		Repo:           repo,
		ContextTimeout: timeout,
	}
}

// GetUserStats retrieves user statistics including storage and monthly activity
func (u *UserStatsCloudRepositoryUseCase) GetUserStats(ctx context.Context, userID uint) (*response.UserStatsResponseDTO, error) {
	c, cancel := context.WithTimeout(ctx, u.ContextTimeout)
	defer cancel()

	// Get total storage used
	used, err := u.Repo.GetTotalStorageUsed(c, userID)
	if err != nil {
		return nil, err
	}

	// Set total storage limit (15GB for now, can be made configurable)
	const totalStorage int64 = 15 * 1024 * 1024 * 1024 // 15GB in bytes

	// Calculate percentage
	percentage := float64(used) / float64(totalStorage) * 100

	// Get current month stats
	now := time.Now()
	year := now.Year()
	month := int(now.Month())

	// Get monthly upload count
	uploads, err := u.Repo.GetMonthlyUploadCount(c, userID, year, month)
	if err != nil {
		return nil, err
	}

	// Get monthly download count
	downloads, err := u.Repo.GetMonthlyDownloadCount(c, userID, year, month)
	if err != nil {
		return nil, err
	}

	// Get monthly tags created count
	tagsCreated, err := u.Repo.GetMonthlyTagsCreatedCount(c, userID, year, month)
	if err != nil {
		return nil, err
	}

	return &response.UserStatsResponseDTO{
		Storage: response.StorageInfoDTO{
			Used:       used,
			Total:      totalStorage,
			Percentage: percentage,
		},
		MonthlyStats: response.MonthlyStatsDTO{
			Uploads:     uploads,
			Downloads:   downloads,
			TagsCreated: tagsCreated,
		},
	}, nil
}