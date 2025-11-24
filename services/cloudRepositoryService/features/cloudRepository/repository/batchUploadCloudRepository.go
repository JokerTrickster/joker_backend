package repository

import (
	"context"
	"time"

	"github.com/JokerTrickster/joker_backend/services/cloudRepositoryService/features/cloudRepository/model/entity"
	_interface "github.com/JokerTrickster/joker_backend/services/cloudRepositoryService/features/cloudRepository/model/interface"
	sharedAws "github.com/JokerTrickster/joker_backend/shared/aws"
	"gorm.io/gorm"
)

type BatchUploadCloudRepositoryRepository struct {
	db     *gorm.DB
	bucket string
}

func NewBatchUploadCloudRepositoryRepository(db *gorm.DB, bucket string) _interface.IBatchUploadCloudRepositoryRepository {
	return &BatchUploadCloudRepositoryRepository{
		db:     db,
		bucket: bucket,
	}
}

// GeneratePresignedUploadURL generates a presigned URL for uploading
func (r *BatchUploadCloudRepositoryRepository) GeneratePresignedUploadURL(ctx context.Context, s3Key, contentType string, expiration time.Duration) (string, error) {
	return sharedAws.GeneratePresignedUploadURL(ctx, r.bucket, s3Key, contentType, expiration)
}

// CreateFile saves file metadata to database
func (r *BatchUploadCloudRepositoryRepository) CreateFile(ctx context.Context, file *entity.CloudFile) error {
	return r.db.WithContext(ctx).Create(file).Error
}
