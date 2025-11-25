package handler

import (
	"time"

	"github.com/JokerTrickster/joker_backend/services/cloudRepositoryService/features/cloudRepository/repository"
	"github.com/JokerTrickster/joker_backend/services/cloudRepositoryService/features/cloudRepository/usecase"
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

// RegisterRoutes registers all cloud repository routes
func RegisterRoutes(e *echo.Group, db *gorm.DB, bucket string) {
	// Repositories
	uploadRepo := repository.NewUploadCloudRepositoryRepository(db, bucket)
	// batchUploadRepo := repository.NewBatchUploadCloudRepositoryRepository(db, bucket) // Unused as usecase reuses uploadUC
	downloadRepo := repository.NewDownloadCloudRepositoryRepository(db, bucket)
	listRepo := repository.NewListCloudRepositoryRepository(db, bucket)
	deleteRepo := repository.NewDeleteCloudRepositoryRepository(db, bucket)

	// UseCases
	uploadUC := usecase.NewUploadCloudRepositoryUseCase(uploadRepo, 10*time.Second)
	batchUploadUC := usecase.NewBatchUploadCloudRepositoryUseCase(uploadUC, 10*time.Second) // Reuses uploadUC logic
	downloadUC := usecase.NewDownloadCloudRepositoryUseCase(downloadRepo, 10*time.Second)
	listUC := usecase.NewListCloudRepositoryUseCase(listRepo, 10*time.Second)
	deleteUC := usecase.NewDeleteCloudRepositoryUseCase(deleteRepo, 10*time.Second)

	// Handlers
	NewUploadCloudRepositoryHandler(e, uploadUC)
	NewBatchUploadCloudRepositoryHandler(e, batchUploadUC)
	NewDownloadCloudRepositoryHandler(e, downloadUC)
	NewListCloudRepositoryHandler(e, listUC)
	NewDeleteCloudRepositoryHandler(e, deleteUC)

}
