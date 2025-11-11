package scheduler

import (
	"context"
	"time"

	"github.com/JokerTrickster/joker_backend/services/weatherService/features/weather/model/entity"
	"github.com/JokerTrickster/joker_backend/services/weatherService/pkg/metrics"
)

// processAlarmsWithMetrics wraps processAlarms with metrics collection
func (s *WeatherSchedulerService) processAlarmsWithMetrics(ctx context.Context, targetTime time.Time) error {
	startTime := time.Now()

	// Get alarms to notify
	alarms, err := s.repository.GetAlarmsToNotify(ctx, targetTime)
	if err != nil {
		return err
	}

	if len(alarms) == 0 {
		// Record tick with no alarms
		metrics.RecordSchedulerTick(0, time.Since(startTime))
		return nil
	}

	processed := 0
	failed := 0

	// Process each alarm
	for _, alarm := range alarms {
		if err := s.processAlarmWithMetrics(ctx, alarm); err != nil {
			failed++
			metrics.RecordSchedulerAlarmStatus("failed")
			continue
		}
		processed++
		metrics.RecordSchedulerAlarmStatus("success")
	}

	duration := time.Since(startTime)
	metrics.RecordSchedulerTick(processed, duration)

	return nil
}

// processAlarmWithMetrics wraps processAlarm with metrics collection
func (s *WeatherSchedulerService) processAlarmWithMetrics(ctx context.Context, alarm entity.UserAlarm) error {
	// Get weather data with metrics
	weatherData, err := s.getWeatherDataWithMetrics(ctx, alarm.Region)
	if err != nil {
		return err
	}

	// Get FCM tokens
	tokenEntities, err := s.repository.GetFCMTokens(ctx, alarm.UserID)
	if err != nil {
		return err
	}

	if len(tokenEntities) == 0 {
		// Update last_sent even without tokens
		if err := s.repository.UpdateLastSent(ctx, alarm.ID, time.Now()); err != nil {
			s.logger.Error("Failed to update last_sent for alarm with no tokens")
		}
		return nil
	}

	// Extract token strings
	tokens := make([]string, len(tokenEntities))
	for i, entity := range tokenEntities {
		tokens[i] = entity.FCMToken
	}

	// Send notification with metrics
	startTime := time.Now()
	err = s.notifier.SendWeatherNotification(ctx, tokens, weatherData, alarm.Region)
	duration := time.Since(startTime)

	if err != nil {
		metrics.RecordFCMSent("failure", duration, len(tokens))
		metrics.RecordFCMError("send_failed")
	} else {
		metrics.RecordFCMSent("success", duration, len(tokens))
	}

	// Update last_sent timestamp
	if err := s.repository.UpdateLastSent(ctx, alarm.ID, time.Now()); err != nil {
		return err
	}

	return nil
}

// getWeatherDataWithMetrics wraps weather data retrieval with metrics
func (s *WeatherSchedulerService) getWeatherDataWithMetrics(ctx context.Context, region string) (*entity.WeatherData, error) {
	// Try cache first
	weatherData, err := s.cache.Get(ctx, region)
	if err != nil {
		metrics.RecordCacheError("get")
		s.logger.Warn("Cache error, falling back to crawler")
		weatherData = nil
	}

	if weatherData != nil {
		// Cache hit
		metrics.RecordCacheHit()
		return weatherData, nil
	}

	// Cache miss - fetch from crawler
	metrics.RecordCacheMiss()

	startTime := time.Now()
	weatherData, err = s.crawler.Fetch(ctx, region)
	duration := time.Since(startTime)

	if err != nil {
		metrics.RecordCrawlRequest(region, "failure", duration)
		metrics.RecordCrawlError(region, "fetch_failed")
		return nil, err
	}

	metrics.RecordCrawlRequest(region, "success", duration)

	// Cache the fetched data
	if err := s.cache.Set(ctx, region, weatherData); err != nil {
		metrics.RecordCacheError("set")
		s.logger.Warn("Failed to cache weather data")
	}

	return weatherData, nil
}
