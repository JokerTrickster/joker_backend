package repository

import (
	"context"
	"time"

	"github.com/JokerTrickster/joker_backend/services/cloudRepositoryService/features/cloudRepository/model/entity"
	_interface "github.com/JokerTrickster/joker_backend/services/cloudRepositoryService/features/cloudRepository/model/interface"
	sharedAws "github.com/JokerTrickster/joker_backend/shared/aws"
	"gorm.io/gorm"
)

type UploadCloudRepositoryRepository struct {
	db     *gorm.DB
	bucket string
}

func NewUploadCloudRepositoryRepository(db *gorm.DB, bucket string) _interface.IUploadCloudRepositoryRepository {
	return &UploadCloudRepositoryRepository{
		db:     db,
		bucket: bucket,
	}
}

// GeneratePresignedUploadURL generates a presigned URL for uploading
func (r *UploadCloudRepositoryRepository) GeneratePresignedUploadURL(ctx context.Context, s3Key, contentType string, expiration time.Duration) (string, error) {
	return sharedAws.GeneratePresignedUploadURL(ctx, r.bucket, s3Key, contentType, expiration)
}

// CreateFile saves file metadata to database
func (r *UploadCloudRepositoryRepository) CreateFile(ctx context.Context, file *entity.CloudFile) error {
	// Create file record (GORM will handle tag associations since they have IDs)
	if err := r.db.WithContext(ctx).Create(file).Error; err != nil {
		return err
	}

	return nil
}
