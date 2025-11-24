package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/JokerTrickster/joker_backend/services/cloudRepositoryService/features/cloudRepository/model/entity"
	_interface "github.com/JokerTrickster/joker_backend/services/cloudRepositoryService/features/cloudRepository/model/interface"
	sharedAws "github.com/JokerTrickster/joker_backend/shared/aws"
	"gorm.io/gorm"
)

type DeleteCloudRepositoryRepository struct {
	db     *gorm.DB
	bucket string
}

func NewDeleteCloudRepositoryRepository(db *gorm.DB, bucket string) _interface.IDeleteCloudRepositoryRepository {
	return &DeleteCloudRepositoryRepository{
		db:     db,
		bucket: bucket,
	}
}

// DeleteFromS3 deletes a file from S3
func (r *DeleteCloudRepositoryRepository) DeleteFromS3(ctx context.Context, s3Key string) error {
	return sharedAws.DeleteObject(ctx, r.bucket, s3Key)
}

// SoftDeleteFile soft deletes a file
func (r *DeleteCloudRepositoryRepository) SoftDeleteFile(ctx context.Context, id uint, userID uint) error {
	now := time.Now()
	result := r.db.WithContext(ctx).Model(&entity.CloudFile{}).
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

// GetFileByID retrieves a file by ID
func (r *DeleteCloudRepositoryRepository) GetFileByID(ctx context.Context, id uint) (*entity.CloudFile, error) {
	var file entity.CloudFile
	err := r.db.WithContext(ctx).Where("id = ? AND deleted_at IS NULL", id).First(&file).Error
	if err != nil {
		return nil, err
	}
	return &file, nil
}

// HardDeleteFile permanently deletes a file from database
func (r *DeleteCloudRepositoryRepository) HardDeleteFile(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Unscoped().Delete(&entity.CloudFile{}, id).Error
}
