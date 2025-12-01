package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/JokerTrickster/joker_backend/services/cloudRepositoryService/features/cloudRepository/model/entity"
	_interface "github.com/JokerTrickster/joker_backend/services/cloudRepositoryService/features/cloudRepository/model/interface"
	"github.com/JokerTrickster/joker_backend/services/cloudRepositoryService/features/cloudRepository/model/response"
	"github.com/JokerTrickster/joker_backend/shared/aws"
	"gorm.io/gorm"
)

type ProcessingStatusUseCase struct {
	DB     *gorm.DB
	Bucket string
}

func NewProcessingStatusUseCase(db *gorm.DB, bucket string) _interface.IProcessingStatusUseCase {
	return &ProcessingStatusUseCase{
		DB:     db,
		Bucket: bucket,
	}
}

// GetProcessingStatus retrieves the processing status of a file
func (u *ProcessingStatusUseCase) GetProcessingStatus(ctx context.Context, userID uint, fileID uint) (*response.ProcessingStatusResponse, error) {
	var file entity.CloudFile

	// Find file and verify ownership
	if err := u.DB.WithContext(ctx).Where("id = ? AND user_id = ?", fileID, userID).First(&file).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("FILE_NOT_FOUND: file with ID %d not found", fileID)
		}
		return nil, fmt.Errorf("failed to fetch file: %w", err)
	}

	// Build response
	resp := &response.ProcessingStatusResponse{
		FileID:    file.ID,
		Status:    string(file.ProcessingStatus),
		Progress:  file.ProcessingProgress,
		CreatedAt: file.CreatedAt,
		UpdatedAt: file.UpdatedAt,
	}

	// Add stage if present
	if file.ProcessingStage != "" {
		stage := string(file.ProcessingStage)
		resp.Stage = &stage
	}

	// Calculate estimated completion time for processing files
	if file.ProcessingStatus == entity.ProcessingStatusProcessing && file.ProcessingStartedAt != nil {
		estimatedTime := u.calculateEstimatedCompletionTime(&file)
		if estimatedTime != nil {
			timeStr := estimatedTime.Format(time.RFC3339)
			resp.EstimatedCompletionTime = &timeStr
		}
	}

	// Add timestamps based on status
	if file.ProcessingStatus == entity.ProcessingStatusCompleted && file.ProcessingCompletedAt != nil {
		resp.CompletedAt = file.ProcessingCompletedAt

		// Add thumbnail URL if available
		if file.ThumbnailKey != "" {
			thumbnailURL, err := aws.GeneratePresignedDownloadURL(ctx, u.Bucket, file.ThumbnailKey, 15*time.Minute)
			if err == nil {
				resp.ThumbnailURL = &thumbnailURL
			}
		}

		// Add duration for videos
		if file.Duration != nil {
			resp.Duration = file.Duration
		}

		// TODO: Add metadata (width, height, codec) from file metadata if stored
		// This would require additional fields in the CloudFile entity or separate metadata table
	}

	if file.ProcessingStatus == entity.ProcessingStatusFailed {
		resp.FailedAt = file.ProcessingCompletedAt

		// Add error information
		if file.ProcessingError != "" {
			resp.Error = &response.ErrorInfo{
				Code:    "PROCESSING_FAILED",
				Message: file.ProcessingError,
			}
		}
	}

	return resp, nil
}

// GetBatchProcessingStatus retrieves processing status for multiple files
func (u *ProcessingStatusUseCase) GetBatchProcessingStatus(ctx context.Context, userID uint, fileIDs []uint) (*response.BatchProcessingStatusResponse, error) {
	var files []entity.CloudFile

	// Find files and verify ownership
	if err := u.DB.WithContext(ctx).
		Where("id IN ? AND user_id = ?", fileIDs, userID).
		Find(&files).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch files: %w", err)
	}

	// Build response for each file
	results := make([]response.ProcessingStatusResponse, 0, len(files))
	for _, file := range files {
		resp := response.ProcessingStatusResponse{
			FileID:    file.ID,
			Status:    string(file.ProcessingStatus),
			Progress:  file.ProcessingProgress,
			CreatedAt: file.CreatedAt,
			UpdatedAt: file.UpdatedAt,
		}

		// Add stage if present
		if file.ProcessingStage != "" {
			stage := string(file.ProcessingStage)
			resp.Stage = &stage
		}

		// Add completed/failed timestamps and additional data
		if file.ProcessingStatus == entity.ProcessingStatusCompleted {
			resp.CompletedAt = file.ProcessingCompletedAt
		} else if file.ProcessingStatus == entity.ProcessingStatusFailed {
			resp.FailedAt = file.ProcessingCompletedAt
			if file.ProcessingError != "" {
				resp.Error = &response.ErrorInfo{
					Code:    "PROCESSING_FAILED",
					Message: file.ProcessingError,
				}
			}
		}

		results = append(results, resp)
	}

	return &response.BatchProcessingStatusResponse{
		Results: results,
	}, nil
}

// calculateEstimatedCompletionTime estimates when processing will complete
// based on file size, current progress, and elapsed time
func (u *ProcessingStatusUseCase) calculateEstimatedCompletionTime(file *entity.CloudFile) *time.Time {
	// Can't estimate if no start time or no progress
	if file.ProcessingStartedAt == nil || file.ProcessingProgress <= 0 {
		return nil
	}

	// Calculate elapsed time
	elapsed := time.Since(*file.ProcessingStartedAt)

	// Calculate average time per percent
	timePerPercent := elapsed / time.Duration(file.ProcessingProgress)

	// Calculate remaining time
	remainingPercent := 100 - file.ProcessingProgress
	remainingTime := timePerPercent * time.Duration(remainingPercent)

	// Add buffer (20%) for variability in processing stages
	remainingTime = time.Duration(float64(remainingTime) * 1.2)

	// Estimated completion time
	estimatedTime := time.Now().Add(remainingTime)
	return &estimatedTime
}
