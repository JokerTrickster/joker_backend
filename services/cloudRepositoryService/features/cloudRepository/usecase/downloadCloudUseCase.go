package usecase

import (
	"context"
	"fmt"
	"time"

	_interface "github.com/JokerTrickster/joker_backend/services/cloudRepositoryService/features/cloudRepository/model/interface"
	"github.com/JokerTrickster/joker_backend/services/cloudRepositoryService/features/cloudRepository/model/response"
)

type DownloadCloudRepositoryUseCase struct {
	Repo           _interface.IDownloadCloudRepositoryRepository
	ContextTimeout time.Duration
}

func NewDownloadCloudRepositoryUseCase(repo _interface.IDownloadCloudRepositoryRepository, timeout time.Duration) _interface.IDownloadCloudRepositoryUseCase {
	return &DownloadCloudRepositoryUseCase{
		Repo:           repo,
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

	// Generate presigned download URL
	downloadURL, err := u.Repo.GeneratePresignedDownloadURL(ctx, file.S3Key, 1*time.Hour)
	if err != nil {
		return nil, fmt.Errorf("failed to generate download URL: %w", err)
	}

	return &response.DownloadResponseDTO{
		DownloadURL: downloadURL,
		FileName:    file.FileName,
		ExpiresIn:   int(time.Hour.Seconds()),
	}, nil
}
