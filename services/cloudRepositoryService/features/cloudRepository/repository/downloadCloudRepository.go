package repository

import (
	"context"
	"time"

	"github.com/JokerTrickster/joker_backend/services/cloudRepositoryService/features/cloudRepository/model/entity"
	_interface "github.com/JokerTrickster/joker_backend/services/cloudRepositoryService/features/cloudRepository/model/interface"
	sharedAws "github.com/JokerTrickster/joker_backend/shared/aws"
	"gorm.io/gorm"
)

type DownloadCloudRepositoryRepository struct {
	db     *gorm.DB
	bucket string
}

func NewDownloadCloudRepositoryRepository(db *gorm.DB, bucket string) _interface.IDownloadCloudRepositoryRepository {
	return &DownloadCloudRepositoryRepository{
		db:     db,
		bucket: bucket,
	}
}

// GeneratePresignedDownloadURL generates a presigned URL for downloading
func (r *DownloadCloudRepositoryRepository) GeneratePresignedDownloadURL(ctx context.Context, s3Key string, expiration time.Duration) (string, error) {
	return sharedAws.GeneratePresignedDownloadURL(ctx, r.bucket, s3Key, expiration)
}

// GeneratePresignedDownloadURLWithFilename generates a presigned URL for downloading with Content-Disposition header
func (r *DownloadCloudRepositoryRepository) GeneratePresignedDownloadURLWithFilename(ctx context.Context, s3Key, filename string, expiration time.Duration) (string, error) {
	return sharedAws.GeneratePresignedDownloadURLWithFilename(ctx, r.bucket, s3Key, filename, expiration)
}

// GetFileByID retrieves a file by ID
func (r *DownloadCloudRepositoryRepository) GetFileByID(ctx context.Context, id uint) (*entity.CloudFile, error) {
	var file entity.CloudFile
	err := r.db.WithContext(ctx).Where("id = ? AND deleted_at IS NULL", id).First(&file).Error
	if err != nil {
		return nil, err
	}
	return &file, nil
}
