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
	userStatsRepo := repository.NewUserStatsCloudRepositoryRepository(db)
	activityHistoryRepo := repository.NewActivityHistoryCloudRepositoryRepository(db)
	favoriteRepo := repository.NewFavoriteRepository(db)

	// UseCases - using 30s timeout to match Echo server timeout and provide buffer for DB operations
	uploadUC := usecase.NewUploadCloudRepositoryUseCase(uploadRepo, userStatsRepo, db, 30*time.Second)
	batchUploadUC := usecase.NewBatchUploadCloudRepositoryUseCase(uploadUC, 30*time.Second) // Reuses uploadUC logic
	downloadUC := usecase.NewDownloadCloudRepositoryUseCase(downloadRepo, userStatsRepo, 30*time.Second)
	listUC := usecase.NewListCloudRepositoryUseCase(listRepo, 30*time.Second)
	deleteUC := usecase.NewDeleteCloudRepositoryUseCase(deleteRepo, 30*time.Second)
	userStatsUC := usecase.NewUserStatsCloudRepositoryUseCase(userStatsRepo, 30*time.Second)
	activityHistoryUC := usecase.NewActivityHistoryCloudRepositoryUseCase(activityHistoryRepo, 30*time.Second)
	favoriteUC := usecase.NewFavoriteUseCase(favoriteRepo, downloadRepo, listRepo, 30*time.Second)

	// Handlers
	NewUploadCloudRepositoryHandler(e, uploadUC)
	NewBatchUploadCloudRepositoryHandler(e, batchUploadUC)
	NewDownloadCloudRepositoryHandler(e, downloadUC)
	NewListCloudRepositoryHandler(e, listUC)
	NewDeleteCloudRepositoryHandler(e, deleteUC)
	NewUserStatsCloudRepositoryHandler(e, userStatsUC)
	NewActivityHistoryCloudRepositoryHandler(e, activityHistoryUC)
	NewFavoriteHandler(e, favoriteUC)

}
