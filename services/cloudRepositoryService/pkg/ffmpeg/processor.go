package ffmpeg

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	ffmpeg "github.com/u2takey/ffmpeg-go"
)

// Processor handles video processing operations
type Processor struct {
	s3Client *s3.Client
	bucket   string
}

// NewProcessor creates a new FFmpeg processor
func NewProcessor(s3Client *s3.Client, bucket string) *Processor {
	return &Processor{
		s3Client: s3Client,
		bucket:   bucket,
	}
}

// ProbeData represents ffprobe JSON output
type ProbeData struct {
	Format struct {
		Duration string `json:"duration"`
	} `json:"format"`
}

// ExtractDuration extracts the duration of a video file
func (p *Processor) ExtractDuration(ctx context.Context, s3Key string) (float64, error) {
	// Download video to temporary file
	tmpFile, err := p.downloadToTemp(ctx, s3Key)
	if err != nil {
		return 0, fmt.Errorf("failed to download video: %w", err)
	}
	defer os.Remove(tmpFile)

	// Probe video metadata
	data, err := ffmpeg.Probe(tmpFile)
	if err != nil {
		return 0, fmt.Errorf("failed to probe video: %w", err)
	}

	// Parse JSON output
	var probeData ProbeData
	if err := json.Unmarshal([]byte(data), &probeData); err != nil {
		return 0, fmt.Errorf("failed to parse probe data: %w", err)
	}

	// Convert duration string to float64
	duration, err := strconv.ParseFloat(probeData.Format.Duration, 64)
	if err != nil {
		return 0, fmt.Errorf("failed to parse duration: %w", err)
	}

	return duration, nil
}

// GenerateThumbnail generates a thumbnail from a video file
func (p *Processor) GenerateThumbnail(ctx context.Context, s3Key string, seekTime float64) ([]byte, error) {
	// Download video to temporary file
	tmpFile, err := p.downloadToTemp(ctx, s3Key)
	if err != nil {
		return nil, fmt.Errorf("failed to download video: %w", err)
	}
	defer os.Remove(tmpFile)

	// Create temporary output file for thumbnail
	tmpOutput := filepath.Join(os.TempDir(), fmt.Sprintf("thumb_%s.jpg", filepath.Base(s3Key)))
	defer os.Remove(tmpOutput)

	// Generate thumbnail at specified seek time
	err = ffmpeg.Input(tmpFile, ffmpeg.KwArgs{"ss": seekTime}).
		Output(tmpOutput, ffmpeg.KwArgs{
			"vframes": 1,
			"s":       "200x200",
			"f":       "image2",
		}).
		OverWriteOutput().
		ErrorToStdOut().
		Run()

	if err != nil {
		return nil, fmt.Errorf("failed to generate thumbnail: %w", err)
	}

	// Read thumbnail data
	data, err := os.ReadFile(tmpOutput)
	if err != nil {
		return nil, fmt.Errorf("failed to read thumbnail: %w", err)
	}

	return data, nil
}

// UploadThumbnail uploads a thumbnail to S3
func (p *Processor) UploadThumbnail(ctx context.Context, thumbnailKey string, data []byte) error {
	_, err := p.s3Client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(p.bucket),
		Key:         aws.String(thumbnailKey),
		Body:        bytes.NewReader(data),
		ContentType: aws.String("image/jpeg"),
	})

	if err != nil {
		return fmt.Errorf("failed to upload thumbnail: %w", err)
	}

	return nil
}

// downloadToTemp downloads an S3 object to a temporary file
func (p *Processor) downloadToTemp(ctx context.Context, s3Key string) (string, error) {
	// Get object from S3
	result, err := p.s3Client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(p.bucket),
		Key:    aws.String(s3Key),
	})
	if err != nil {
		return "", fmt.Errorf("failed to get object from S3: %w", err)
	}
	defer result.Body.Close()

	// Create temporary file
	tmpFile := filepath.Join(os.TempDir(), filepath.Base(s3Key))
	file, err := os.Create(tmpFile)
	if err != nil {
		return "", fmt.Errorf("failed to create temp file: %w", err)
	}
	defer file.Close()

	// Copy S3 content to temp file
	_, err = io.Copy(file, result.Body)
	if err != nil {
		os.Remove(tmpFile)
		return "", fmt.Errorf("failed to write temp file: %w", err)
	}

	return tmpFile, nil
}
