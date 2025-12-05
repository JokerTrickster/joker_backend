package repository

import (
	"context"
	"time"

	"github.com/JokerTrickster/joker_backend/services/cloudRepositoryService/features/cloudRepository/model/entity"
	_interface "github.com/JokerTrickster/joker_backend/services/cloudRepositoryService/features/cloudRepository/model/interface"
	sharedAws "github.com/JokerTrickster/joker_backend/shared/aws"
	"gorm.io/gorm"
)

type MultipartUploadRepository struct {
	db     *gorm.DB
	bucket string
}

func NewMultipartUploadRepository(db *gorm.DB, bucket string) _interface.IMultipartUploadRepository {
	return &MultipartUploadRepository{
		db:     db,
		bucket: bucket,
	}
}

// CreateMultipartUpload initiates a multipart upload in S3
func (r *MultipartUploadRepository) CreateMultipartUpload(ctx context.Context, bucket, key, contentType string) (string, error) {
	return sharedAws.CreateMultipartUpload(ctx, bucket, key, contentType)
}

// GeneratePresignedUploadPartURL generates a presigned URL for uploading a part
func (r *MultipartUploadRepository) GeneratePresignedUploadPartURL(ctx context.Context, bucket, key, uploadID string, partNumber int, expiration time.Duration) (string, error) {
	return sharedAws.GeneratePresignedUploadPartURL(ctx, bucket, key, uploadID, partNumber, expiration)
}

// CompleteMultipartUpload completes a multipart upload in S3
func (r *MultipartUploadRepository) CompleteMultipartUpload(ctx context.Context, bucket, key, uploadID string, parts []_interface.CompletedPart) error {
	// Convert to AWS SDK types
	awsParts := make([]sharedAws.CompletedPart, len(parts))
	for i, part := range parts {
		awsParts[i] = sharedAws.CompletedPart{
			PartNumber: part.PartNumber,
			ETag:       part.ETag,
		}
	}
	return sharedAws.CompleteMultipartUpload(ctx, bucket, key, uploadID, awsParts)
}

// AbortMultipartUpload aborts a multipart upload in S3
func (r *MultipartUploadRepository) AbortMultipartUpload(ctx context.Context, bucket, key, uploadID string) error {
	return sharedAws.AbortMultipartUpload(ctx, bucket, key, uploadID)
}

// CreateMultipartUploadRecord creates a record in the database
func (r *MultipartUploadRepository) CreateMultipartUploadRecord(ctx context.Context, upload *entity.MultipartUpload) error {
	return r.db.WithContext(ctx).Create(upload).Error
}

// UpdateMultipartUploadStatus updates the status of a multipart upload
func (r *MultipartUploadRepository) UpdateMultipartUploadStatus(ctx context.Context, uploadID string, status entity.MultipartUploadStatus) error {
	return r.db.WithContext(ctx).Model(&entity.MultipartUpload{}).
		Where("upload_id = ?", uploadID).
		Update("status", status).Error
}

// GetMultipartUpload retrieves a multipart upload record
func (r *MultipartUploadRepository) GetMultipartUpload(ctx context.Context, uploadID string, userID uint) (*entity.MultipartUpload, error) {
	var upload entity.MultipartUpload
	err := r.db.WithContext(ctx).
		Where("upload_id = ? AND user_id = ?", uploadID, userID).
		First(&upload).Error
	if err != nil {
		return nil, err
	}
	return &upload, nil
}
