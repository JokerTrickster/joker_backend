package repository

import (
	"context"
	"time"

	"github.com/JokerTrickster/joker_backend/services/cloudRepositoryService/features/cloudRepository/model"
	sharedAws "github.com/JokerTrickster/joker_backend/shared/aws"
)

// GeneratePresignedDownloadURL generates a presigned URL for downloading
func (r *CloudRepositoryRepository) GeneratePresignedDownloadURL(ctx context.Context, s3Key string, expiration time.Duration) (string, error) {
	return sharedAws.GeneratePresignedDownloadURL(ctx, r.bucket, s3Key, expiration)
}

// GetFileByID retrieves a file by ID
func (r *CloudRepositoryRepository) GetFileByID(ctx context.Context, id uint) (*model.CloudFile, error) {
	var file model.CloudFile
	err := r.db.WithContext(ctx).Where("id = ? AND deleted_at IS NULL", id).First(&file).Error
	if err != nil {
		return nil, err
	}
	return &file, nil
}
