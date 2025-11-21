package repository

import (
	"context"
	"time"

	"github.com/JokerTrickster/joker_backend/services/cloudRepositoryService/features/cloudRepository/model"
	sharedAws "github.com/JokerTrickster/joker_backend/shared/aws"
)

// GeneratePresignedUploadURL generates a presigned URL for uploading
func (r *CloudRepositoryRepository) GeneratePresignedUploadURL(ctx context.Context, s3Key, contentType string, expiration time.Duration) (string, error) {
	return sharedAws.GeneratePresignedUploadURL(ctx, r.bucket, s3Key, contentType, expiration)
}

// CreateFile saves file metadata to database
func (r *CloudRepositoryRepository) CreateFile(ctx context.Context, file *model.CloudFile) error {
	return r.db.WithContext(ctx).Create(file).Error
}
