package handler

import (
	"net/http"
	"strconv"

	"github.com/JokerTrickster/joker_backend/services/cloudRepositoryService/features/cloudRepository/model/entity"
	"github.com/JokerTrickster/joker_backend/services/cloudRepositoryService/pkg/queue"
	"github.com/JokerTrickster/joker_backend/shared/logger"
	"github.com/hibiken/asynq"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// CompleteUploadHandler handles file upload completion
type CompleteUploadHandler struct {
	db          *gorm.DB
	queueClient *asynq.Client
	bucket      string
}

// NewCompleteUploadHandler creates a new complete upload handler
func NewCompleteUploadHandler(e *echo.Group, db *gorm.DB, queueClient *asynq.Client, bucket string) {
	handler := &CompleteUploadHandler{
		db:          db,
		queueClient: queueClient,
		bucket:      bucket,
	}

	e.POST("/files/:id/complete-upload", handler.CompleteUpload)
}

// CompleteUploadRequest represents the request body
type CompleteUploadRequest struct {
	FileID uint `json:"file_id"`
}

// CompleteUploadResponse represents the response
type CompleteUploadResponse struct {
	Success bool              `json:"success"`
	File    *entity.CloudFile `json:"file"`
	Message string            `json:"message,omitempty"`
}

// CompleteUpload godoc
// @Summary Complete file upload
// @Description Notify backend that file upload is complete and trigger async processing for videos
// @Tags files
// @Accept json
// @Produce json
// @Param id path int true "File ID"
// @Success 200 {object} CompleteUploadResponse
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/files/{id}/complete-upload [post]
func (h *CompleteUploadHandler) CompleteUpload(c echo.Context) error {
	// Get file ID from path parameter
	fileIDStr := c.Param("id")
	fileID, err := strconv.ParseUint(fileIDStr, 10, 32)
	if err != nil {
		logger.Warn("Invalid file ID", zap.String("file_id", fileIDStr), zap.Error(err))
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid file ID",
		})
	}

	// Get userID from context
	userID, ok := c.Get("userID").(uint)
	if !ok {
		logger.Warn("User ID not found in context")
		return c.JSON(http.StatusUnauthorized, map[string]string{
			"error": "Unauthorized",
		})
	}

	// Find the file
	var file entity.CloudFile
	if err := h.db.Where("id = ? AND user_id = ?", uint(fileID), userID).First(&file).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			logger.Warn("File not found",
				zap.Uint("file_id", uint(fileID)),
				zap.Uint("user_id", userID),
			)
			return c.JSON(http.StatusNotFound, map[string]string{
				"error": "File not found",
			})
		}

		logger.Error("Database error", zap.Error(err))
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to fetch file",
		})
	}

	// Check if file is a video
	if file.FileType == entity.FileTypeVideo {
		logger.Info("Enqueueing video processing task",
			zap.Uint("file_id", file.ID),
			zap.String("s3_key", file.S3Key),
		)

		// Enqueue video processing task
		payload := &queue.VideoProcessingPayload{
			FileID:   file.ID,
			S3Key:    file.S3Key,
			UserID:   userID,
			Bucket:   h.bucket,
			FileName: file.FileName,
		}

		if err := queue.EnqueueVideoProcessing(h.queueClient, payload); err != nil {
			logger.Error("Failed to enqueue video processing task",
				zap.Uint("file_id", file.ID),
				zap.Error(err),
			)
			return c.JSON(http.StatusInternalServerError, map[string]string{
				"error": "Failed to start video processing",
			})
		}

		// Update file status to pending
		file.ProcessingStatus = entity.ProcessingStatusPending

		logger.Info("Video processing task enqueued successfully",
			zap.Uint("file_id", file.ID),
		)

		return c.JSON(http.StatusOK, CompleteUploadResponse{
			Success: true,
			File:    &file,
			Message: "Video processing started",
		})
	}

	// For images, mark as completed immediately
	file.ProcessingStatus = entity.ProcessingStatusCompleted

	logger.Info("Image upload completed",
		zap.Uint("file_id", file.ID),
		zap.String("file_type", string(file.FileType)),
	)

	return c.JSON(http.StatusOK, CompleteUploadResponse{
		Success: true,
		File:    &file,
		Message: "Upload completed",
	})
}
