package usecase

import (
	"context"
	"fmt"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/JokerTrickster/joker_backend/services/morandoranService/features/upload/model/response"
	"github.com/JokerTrickster/joker_backend/shared/aws"
	s3svc "github.com/aws/aws-sdk-go-v2/service/s3"
)

const maxFileSize = 5 * 1024 * 1024 // 5MB

var allowedExtensions = map[string]string{
	".jpg":  "image/jpeg",
	".jpeg": "image/jpeg",
	".png":  "image/png",
	".webp": "image/webp",
}

type UploadUseCase struct {
	ContextTimeout time.Duration
}

func NewUploadUseCase(timeout time.Duration) *UploadUseCase {
	return &UploadUseCase{ContextTimeout: timeout}
}

func (uc *UploadUseCase) Upload(c context.Context, file *multipart.FileHeader, fileType string) (*response.ResUpload, error) {
	ctx, cancel := context.WithTimeout(c, uc.ContextTimeout)
	defer cancel()

	if file.Size > maxFileSize {
		return nil, fmt.Errorf("file size exceeds 5MB limit")
	}

	ext := strings.ToLower(filepath.Ext(file.Filename))
	contentType, ok := allowedExtensions[ext]
	if !ok {
		return nil, fmt.Errorf("unsupported file type: %s", ext)
	}

	bucket := os.Getenv("MORANDORAN_S3_BUCKET")
	if bucket == "" {
		bucket = "morandoran-uploads-dev"
	}

	randomName := aws.FileNameGenerateRandom()
	key := fmt.Sprintf("%s/%d_%s%s", fileType, time.Now().Unix(), randomName, ext)

	src, err := file.Open()
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer src.Close()

	if aws.S3Client == nil {
		return nil, fmt.Errorf("S3 client not initialized")
	}

	_, err = aws.S3Client.PutObject(ctx, &s3svc.PutObjectInput{
		Bucket:      &bucket,
		Key:         &key,
		Body:        src,
		ContentType: &contentType,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to upload to S3: %w", err)
	}

	region := os.Getenv("AWS_REGION")
	if region == "" {
		region = "ap-northeast-2"
	}
	url := fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s", bucket, region, key)

	return &response.ResUpload{URL: url}, nil
}
