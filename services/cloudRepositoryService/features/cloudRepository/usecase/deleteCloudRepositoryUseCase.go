package usecase

import (
	"context"
	"fmt"
)

// DeleteFile soft deletes a file and removes it from S3
func (u *CloudRepositoryUsecase) DeleteFile(ctx context.Context, userID, fileID uint) error {
	// Get file from database
	file, err := u.repo.GetFileByID(ctx, fileID)
	if err != nil {
		return fmt.Errorf("file not found: %w", err)
	}

	// Check if user owns the file
	if file.UserID != userID {
		return fmt.Errorf("unauthorized access to file")
	}

	// Soft delete from database
	if err := u.repo.SoftDeleteFile(ctx, fileID, userID); err != nil {
		return fmt.Errorf("failed to delete file: %w", err)
	}

	// Delete from S3 (optional: can be done asynchronously)
	if err := u.repo.DeleteFromS3(ctx, file.S3Key); err != nil {
		// Log error but don't fail the operation
		// In production, you might want to queue this for retry
		fmt.Printf("Warning: failed to delete file from S3: %v\n", err)
	}

	return nil
}
