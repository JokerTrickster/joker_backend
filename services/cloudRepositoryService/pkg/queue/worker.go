package queue

import (
	"context"
	"fmt"

	"github.com/JokerTrickster/joker_backend/services/cloudRepositoryService/features/cloudRepository/model/entity"
	"github.com/JokerTrickster/joker_backend/services/cloudRepositoryService/pkg/ffmpeg"
	"github.com/JokerTrickster/joker_backend/shared/logger"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/hibiken/asynq"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// VideoProcessor handles video processing tasks
type VideoProcessor struct {
	db         *gorm.DB
	ffmpeg     *ffmpeg.Processor
	wsNotifier WSNotifier
}

// WSNotifier is an interface for WebSocket notifications
type WSNotifier interface {
	NotifyFileProcessed(userID uint, fileID uint, status string, thumbnailKey string, duration float64, err error)
}

// NewVideoProcessor creates a new video processor
func NewVideoProcessor(db *gorm.DB, s3Client *s3.Client, bucket string, wsNotifier WSNotifier) *VideoProcessor {
	return &VideoProcessor{
		db:         db,
		ffmpeg:     ffmpeg.NewProcessor(s3Client, bucket),
		wsNotifier: wsNotifier,
	}
}

// ProcessTask processes a video processing task
func (p *VideoProcessor) ProcessTask(ctx context.Context, task *asynq.Task) error {
	// Parse payload
	payload, err := ParseVideoProcessingPayload(task)
	if err != nil {
		return fmt.Errorf("failed to parse payload: %w", err)
	}

	logger.Info("Processing video",
		zap.Uint("file_id", payload.FileID),
		zap.String("s3_key", payload.S3Key),
		zap.Uint("user_id", payload.UserID),
	)

	// Update status to processing with initial stage
	if err := p.updateFileProgress(payload.FileID, entity.ProcessingStatusProcessing, entity.ProcessingStageValidating, 0, ""); err != nil {
		return fmt.Errorf("failed to update status to processing: %w", err)
	}

	// Extract duration and metadata
	logger.Info("Extracting video duration", zap.Uint("file_id", payload.FileID))
	if err := p.updateFileProgress(payload.FileID, entity.ProcessingStatusProcessing, entity.ProcessingStageExtractingMetadata, 20, ""); err != nil {
		logger.Error("Failed to update progress", zap.Error(err))
	}

	duration, err := p.ffmpeg.ExtractDuration(ctx, payload.S3Key)
	if err != nil {
		p.handleProcessingError(payload, fmt.Errorf("failed to extract duration: %w", err))
		return err
	}
	logger.Info("Duration extracted", zap.Uint("file_id", payload.FileID), zap.Float64("duration", duration))

	// Generate thumbnail at middle of video
	logger.Info("Generating thumbnail", zap.Uint("file_id", payload.FileID))
	if err := p.updateFileProgress(payload.FileID, entity.ProcessingStatusProcessing, entity.ProcessingStageGeneratingThumbnail, 50, ""); err != nil {
		logger.Error("Failed to update progress", zap.Error(err))
	}

	seekTime := duration / 2
	if duration <= 0 {
		seekTime = 0
	}

	thumbnailData, err := p.ffmpeg.GenerateThumbnail(ctx, payload.S3Key, seekTime)
	if err != nil {
		p.handleProcessingError(payload, fmt.Errorf("failed to generate thumbnail: %w", err))
		return err
	}

	// Upload thumbnail to S3
	thumbnailKey := fmt.Sprintf("thumbnails/%d_%s.jpg", payload.FileID, payload.FileName)
	logger.Info("Uploading thumbnail", zap.Uint("file_id", payload.FileID), zap.String("thumbnail_key", thumbnailKey))
	if err := p.updateFileProgress(payload.FileID, entity.ProcessingStatusProcessing, entity.ProcessingStageUploadingThumbnail, 75, ""); err != nil {
		logger.Error("Failed to update progress", zap.Error(err))
	}

	if err := p.ffmpeg.UploadThumbnail(ctx, thumbnailKey, thumbnailData); err != nil {
		p.handleProcessingError(payload, fmt.Errorf("failed to upload thumbnail: %w", err))
		return err
	}

	// Finalize processing
	if err := p.updateFileProgress(payload.FileID, entity.ProcessingStatusProcessing, entity.ProcessingStageFinalizing, 90, ""); err != nil {
		logger.Error("Failed to update progress", zap.Error(err))
	}

	// Update database with results
	if err := p.updateFileWithResults(payload.FileID, duration, thumbnailKey); err != nil {
		p.handleProcessingError(payload, fmt.Errorf("failed to update file with results: %w", err))
		return err
	}

	// Send WebSocket notification
	if p.wsNotifier != nil {
		p.wsNotifier.NotifyFileProcessed(payload.UserID, payload.FileID, "completed", thumbnailKey, duration, nil)
	}

	logger.Info("Video processing completed successfully", zap.Uint("file_id", payload.FileID))
	return nil
}

// updateFileProgress updates the processing status, stage, and progress of a file
func (p *VideoProcessor) updateFileProgress(fileID uint, status entity.ProcessingStatus, stage entity.ProcessingStage, progress int, errorMsg string) error {
	updates := map[string]interface{}{
		"processing_status":   status,
		"processing_stage":    stage,
		"processing_progress": progress,
	}

	// Set processing_started_at when first entering processing status
	if status == entity.ProcessingStatusProcessing && stage == entity.ProcessingStageValidating {
		updates["processing_started_at"] = time.Now()
	}

	if errorMsg != "" {
		updates["processing_error"] = errorMsg
	}

	return p.db.Model(&entity.CloudFile{}).
		Where("id = ?", fileID).
		Updates(updates).Error
}

// updateFileWithResults updates file with processing results
func (p *VideoProcessor) updateFileWithResults(fileID uint, duration float64, thumbnailKey string) error {
	updates := map[string]interface{}{
		"duration":                duration,
		"thumbnail_key":           thumbnailKey,
		"processing_status":       entity.ProcessingStatusCompleted,
		"processing_stage":        entity.ProcessingStageDone,
		"processing_progress":     100,
		"processing_error":        "", // Clear any previous errors
		"processing_completed_at": time.Now(),
	}

	return p.db.Model(&entity.CloudFile{}).
		Where("id = ?", fileID).
		Updates(updates).Error
}

// handleProcessingError handles errors during processing
func (p *VideoProcessor) handleProcessingError(payload *VideoProcessingPayload, err error) {
	logger.Error("Video processing failed",
		zap.Uint("file_id", payload.FileID),
		zap.Error(err),
	)

	// Update database with error and failed timestamp
	updates := map[string]interface{}{
		"processing_status":       entity.ProcessingStatusFailed,
		"processing_error":        err.Error(),
		"processing_completed_at": time.Now(),
	}

	if dbErr := p.db.Model(&entity.CloudFile{}).Where("id = ?", payload.FileID).Updates(updates).Error; dbErr != nil {
		logger.Error("Failed to update error status",
			zap.Uint("file_id", payload.FileID),
			zap.Error(dbErr),
		)
	}

	// Send WebSocket notification
	if p.wsNotifier != nil {
		p.wsNotifier.NotifyFileProcessed(payload.UserID, payload.FileID, "failed", "", 0, err)
	}
}
