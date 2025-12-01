package queue

import (
	"fmt"
	"os"

	"github.com/hibiken/asynq"
)

// Config holds Redis configuration for async queue
type Config struct {
	RedisAddr     string
	RedisPassword string
	RedisDB       int
}

// NewConfig creates a new queue configuration from environment variables
func NewConfig() *Config {
	redisAddr := os.Getenv("REDIS_HOST")
	if redisAddr == "" {
		redisAddr = "localhost:6379"
	}

	return &Config{
		RedisAddr:     redisAddr,
		RedisPassword: os.Getenv("REDIS_PASSWORD"),
		RedisDB:       0,
	}
}

// NewClient creates a new asynq client for enqueuing tasks
func NewClient(cfg *Config) *asynq.Client {
	return asynq.NewClient(asynq.RedisClientOpt{
		Addr:     cfg.RedisAddr,
		Password: cfg.RedisPassword,
		DB:       cfg.RedisDB,
	})
}

// NewServer creates a new asynq server for processing tasks
func NewServer(cfg *Config) *asynq.Server {
	return asynq.NewServer(
		asynq.RedisClientOpt{
			Addr:     cfg.RedisAddr,
			Password: cfg.RedisPassword,
			DB:       cfg.RedisDB,
		},
		asynq.Config{
			Concurrency: 10, // Process up to 10 tasks concurrently
			Queues: map[string]int{
				"critical": 6,
				"default":  3,
				"low":      1,
			},
		},
	)
}

// TestConnection tests the Redis connection
func TestConnection(cfg *Config) error {
	client := NewClient(cfg)
	defer client.Close()

	inspector := asynq.NewInspector(asynq.RedisClientOpt{
		Addr:     cfg.RedisAddr,
		Password: cfg.RedisPassword,
		DB:       cfg.RedisDB,
	})

	_, err := inspector.Queues()
	if err != nil {
		return fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return nil
}
