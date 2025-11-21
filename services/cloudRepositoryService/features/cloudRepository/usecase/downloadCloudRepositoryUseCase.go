package usecase

import (
	"context"
	"fmt"

	"github.com/JokerTrickster/joker_backend/services/cloudRepositoryService/features/cloudRepository/model"
)

// RequestDownloadURL generates a presigned download URL for a file
func (u *CloudRepositoryUsecase) RequestDownloadURL(ctx context.Context, userID, fileID uint) (*model.DownloadResponseDTO, error) {
	// Get file from database
	file, err := u.repo.GetFileByID(ctx, fileID)
	if err != nil {
		return nil, fmt.Errorf("file not found: %w", err)
	}

	// Check if user owns the file
	if file.UserID != userID {
		return nil, fmt.Errorf("unauthorized access to file")
	}

	// Generate presigned download URL
	downloadURL, err := u.repo.GeneratePresignedDownloadURL(ctx, file.S3Key, DefaultDownloadExpiration)
	if err != nil {
		return nil, fmt.Errorf("failed to generate download URL: %w", err)
	}

	return &model.DownloadResponseDTO{
		DownloadURL: downloadURL,
		FileName:    file.FileName,
		ExpiresIn:   int(DefaultDownloadExpiration.Seconds()),
	}, nil
}
