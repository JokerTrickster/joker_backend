package usecase

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/JokerTrickster/joker_backend/services/cloudRepositoryService/features/cloudRepository/model/entity"
	_interface "github.com/JokerTrickster/joker_backend/services/cloudRepositoryService/features/cloudRepository/model/interface"
	"github.com/JokerTrickster/joker_backend/services/cloudRepositoryService/features/cloudRepository/model/request"
	"github.com/JokerTrickster/joker_backend/services/cloudRepositoryService/features/cloudRepository/model/response"
	"github.com/JokerTrickster/joker_backend/services/cloudRepositoryService/pkg/queue"
	"github.com/JokerTrickster/joker_backend/shared/logger"
	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

const (
	DefaultUploadExpiration = 12 * time.Hour
)

type UploadCloudRepositoryUseCase struct {
	Repo           _interface.IUploadCloudRepositoryRepository
	StatsRepo      _interface.IUserStatsCloudRepositoryRepository
	DB             *gorm.DB
	QueueClient    *asynq.Client
	Bucket         string
	ContextTimeout time.Duration
}

func NewUploadCloudRepositoryUseCase(repo _interface.IUploadCloudRepositoryRepository, statsRepo _interface.IUserStatsCloudRepositoryRepository, db *gorm.DB, queueClient *asynq.Client, bucket string, timeout time.Duration) _interface.IUploadCloudRepositoryUseCase {
	return &UploadCloudRepositoryUseCase{
		Repo:           repo,
		StatsRepo:      statsRepo,
		DB:             db,
		QueueClient:    queueClient,
		Bucket:         bucket,
		ContextTimeout: timeout,
	}
}

// RequestUploadURL generates a presigned upload URL and creates a file record
func (u *UploadCloudRepositoryUseCase) RequestUploadURL(c context.Context, userID uint, req *request.UploadRequestDTO) (*response.UploadResponseDTO, error) {
	ctx, cancel := context.WithTimeout(c, u.ContextTimeout)
	defer cancel()

	// Validate content type - allow all image/* and video/* MIME types
	fileType := entity.FileType(req.FileType)
	if fileType == entity.FileTypeImage && !strings.HasPrefix(req.ContentType, "image/") {
		return nil, fmt.Errorf("이미지 파일만 업로드할 수 있습니다 (현재: %s)", req.ContentType)
	}
	if fileType == entity.FileTypeVideo && !strings.HasPrefix(req.ContentType, "video/") {
		return nil, fmt.Errorf("동영상 파일만 업로드할 수 있습니다 (현재: %s)", req.ContentType)
	}

	// Check folder write permission if folder_id is provided
	if req.FolderID != nil {
		// Check if user is the folder owner
		var folder entity.Folder
		err := u.DB.WithContext(ctx).
			Where("id = ? AND user_id = ? AND deleted_at IS NULL", *req.FolderID, int32(userID)).
			First(&folder).Error

		if err == nil {
			// User is owner, has write permission
		} else if err == gorm.ErrRecordNotFound {
			// User is not owner, check if they have write permission via share
			var share entity.FolderShare
			err = u.DB.WithContext(ctx).
				Where("folder_id = ? AND shared_with_id = ? AND permission = ? AND deleted_at IS NULL",
					*req.FolderID, int32(userID), entity.SharePermissionWrite).
				First(&share).Error

			if err != nil {
				if err == gorm.ErrRecordNotFound {
					return nil, fmt.Errorf("업로드 권한이 없습니다")
				}
				return nil, fmt.Errorf("failed to check folder write permission: %w", err)
			}
		} else {
			return nil, fmt.Errorf("failed to check folder ownership: %w", err)
		}
	}

	// Generate S3 keys for original and thumbnail
	s3Key := u.generateS3Key(userID, fileType, req.FileName)
	thumbnailKey := ""

	// Generate thumbnail key for both images and videos
	if fileType == entity.FileTypeImage || fileType == entity.FileTypeVideo {
		thumbnailKey = u.generateThumbnailKey(userID, req.FileName)
	}

	// Process tags if provided
	tags := make([]entity.Tag, 0, len(req.Tags))
	if len(req.Tags) > 0 {
		for _, tagName := range req.Tags {
			if tagName == "" {
				continue
			}
			tag := entity.Tag{
				UserID: userID,
				Name:   tagName,
			}
			// Find or create tag
			if err := u.DB.WithContext(ctx).Where("user_id = ? AND name = ?", userID, tagName).FirstOrCreate(&tag).Error; err != nil {
				return nil, fmt.Errorf("failed to process tag %s: %w", tagName, err)
			}
			tags = append(tags, tag)
		}

		// Log tag activity
		if u.StatsRepo != nil {
			for _, tag := range tags {
				activity := &entity.ActivityLog{
					UserID:       userID,
					ActivityType: entity.ActivityTypeTagAdd,
					TagName:      tag.Name,
				}
				_ = u.StatsRepo.LogActivity(ctx, activity) // Don't fail on logging error
			}
		}
	}

	// Create file record in database
	file := &entity.CloudFile{
		UserID:       userID,
		FolderID:     req.FolderID, // Set folder_id if provided
		FileName:     req.FileName,
		S3Key:        s3Key,
		ThumbnailKey: thumbnailKey,
		FileType:     fileType,
		ContentType:  req.ContentType,
		FileSize:     req.FileSize,
		Duration:     req.Duration,
		Tags:         tags,
	}

	if err := u.Repo.CreateFile(ctx, file); err != nil {
		return nil, fmt.Errorf("failed to create file record: %w", err)
	}

	// Log upload activity
	if u.StatsRepo != nil {
		activity := &entity.ActivityLog{
			UserID:       userID,
			FileID:       &file.ID,
			ActivityType: entity.ActivityTypeUpload,
		}
		_ = u.StatsRepo.LogActivity(ctx, activity) // Don't fail on logging error
	}

	// Generate presigned upload URL for original
	uploadURL, err := u.Repo.GeneratePresignedUploadURL(ctx, s3Key, req.ContentType, DefaultUploadExpiration)
	if err != nil {
		return nil, fmt.Errorf("failed to generate upload URL: %w", err)
	}

	// Generate presigned upload URL for thumbnail if it's an image
	thumbnailURL := ""
	if thumbnailKey != "" {
		thumbnailURL, err = u.Repo.GeneratePresignedUploadURL(ctx, thumbnailKey, req.ContentType, DefaultUploadExpiration)
		if err != nil {
			// Log error but don't fail the entire request
			thumbnailURL = ""
		}
	}

	// Automatically enqueue video processing for videos
	if fileType == entity.FileTypeVideo && u.QueueClient != nil {
		logger.Info("Auto-enqueueing video processing task",
			zap.Uint("file_id", file.ID),
			zap.String("s3_key", s3Key),
		)

		payload := &queue.VideoProcessingPayload{
			FileID:   file.ID,
			S3Key:    s3Key,
			UserID:   userID,
			Bucket:   u.Bucket,
			FileName: req.FileName,
		}

		// Enqueue with a delay to allow S3 upload to complete
		if err := queue.EnqueueVideoProcessing(u.QueueClient, payload); err != nil {
			logger.Error("Failed to auto-enqueue video processing task",
				zap.Uint("file_id", file.ID),
				zap.Error(err),
			)
			// Don't fail the request if queueing fails
		} else {
			logger.Info("Video processing task auto-enqueued successfully",
				zap.Uint("file_id", file.ID),
			)
		}
	}

	return &response.UploadResponseDTO{
		FileID:       file.ID,
		UploadURL:    uploadURL,
		S3Key:        s3Key,
		ThumbnailURL: thumbnailURL,
		ThumbnailKey: thumbnailKey,
		ExpiresIn:    int(DefaultUploadExpiration.Seconds()),
	}, nil
}

// generateS3Key generates a unique S3 key for a file
func (u *UploadCloudRepositoryUseCase) generateS3Key(userID uint, fileType entity.FileType, fileName string) string {
	// Generate UUID for uniqueness
	fileID := uuid.New().String()

	// Get file extension
	ext := filepath.Ext(fileName)
	if ext == "" {
		ext = ".bin"
	}

	// Clean extension
	ext = strings.ToLower(ext)

	// Get filename without extension
	baseName := strings.TrimSuffix(fileName, ext)
	if baseName == "" {
		baseName = "file"
	}

	// Format: users/{userID}/files/{uuid}-{baseName}{ext}
	return fmt.Sprintf("users/%d/files/%s-%s%s", userID, fileID, baseName, ext)
}

// generateThumbnailKey generates a unique S3 key for a thumbnail
func (u *UploadCloudRepositoryUseCase) generateThumbnailKey(userID uint, fileName string) string {
	// Generate UUID for uniqueness
	fileID := uuid.New().String()

	// Get file extension
	ext := filepath.Ext(fileName)
	if ext == "" {
		ext = ".jpg" // Default to jpg for thumbnails
	}

	// Clean extension
	ext = strings.ToLower(ext)

	// Get filename without extension
	baseName := strings.TrimSuffix(fileName, ext)
	if baseName == "" {
		baseName = "file"
	}

	// Format: users/{userID}/thumbnails/{uuid}-{baseName}_thumb{ext}
	return fmt.Sprintf("users/%d/thumbnails/%s-%s_thumb%s", userID, fileID, baseName, ext)
}
