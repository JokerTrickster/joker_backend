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
	FileShareRepo  _interface.IFileShareRepository
	ContextTimeout time.Duration
}

func NewDownloadCloudRepositoryUseCase(repo _interface.IDownloadCloudRepositoryRepository, statsRepo _interface.IUserStatsCloudRepositoryRepository, fileShareRepo _interface.IFileShareRepository, timeout time.Duration) _interface.IDownloadCloudRepositoryUseCase {
	return &DownloadCloudRepositoryUseCase{
		Repo:           repo,
		StatsRepo:      statsRepo,
		FileShareRepo:  fileShareRepo,
		ContextTimeout: timeout,
	}
}

// RequestDownloadURL generates a presigned download URL for a file
func (u *DownloadCloudRepositoryUseCase) RequestDownloadURL(ctx context.Context, userID, fileID uint) (*response.DownloadResponseDTO, error) {
	// Check if user has access to the file (owner, shared, or folder shared)
	hasAccess, err := u.FileShareRepo.HasFileAccess(ctx, int32(userID), fileID)
	if err != nil {
		return nil, fmt.Errorf("failed to check file access: %w", err)
	}
	if !hasAccess {
		return nil, fmt.Errorf("file not found or no access")
	}

	// Get file from database
	file, err := u.Repo.GetFileByID(ctx, fileID)
	if err != nil {
		return nil, fmt.Errorf("file not found: %w", err)
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
