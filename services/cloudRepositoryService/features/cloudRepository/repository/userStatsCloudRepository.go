package repository

import (
	"context"
	"time"

	"github.com/JokerTrickster/joker_backend/services/cloudRepositoryService/features/cloudRepository/model/entity"
	_interface "github.com/JokerTrickster/joker_backend/services/cloudRepositoryService/features/cloudRepository/model/interface"
	"gorm.io/gorm"
)

type UserStatsCloudRepositoryRepository struct {
	db *gorm.DB
}

func NewUserStatsCloudRepositoryRepository(db *gorm.DB) _interface.IUserStatsCloudRepositoryRepository {
	return &UserStatsCloudRepositoryRepository{
		db: db,
	}
}

// GetTotalStorageUsed gets the total storage used by a user
func (r *UserStatsCloudRepositoryRepository) GetTotalStorageUsed(ctx context.Context, userID uint) (int64, error) {
	var total int64
	err := r.db.WithContext(ctx).
		Model(&entity.CloudFile{}).
		Where("user_id = ? AND deleted_at IS NULL", userID).
		Select("COALESCE(SUM(file_size), 0)").
		Scan(&total).Error

	return total, err
}

// GetMonthlyUploadCount gets the number of uploads in a specific month
func (r *UserStatsCloudRepositoryRepository) GetMonthlyUploadCount(ctx context.Context, userID uint, year int, month int) (int, error) {
	var count int64
	startDate := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
	endDate := startDate.AddDate(0, 1, 0)

	err := r.db.WithContext(ctx).
		Model(&entity.ActivityLog{}).
		Where("user_id = ? AND activity_type = ? AND created_at >= ? AND created_at < ?",
			userID, entity.ActivityTypeUpload, startDate, endDate).
		Count(&count).Error

	return int(count), err
}

// GetMonthlyDownloadCount gets the number of downloads in a specific month
func (r *UserStatsCloudRepositoryRepository) GetMonthlyDownloadCount(ctx context.Context, userID uint, year int, month int) (int, error) {
	var count int64
	startDate := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
	endDate := startDate.AddDate(0, 1, 0)

	err := r.db.WithContext(ctx).
		Model(&entity.ActivityLog{}).
		Where("user_id = ? AND activity_type = ? AND created_at >= ? AND created_at < ?",
			userID, entity.ActivityTypeDownload, startDate, endDate).
		Count(&count).Error

	return int(count), err
}

// GetMonthlyTagsCreatedCount gets the number of unique tags created in a specific month
func (r *UserStatsCloudRepositoryRepository) GetMonthlyTagsCreatedCount(ctx context.Context, userID uint, year int, month int) (int, error) {
	var count int64
	startDate := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
	endDate := startDate.AddDate(0, 1, 0)

	// Count unique tags added this month
	err := r.db.WithContext(ctx).
		Model(&entity.ActivityLog{}).
		Where("user_id = ? AND activity_type = ? AND created_at >= ? AND created_at < ?",
			userID, entity.ActivityTypeTagAdd, startDate, endDate).
		Distinct("tag_name").
		Count(&count).Error

	return int(count), err
}

// LogActivity logs a user activity
func (r *UserStatsCloudRepositoryRepository) LogActivity(ctx context.Context, activity *entity.ActivityLog) error {
	return r.db.WithContext(ctx).Create(activity).Error
}