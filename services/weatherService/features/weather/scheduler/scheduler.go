package scheduler

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/JokerTrickster/joker_backend/services/weatherService/features/weather/model/entity"
	_interface "github.com/JokerTrickster/joker_backend/services/weatherService/features/weather/model/interface"
	"go.uber.org/zap"
)

// WeatherSchedulerService manages periodic weather alarm notifications
type WeatherSchedulerService struct {
	repository _interface.ISchedulerWeatherRepository
	crawler    _interface.IWeatherCrawler
	cache      _interface.IWeatherCache
	notifier   _interface.IFCMNotifier
	logger     *zap.Logger
	interval   time.Duration
	stopChan   chan struct{}
	wg         sync.WaitGroup
	mu         sync.Mutex
	running    bool
}

// NewWeatherSchedulerService creates a new scheduler service instance
func NewWeatherSchedulerService(
	repository _interface.ISchedulerWeatherRepository,
	crawler _interface.IWeatherCrawler,
	cache _interface.IWeatherCache,
	notifier _interface.IFCMNotifier,
	logger *zap.Logger,
	interval time.Duration,
) *WeatherSchedulerService {
	if logger == nil {
		logger, _ = zap.NewProduction()
	}

	return &WeatherSchedulerService{
		repository: repository,
		crawler:    crawler,
		cache:      cache,
		notifier:   notifier,
		logger:     logger,
		interval:   interval,
		stopChan:   make(chan struct{}),
		running:    false,
	}
}

// Start begins the scheduler goroutine with ticker
// Creates 1-minute ticker that processes alarms at (current_time + 1 minute)
func (s *WeatherSchedulerService) Start(ctx context.Context) error {
	s.mu.Lock()
	if s.running {
		s.mu.Unlock()
		return fmt.Errorf("scheduler is already running")
	}
	s.running = true
	s.mu.Unlock()

	s.logger.Info("Starting weather scheduler service",
		zap.Duration("interval", s.interval))

	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()

	// Process immediately on start
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		targetTime := time.Now().Add(s.interval)
		if err := s.processAlarms(ctx, targetTime); err != nil {
			s.logger.Error("Failed to process alarms on startup",
				zap.Error(err))
		}
	}()

	for {
		select {
		case <-ctx.Done():
			s.logger.Info("Scheduler context cancelled, shutting down")
			return ctx.Err()

		case <-s.stopChan:
			s.logger.Info("Scheduler stop signal received, shutting down gracefully")
			return nil

		case tickTime := <-ticker.C:
			// Calculate target time (current tick + 1 minute)
			targetTime := tickTime.Add(s.interval)

			s.logger.Debug("Scheduler tick",
				zap.Time("tick_time", tickTime),
				zap.Time("target_time", targetTime))

			// Process alarms in separate goroutine with WaitGroup tracking
			s.wg.Add(1)
			go func(target time.Time) {
				defer s.wg.Done()

				if err := s.processAlarms(ctx, target); err != nil {
					s.logger.Error("Failed to process alarms",
						zap.Time("target_time", target),
						zap.Error(err))
				}
			}(targetTime)
		}
	}
}

// Stop gracefully shuts down the scheduler
// Waits for in-flight alarm processing to complete (max 30 seconds)
func (s *WeatherSchedulerService) Stop() error {
	s.mu.Lock()
	if !s.running {
		s.mu.Unlock()
		return nil // Already stopped, idempotent
	}
	s.running = false
	s.mu.Unlock()

	s.logger.Info("Stopping weather scheduler service")

	// Signal stop
	close(s.stopChan)

	// Wait for in-flight processing with timeout
	done := make(chan struct{})
	go func() {
		s.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		s.logger.Info("Scheduler stopped gracefully")
	case <-time.After(30 * time.Second):
		s.logger.Warn("Scheduler stop timeout, some operations may not have completed")
	}

	return nil
}

// processAlarms handles alarm processing for the target time
// Returns summary of processed/failed alarms
func (s *WeatherSchedulerService) processAlarms(ctx context.Context, targetTime time.Time) error {
	startTime := time.Now()

	s.logger.Info("Processing alarms",
		zap.Time("target_time", targetTime))

	// Delegate to metrics-enabled version
	err := s.processAlarmsWithMetrics(ctx, targetTime)
	if err != nil {
		return err
	}

	duration := time.Since(startTime)
	s.logger.Info("Completed alarm processing",
		zap.Duration("duration", duration))

	return nil
}

// processAlarm handles a single alarm notification
func (s *WeatherSchedulerService) processAlarm(ctx context.Context, alarm entity.UserAlarm) error {
	s.logger.Info("Processing individual alarm",
		zap.Int("alarm_id", alarm.ID),
		zap.Int("user_id", alarm.UserID),
		zap.String("region", alarm.Region))

	// Step 1: Get weather data (cache first, then crawler)
	weatherData, err := s.cache.Get(ctx, alarm.Region)
	if err != nil {
		s.logger.Warn("Cache error, falling back to crawler",
			zap.String("region", alarm.Region),
			zap.Error(err))
		weatherData = nil // Force crawler fallback
	}

	if weatherData == nil {
		// Cache miss - fetch from crawler
		s.logger.Debug("Cache miss, fetching from crawler",
			zap.String("region", alarm.Region))

		weatherData, err = s.crawler.Fetch(ctx, alarm.Region)
		if err != nil {
			return fmt.Errorf("failed to fetch weather data from crawler: %w", err)
		}

		// Cache the fetched data
		if err := s.cache.Set(ctx, alarm.Region, weatherData); err != nil {
			s.logger.Warn("Failed to cache weather data",
				zap.String("region", alarm.Region),
				zap.Error(err))
			// Continue even if caching fails
		}
	} else {
		s.logger.Debug("Cache hit",
			zap.String("region", alarm.Region))
	}

	// Step 2: Get FCM tokens for user
	tokenEntities, err := s.repository.GetFCMTokens(ctx, alarm.UserID)
	if err != nil {
		return fmt.Errorf("failed to get FCM tokens: %w", err)
	}

	if len(tokenEntities) == 0 {
		s.logger.Warn("No FCM tokens found for user",
			zap.Int("user_id", alarm.UserID))
		// Still update last_sent to prevent retry
		if err := s.repository.UpdateLastSent(ctx, alarm.ID, time.Now()); err != nil {
			s.logger.Error("Failed to update last_sent for alarm with no tokens",
				zap.Int("alarm_id", alarm.ID),
				zap.Error(err))
		}
		return nil
	}

	// Extract token strings
	tokens := make([]string, len(tokenEntities))
	for i, entity := range tokenEntities {
		tokens[i] = entity.FCMToken
	}

	s.logger.Debug("Sending FCM notification",
		zap.Int("user_id", alarm.UserID),
		zap.Int("token_count", len(tokens)),
		zap.String("region", alarm.Region))

	// Step 3: Send notification
	if err := s.notifier.SendWeatherNotification(ctx, tokens, weatherData, alarm.Region); err != nil {
		s.logger.Error("FCM notification failed",
			zap.Int("user_id", alarm.UserID),
			zap.String("region", alarm.Region),
			zap.Error(err))
		// Mark as sent anyway to prevent retry storm
	}

	// Step 4: Update last_sent timestamp
	if err := s.repository.UpdateLastSent(ctx, alarm.ID, time.Now()); err != nil {
		return fmt.Errorf("failed to update last_sent: %w", err)
	}

	s.logger.Info("Successfully processed alarm",
		zap.Int("alarm_id", alarm.ID),
		zap.Int("user_id", alarm.UserID),
		zap.String("region", alarm.Region))

	return nil
}
