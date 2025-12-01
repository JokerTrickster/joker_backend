package queue

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/hibiken/asynq"
)

// Task type constants
const (
	TypeVideoProcessing = "video:process"
)

// VideoProcessingPayload contains information for video processing
type VideoProcessingPayload struct {
	FileID   uint   `json:"file_id"`
	S3Key    string `json:"s3_key"`
	UserID   uint   `json:"user_id"`
	Bucket   string `json:"bucket"`
	FileName string `json:"file_name"`
}

// NewVideoProcessingTask creates a new video processing task
func NewVideoProcessingTask(payload *VideoProcessingPayload) (*asynq.Task, error) {
	data, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal payload: %w", err)
	}

	return asynq.NewTask(TypeVideoProcessing, data), nil
}

// EnqueueVideoProcessing enqueues a video processing task
func EnqueueVideoProcessing(client *asynq.Client, payload *VideoProcessingPayload) error {
	task, err := NewVideoProcessingTask(payload)
	if err != nil {
		return err
	}

	// Enqueue task with options
	info, err := client.Enqueue(
		task,
		asynq.Queue("default"),
		asynq.MaxRetry(3),
		asynq.Timeout(10*time.Minute),
		asynq.ProcessIn(5*time.Second), // Start processing after 5 seconds
	)

	if err != nil {
		return fmt.Errorf("failed to enqueue task: %w", err)
	}

	fmt.Printf("Enqueued video processing task: ID=%s, Queue=%s\n", info.ID, info.Queue)
	return nil
}

// ParseVideoProcessingPayload parses the video processing payload from task
func ParseVideoProcessingPayload(task *asynq.Task) (*VideoProcessingPayload, error) {
	var payload VideoProcessingPayload
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		return nil, fmt.Errorf("failed to unmarshal payload: %w", err)
	}
	return &payload, nil
}
