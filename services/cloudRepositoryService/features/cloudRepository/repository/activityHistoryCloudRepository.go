package repository

import (
	"context"
	"time"

	"github.com/JokerTrickster/joker_backend/services/cloudRepositoryService/features/cloudRepository/model/entity"
	_interface "github.com/JokerTrickster/joker_backend/services/cloudRepositoryService/features/cloudRepository/model/interface"
	"gorm.io/gorm"
)

type ActivityHistoryCloudRepositoryRepository struct {
	db *gorm.DB
}

func NewActivityHistoryCloudRepositoryRepository(db *gorm.DB) _interface.IActivityHistoryCloudRepositoryRepository {
	return &ActivityHistoryCloudRepositoryRepository{
		db: db,
	}
}

// GetMonthlyActivity retrieves all activities for a user in a specific month
func (r *ActivityHistoryCloudRepositoryRepository) GetMonthlyActivity(ctx context.Context, userID uint, year int, month int) ([]entity.ActivityLog, error) {
	var activities []entity.ActivityLog
	startDate := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
	endDate := startDate.AddDate(0, 1, 0)

	err := r.db.WithContext(ctx).
		Where("user_id = ? AND created_at >= ? AND created_at < ?",
			userID, startDate, endDate).
		Order("created_at DESC").
		Find(&activities).Error

	return activities, err
}

// GetMonthlyUsedTags retrieves tags used for each day in a month
func (r *ActivityHistoryCloudRepositoryRepository) GetMonthlyUsedTags(ctx context.Context, userID uint, year int, month int) (map[string][]string, error) {
	startDate := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
	endDate := startDate.AddDate(0, 1, 0)

	// Query to get unique tags per day from file uploads and tag activities
	query := `
		SELECT DISTINCT
			DATE(cf.created_at) as date,
			t.name as tag_name
		FROM cloud_files cf
		JOIN file_tags ft ON cf.id = ft.cloud_file_id
		JOIN tags t ON ft.tag_id = t.id
		WHERE cf.user_id = ?
			AND cf.created_at >= ?
			AND cf.created_at < ?
			AND cf.deleted_at IS NULL
		UNION
		SELECT DISTINCT
			DATE(al.created_at) as date,
			al.tag_name
		FROM activity_logs al
		WHERE al.user_id = ?
			AND al.activity_type = ?
			AND al.created_at >= ?
			AND al.created_at < ?
			AND al.tag_name IS NOT NULL
		ORDER BY date, tag_name
	`

	type TagResult struct {
		Date    time.Time `gorm:"column:date"`
		TagName string    `gorm:"column:tag_name"`
	}

	var results []TagResult
	err := r.db.WithContext(ctx).Raw(query,
		userID, startDate, endDate,
		userID, entity.ActivityTypeTagAdd, startDate, endDate).
		Scan(&results).Error

	if err != nil {
		return nil, err
	}

	// Group tags by date
	tagsByDate := make(map[string][]string)
	for _, result := range results {
		dateStr := result.Date.Format("2006-01-02")
		if _, exists := tagsByDate[dateStr]; !exists {
			tagsByDate[dateStr] = []string{}
		}
		// Check if tag already exists in the array for that date
		found := false
		for _, tag := range tagsByDate[dateStr] {
			if tag == result.TagName {
				found = true
				break
			}
		}
		if !found {
			tagsByDate[dateStr] = append(tagsByDate[dateStr], result.TagName)
		}
	}

	return tagsByDate, nil
}