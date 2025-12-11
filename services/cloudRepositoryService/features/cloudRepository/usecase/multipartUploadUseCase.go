package usecase

import (
	"context"
	"fmt"
	"math"
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
	// DefaultPartSize is 10MB
	DefaultPartSize = 10 * 1024 * 1024
	// MinPartSize is 5MB (S3 minimum except for last part)
	MinPartSize = 5 * 1024 * 1024
	// MaxPartSize is 5GB
	MaxPartSize = 5 * 1024 * 1024 * 1024
	// MaxParts is S3 maximum
	MaxParts = 10000
	// PresignedURLExpiration for multipart uploads
	PresignedURLExpiration = 15 * time.Minute
)

type MultipartUploadUseCase struct {
	Repo           _interface.IMultipartUploadRepository
	StatsRepo      _interface.IUserStatsCloudRepositoryRepository
	DB             *gorm.DB
	QueueClient    *asynq.Client
	Bucket         string
	ContextTimeout time.Duration
}

func NewMultipartUploadUseCase(
	repo _interface.IMultipartUploadRepository,
	statsRepo _interface.IUserStatsCloudRepositoryRepository,
	db *gorm.DB,
	queueClient *asynq.Client,
	bucket string,
	timeout time.Duration,
) _interface.IMultipartUploadUseCase {
	return &MultipartUploadUseCase{
		Repo:           repo,
		StatsRepo:      statsRepo,
		DB:             db,
		QueueClient:    queueClient,
		Bucket:         bucket,
		ContextTimeout: timeout,
	}
}

// InitiateMultipartUpload initiates a multipart upload session
func (u *MultipartUploadUseCase) InitiateMultipartUpload(
	c context.Context,
	userID uint,
	req *request.InitiateMultipartUploadRequestDTO,
) (*response.InitiateMultipartUploadResponseDTO, error) {
	ctx, cancel := context.WithTimeout(c, u.ContextTimeout)
	defer cancel()

	// Validate file type - allow all image/* and video/* MIME types
	fileType := entity.FileType(req.FileType)
	if fileType == entity.FileTypeImage && !strings.HasPrefix(req.ContentType, "image/") {
		return nil, fmt.Errorf("이미지 파일만 업로드할 수 있습니다 (현재: %s)", req.ContentType)
	}
	if fileType == entity.FileTypeVideo && !strings.HasPrefix(req.ContentType, "video/") {
		return nil, fmt.Errorf("동영상 파일만 업로드할 수 있습니다 (현재: %s)", req.ContentType)
	}

	// Generate S3 key
	s3Key := u.generateS3Key(userID, fileType, req.FileName)

	// Calculate part size and total parts
	partSize := DefaultPartSize
	totalParts := int(math.Ceil(float64(req.FileSize) / float64(partSize)))

	// Validate part count
	if totalParts > MaxParts {
		// Recalculate with larger part size
		partSize = int(math.Ceil(float64(req.FileSize) / float64(MaxParts)))
		if partSize > MaxPartSize {
			return nil, fmt.Errorf("file size exceeds maximum allowed for multipart upload")
		}
		totalParts = int(math.Ceil(float64(req.FileSize) / float64(partSize)))
	}

	// Initiate multipart upload in S3
	uploadID, err := u.Repo.CreateMultipartUpload(ctx, u.Bucket, s3Key, req.ContentType)
	if err != nil {
		return nil, fmt.Errorf("failed to initiate multipart upload: %w", err)
	}

	// Create database record
	upload := &entity.MultipartUpload{
		UserID:      userID,
		UploadID:    uploadID,
		FileKey:     s3Key,
		FileName:    req.FileName,
		FileSize:    req.FileSize,
		ContentType: req.ContentType,
		PartSize:    partSize,
		TotalParts:  totalParts,
		Status:      entity.MultipartUploadStatusInitiated,
	}

	if err := u.Repo.CreateMultipartUploadRecord(ctx, upload); err != nil {
		// Try to abort S3 upload if DB record creation fails
		_ = u.Repo.AbortMultipartUpload(ctx, u.Bucket, s3Key, uploadID)
		return nil, fmt.Errorf("failed to create upload record: %w", err)
	}

	logger.Info("Multipart upload initiated",
		zap.Uint("user_id", userID),
		zap.String("upload_id", uploadID),
		zap.String("file_key", s3Key),
		zap.Int("total_parts", totalParts),
	)

	return &response.InitiateMultipartUploadResponseDTO{
		UploadID:   uploadID,
		FileKey:    s3Key,
		PartSize:   partSize,
		TotalParts: totalParts,
	}, nil
}

// GeneratePresignedURLs generates presigned URLs for uploading parts
func (u *MultipartUploadUseCase) GeneratePresignedURLs(
	c context.Context,
	userID uint,
	req *request.GeneratePresignedURLsRequestDTO,
) (*response.GeneratePresignedURLsResponseDTO, error) {
	ctx, cancel := context.WithTimeout(c, u.ContextTimeout)
	defer cancel()

	// Verify upload exists and belongs to user
	upload, err := u.Repo.GetMultipartUpload(ctx, req.UploadID, userID)
	if err != nil {
		return nil, fmt.Errorf("upload not found: %w", err)
	}

	// Verify upload is in correct status
	if upload.Status == entity.MultipartUploadStatusCompleted {
		return nil, fmt.Errorf("upload already completed")
	}
	if upload.Status == entity.MultipartUploadStatusAborted {
		return nil, fmt.Errorf("upload already aborted")
	}

	// Verify file key matches
	if upload.FileKey != req.FileKey {
		return nil, fmt.Errorf("file key mismatch")
	}

	// Generate presigned URLs for each part
	urls := make([]response.PresignedURLPart, 0, len(req.PartNumbers))
	for _, partNumber := range req.PartNumbers {
		// Validate part number
		if partNumber < 1 || partNumber > upload.TotalParts {
			return nil, fmt.Errorf("invalid part number: %d (valid range: 1-%d)", partNumber, upload.TotalParts)
		}

		url, err := u.Repo.GeneratePresignedUploadPartURL(
			ctx,
			u.Bucket,
			upload.FileKey,
			upload.UploadID,
			partNumber,
			PresignedURLExpiration,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to generate URL for part %d: %w", partNumber, err)
		}

		urls = append(urls, response.PresignedURLPart{
			PartNumber: partNumber,
			URL:        url,
		})
	}

	// Update status to uploading if it was initiated
	if upload.Status == entity.MultipartUploadStatusInitiated {
		if err := u.Repo.UpdateMultipartUploadStatus(ctx, upload.UploadID, entity.MultipartUploadStatusUploading); err != nil {
			logger.Warn("Failed to update upload status to uploading",
				zap.String("upload_id", upload.UploadID),
				zap.Error(err),
			)
		}
	}

	logger.Info("Generated presigned URLs for parts",
		zap.String("upload_id", upload.UploadID),
		zap.Int("parts_count", len(urls)),
	)

	return &response.GeneratePresignedURLsResponseDTO{
		URLs:      urls,
		ExpiresIn: int(PresignedURLExpiration.Seconds()),
	}, nil
}

// CompleteMultipartUpload completes a multipart upload
func (u *MultipartUploadUseCase) CompleteMultipartUpload(
	c context.Context,
	userID uint,
	req *request.CompleteMultipartUploadRequestDTO,
) (*response.CompleteMultipartUploadResponseDTO, error) {
	ctx, cancel := context.WithTimeout(c, u.ContextTimeout)
	defer cancel()

	// Verify upload exists and belongs to user
	upload, err := u.Repo.GetMultipartUpload(ctx, req.UploadID, userID)
	if err != nil {
		return nil, fmt.Errorf("upload not found: %w", err)
	}

	// Verify upload is in correct status
	if upload.Status == entity.MultipartUploadStatusCompleted {
		return nil, fmt.Errorf("upload already completed")
	}
	if upload.Status == entity.MultipartUploadStatusAborted {
		return nil, fmt.Errorf("upload already aborted")
	}

	// Verify file key matches
	if upload.FileKey != req.FileKey {
		return nil, fmt.Errorf("file key mismatch")
	}

	// Convert parts to interface type
	parts := make([]_interface.CompletedPart, len(req.Parts))
	for i, part := range req.Parts {
		parts[i] = _interface.CompletedPart{
			PartNumber: part.PartNumber,
			ETag:       part.ETag,
		}
	}

	// Complete multipart upload in S3
	if err := u.Repo.CompleteMultipartUpload(ctx, u.Bucket, upload.FileKey, upload.UploadID, parts); err != nil {
		return nil, fmt.Errorf("failed to complete multipart upload: %w", err)
	}

	// Determine file type from content type
	fileType := entity.FileTypeVideo
	if strings.HasPrefix(upload.ContentType, "image/") {
		fileType = entity.FileTypeImage
	}

	// Generate thumbnail key
	thumbnailKey := ""
	if fileType == entity.FileTypeImage || fileType == entity.FileTypeVideo {
		thumbnailKey = u.generateThumbnailKey(userID, upload.FileName)
	}

	// Create file record in database
	file := &entity.CloudFile{
		UserID:       userID,
		FileName:     upload.FileName,
		S3Key:        upload.FileKey,
		ThumbnailKey: thumbnailKey,
		FileType:     fileType,
		ContentType:  upload.ContentType,
		FileSize:     upload.FileSize,
	}

	if err := u.DB.WithContext(ctx).Create(file).Error; err != nil {
		return nil, fmt.Errorf("failed to create file record: %w", err)
	}

	// Update upload status to completed
	if err := u.Repo.UpdateMultipartUploadStatus(ctx, upload.UploadID, entity.MultipartUploadStatusCompleted); err != nil {
		logger.Warn("Failed to update upload status to completed",
			zap.String("upload_id", upload.UploadID),
			zap.Error(err),
		)
	}

	// Log upload activity
	if u.StatsRepo != nil {
		activity := &entity.ActivityLog{
			UserID:       userID,
			FileID:       &file.ID,
			ActivityType: entity.ActivityTypeUpload,
		}
		_ = u.StatsRepo.LogActivity(ctx, activity)
	}

	// Enqueue video processing for videos
	if fileType == entity.FileTypeVideo && u.QueueClient != nil {
		logger.Info("Enqueueing video processing task for multipart upload",
			zap.Uint("file_id", file.ID),
			zap.String("s3_key", upload.FileKey),
		)

		payload := &queue.VideoProcessingPayload{
			FileID:   file.ID,
			S3Key:    upload.FileKey,
			UserID:   userID,
			Bucket:   u.Bucket,
			FileName: upload.FileName,
		}

		if err := queue.EnqueueVideoProcessing(u.QueueClient, payload); err != nil {
			logger.Error("Failed to enqueue video processing task",
				zap.Uint("file_id", file.ID),
				zap.Error(err),
			)
		} else {
			file.ProcessingStatus = entity.ProcessingStatusPending
			logger.Info("Video processing task enqueued successfully",
				zap.Uint("file_id", file.ID),
			)
		}
	}

	// Generate download URL
	downloadURL := fmt.Sprintf("https://%s.s3.amazonaws.com/%s", u.Bucket, upload.FileKey)

	logger.Info("Multipart upload completed",
		zap.Uint("user_id", userID),
		zap.String("upload_id", upload.UploadID),
		zap.Uint("file_id", file.ID),
	)

	return &response.CompleteMultipartUploadResponseDTO{
		FileID: file.ID,
		URL:    downloadURL,
		Size:   upload.FileSize,
	}, nil
}

// AbortMultipartUpload aborts a multipart upload
func (u *MultipartUploadUseCase) AbortMultipartUpload(
	c context.Context,
	userID uint,
	req *request.AbortMultipartUploadRequestDTO,
) (*response.AbortMultipartUploadResponseDTO, error) {
	ctx, cancel := context.WithTimeout(c, u.ContextTimeout)
	defer cancel()

	// Verify upload exists and belongs to user
	upload, err := u.Repo.GetMultipartUpload(ctx, req.UploadID, userID)
	if err != nil {
		return nil, fmt.Errorf("upload not found: %w", err)
	}

	// Verify file key matches
	if upload.FileKey != req.FileKey {
		return nil, fmt.Errorf("file key mismatch")
	}

	// Don't abort if already completed
	if upload.Status == entity.MultipartUploadStatusCompleted {
		return nil, fmt.Errorf("cannot abort completed upload")
	}

	// Abort multipart upload in S3 (only if not already aborted)
	if upload.Status != entity.MultipartUploadStatusAborted {
		if err := u.Repo.AbortMultipartUpload(ctx, u.Bucket, upload.FileKey, upload.UploadID); err != nil {
			return nil, fmt.Errorf("failed to abort multipart upload: %w", err)
		}
	}

	// Update upload status to aborted
	if err := u.Repo.UpdateMultipartUploadStatus(ctx, upload.UploadID, entity.MultipartUploadStatusAborted); err != nil {
		logger.Warn("Failed to update upload status to aborted",
			zap.String("upload_id", upload.UploadID),
			zap.Error(err),
		)
	}

	logger.Info("Multipart upload aborted",
		zap.Uint("user_id", userID),
		zap.String("upload_id", upload.UploadID),
	)

	return &response.AbortMultipartUploadResponseDTO{
		Success: true,
		Message: "Upload aborted successfully",
	}, nil
}

// generateS3Key generates a unique S3 key for a file
func (u *MultipartUploadUseCase) generateS3Key(userID uint, fileType entity.FileType, fileName string) string {
	fileID := uuid.New().String()
	ext := filepath.Ext(fileName)
	if ext == "" {
		ext = ".bin"
	}
	ext = strings.ToLower(ext)
	baseName := strings.TrimSuffix(fileName, ext)
	if baseName == "" {
		baseName = "file"
	}
	return fmt.Sprintf("users/%d/files/%s-%s%s", userID, fileID, baseName, ext)
}

// generateThumbnailKey generates a unique S3 key for a thumbnail
func (u *MultipartUploadUseCase) generateThumbnailKey(userID uint, fileName string) string {
	fileID := uuid.New().String()
	ext := filepath.Ext(fileName)
	if ext == "" {
		ext = ".jpg"
	}
	ext = strings.ToLower(ext)
	baseName := strings.TrimSuffix(fileName, ext)
	if baseName == "" {
		baseName = "file"
	}
	return fmt.Sprintf("users/%d/thumbnails/%s-%s_thumb%s", userID, fileID, baseName, ext)
}
