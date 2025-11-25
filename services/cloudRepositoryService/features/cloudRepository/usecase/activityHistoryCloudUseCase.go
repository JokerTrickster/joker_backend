package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/JokerTrickster/joker_backend/services/cloudRepositoryService/features/cloudRepository/model/entity"
	_interface "github.com/JokerTrickster/joker_backend/services/cloudRepositoryService/features/cloudRepository/model/interface"
	"github.com/JokerTrickster/joker_backend/services/cloudRepositoryService/features/cloudRepository/model/request"
	"github.com/JokerTrickster/joker_backend/services/cloudRepositoryService/features/cloudRepository/model/response"
)

type ActivityHistoryCloudRepositoryUseCase struct {
	Repo           _interface.IActivityHistoryCloudRepositoryRepository
	ContextTimeout time.Duration
}

func NewActivityHistoryCloudRepositoryUseCase(repo _interface.IActivityHistoryCloudRepositoryRepository, timeout time.Duration) _interface.IActivityHistoryCloudRepositoryUseCase {
	return &ActivityHistoryCloudRepositoryUseCase{
		Repo:           repo,
		ContextTimeout: timeout,
	}
}

// GetActivityHistory retrieves activity history for a specific month
func (u *ActivityHistoryCloudRepositoryUseCase) GetActivityHistory(ctx context.Context, userID uint, req *request.ActivityHistoryRequestDTO) (*response.ActivityHistoryResponseDTO, error) {
	c, cancel := context.WithTimeout(ctx, u.ContextTimeout)
	defer cancel()

	// Parse month from request
	if req.Month == "" {
		// Default to current month if not provided
		now := time.Now()
		req.Month = now.Format("2006-01")
	}

	// Parse year and month from the request
	parsedTime, err := time.Parse("2006-01", req.Month)
	if err != nil {
		return nil, fmt.Errorf("invalid month format, expected YYYY-MM: %w", err)
	}

	year := parsedTime.Year()
	month := int(parsedTime.Month())

	// Get all activities for the month
	activities, err := u.Repo.GetMonthlyActivity(c, userID, year, month)
	if err != nil {
		return nil, err
	}

	// Get tags used per day
	tagsByDate, err := u.Repo.GetMonthlyUsedTags(c, userID, year, month)
	if err != nil {
		return nil, err
	}

	// Aggregate activities by day
	dailyActivities := make(map[string]*response.DailyActivityDTO)

	for _, activity := range activities {
		dateStr := activity.CreatedAt.Format("2006-01-02")

		if _, exists := dailyActivities[dateStr]; !exists {
			dailyActivities[dateStr] = &response.DailyActivityDTO{
				Uploads:   0,
				Downloads: 0,
				Tags:      []string{},
			}
		}

		switch activity.ActivityType {
		case entity.ActivityTypeUpload:
			dailyActivities[dateStr].Uploads++
		case entity.ActivityTypeDownload:
			dailyActivities[dateStr].Downloads++
		}
	}

	// Add tags to daily activities
	for date, tags := range tagsByDate {
		if _, exists := dailyActivities[date]; !exists {
			dailyActivities[date] = &response.DailyActivityDTO{
				Uploads:   0,
				Downloads: 0,
				Tags:      tags,
			}
		} else {
			dailyActivities[date].Tags = tags
		}
	}

	// Convert map to response format
	result := make(response.ActivityHistoryResponseDTO)
	for date, activity := range dailyActivities {
		result[date] = *activity
	}

	return &result, nil
}