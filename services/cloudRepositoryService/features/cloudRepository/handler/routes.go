package handler

import (
	"github.com/JokerTrickster/joker_backend/services/cloudRepositoryService/features/cloudRepository/repository"
	"github.com/JokerTrickster/joker_backend/services/cloudRepositoryService/features/cloudRepository/usecase"
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

// RegisterRoutes registers all cloud repository routes
func RegisterRoutes(e *echo.Group, db *gorm.DB, bucket string) {
	// Initialize layers
	repo := repository.NewCloudRepositoryRepository(db, bucket)
	uc := usecase.NewCloudRepositoryUsecase(repo)
	handler := NewCloudRepositoryHandler(uc)

	// Register routes
	files := e.Group("/files")
	{
		files.POST("/upload", handler.RequestUploadURL)           // Request presigned upload URL (single file)
		files.POST("/upload/batch", handler.RequestBatchUploadURL) // Request presigned upload URLs (batch, max 30)
		files.GET("", handler.ListFiles)                          // List files
		files.GET("/:id/download", handler.RequestDownloadURL)    // Request presigned download URL
		files.DELETE("/:id", handler.DeleteFile)                  // Delete file
	}
}
