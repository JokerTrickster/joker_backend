package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/JokerTrickster/joker_backend/services/cloudRepositoryService/features/cloudRepository/model"
	sharedAws "github.com/JokerTrickster/joker_backend/shared/aws"
)

// DeleteFromS3 deletes a file from S3
func (r *CloudRepositoryRepository) DeleteFromS3(ctx context.Context, s3Key string) error {
	return sharedAws.DeleteObject(ctx, r.bucket, s3Key)
}

// SoftDeleteFile soft deletes a file
func (r *CloudRepositoryRepository) SoftDeleteFile(ctx context.Context, id uint, userID uint) error {
	now := time.Now()
	result := r.db.WithContext(ctx).Model(&model.CloudFile{}).
		Where("id = ? AND user_id = ? AND deleted_at IS NULL", id, userID).
		Update("deleted_at", now)
	
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("file not found or already deleted")
	}
	return nil
}

// HardDeleteFile permanently deletes a file from database
func (r *CloudRepositoryRepository) HardDeleteFile(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Unscoped().Delete(&model.CloudFile{}, id).Error
}
