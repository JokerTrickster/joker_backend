package usecase

import (
	"context"
	"fmt"
	"time"

	_interface "github.com/JokerTrickster/joker_backend/services/cloudRepositoryService/features/cloudRepository/model/interface"
)

type DeleteCloudRepositoryUseCase struct {
	Repo           _interface.IDeleteCloudRepositoryRepository
	ContextTimeout time.Duration
}

func NewDeleteCloudRepositoryUseCase(repo _interface.IDeleteCloudRepositoryRepository, timeout time.Duration) _interface.IDeleteCloudRepositoryUseCase {
	return &DeleteCloudRepositoryUseCase{
		Repo:           repo,
		ContextTimeout: timeout,
	}
}

// DeleteFile soft deletes a file and removes it from S3
func (u *DeleteCloudRepositoryUseCase) DeleteFile(c context.Context, userID, fileID uint) error {
	ctx, cancel := context.WithTimeout(c, u.ContextTimeout)
	defer cancel()
	// Get file from database
	file, err := u.Repo.GetFileByID(ctx, fileID)
	if err != nil {
		return fmt.Errorf("file not found: %w", err)
	}

	// Check if user owns the file
	if file.UserID != userID {
		return fmt.Errorf("unauthorized access to file")
	}

	// Soft delete from database
	if err := u.Repo.SoftDeleteFile(ctx, fileID, userID); err != nil {
		return fmt.Errorf("failed to delete file: %w", err)
	}

	// Delete from S3 (optional: can be done asynchronously)
	if err := u.Repo.DeleteFromS3(ctx, file.S3Key); err != nil {
		// Log error but don't fail the operation
		// In production, you might want to queue this for retry
		fmt.Printf("Warning: failed to delete file from S3: %v\n", err)
	}

	return nil
}
