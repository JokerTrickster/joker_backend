package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/JokerTrickster/joker_backend/services/cloudRepositoryService/features/cloudRepository/model/entity"
	_interface "github.com/JokerTrickster/joker_backend/services/cloudRepositoryService/features/cloudRepository/model/interface"
	"github.com/JokerTrickster/joker_backend/services/cloudRepositoryService/features/cloudRepository/model/response"
)

type DownloadCloudRepositoryUseCase struct {
	Repo           _interface.IDownloadCloudRepositoryRepository
	StatsRepo      _interface.IUserStatsCloudRepositoryRepository
	ContextTimeout time.Duration
}

func NewDownloadCloudRepositoryUseCase(repo _interface.IDownloadCloudRepositoryRepository, statsRepo _interface.IUserStatsCloudRepositoryRepository, timeout time.Duration) _interface.IDownloadCloudRepositoryUseCase {
	return &DownloadCloudRepositoryUseCase{
		Repo:           repo,
		StatsRepo:      statsRepo,
		ContextTimeout: timeout,
	}
}

// RequestDownloadURL generates a presigned download URL for a file
func (u *DownloadCloudRepositoryUseCase) RequestDownloadURL(ctx context.Context, userID, fileID uint) (*response.DownloadResponseDTO, error) {
	// Get file from database
	file, err := u.Repo.GetFileByID(ctx, fileID)
	if err != nil {
		return nil, fmt.Errorf("file not found: %w", err)
	}

	// Check if user owns the file
	if file.UserID != userID {
		return nil, fmt.Errorf("unauthorized access to file")
	}

	// Log download activity
	if u.StatsRepo != nil {
		activity := &entity.ActivityLog{
			UserID:       userID,
			FileID:       &fileID,
			ActivityType: entity.ActivityTypeDownload,
		}
		_ = u.StatsRepo.LogActivity(ctx, activity) // Don't fail on logging error
	}

	// Generate presigned download URL with Content-Disposition header for forced download
	downloadURL, err := u.Repo.GeneratePresignedDownloadURLWithFilename(ctx, file.S3Key, file.FileName, 1*time.Hour)
	if err != nil {
		return nil, fmt.Errorf("failed to generate download URL: %w", err)
	}

	return &response.DownloadResponseDTO{
		DownloadURL: downloadURL,
		FileName:    file.FileName,
		ExpiresIn:   int(time.Hour.Seconds()),
	}, nil
}
