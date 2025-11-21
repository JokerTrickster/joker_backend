package usecase

import (
	"context"
	"fmt"

	"github.com/JokerTrickster/joker_backend/services/cloudRepositoryService/features/cloudRepository/model"
)

// RequestUploadURL generates a presigned upload URL and creates a file record
func (u *CloudRepositoryUsecase) RequestUploadURL(ctx context.Context, userID uint, req *model.UploadRequestDTO) (*model.UploadResponseDTO, error) {
	// Validate file size
	if req.FileSize > MaxFileSize {
		return nil, fmt.Errorf("file size exceeds maximum allowed size of %d bytes", MaxFileSize)
	}

	// Validate content type
	fileType := model.FileType(req.FileType)
	if fileType == model.FileTypeImage && !AllowedImageTypes[req.ContentType] {
		return nil, fmt.Errorf("invalid image content type: %s", req.ContentType)
	}
	if fileType == model.FileTypeVideo && !AllowedVideoTypes[req.ContentType] {
		return nil, fmt.Errorf("invalid video content type: %s", req.ContentType)
	}

	// Generate S3 key
	s3Key := u.generateS3Key(userID, fileType, req.FileName)

	// Create file record in database
	file := &model.CloudFile{
		UserID:      userID,
		FileName:    req.FileName,
		S3Key:       s3Key,
		FileType:    fileType,
		ContentType: req.ContentType,
		FileSize:    req.FileSize,
	}

	if err := u.repo.CreateFile(ctx, file); err != nil {
		return nil, fmt.Errorf("failed to create file record: %w", err)
	}

	// Generate presigned upload URL
	uploadURL, err := u.repo.GeneratePresignedUploadURL(ctx, s3Key, req.ContentType, DefaultUploadExpiration)
	if err != nil {
		return nil, fmt.Errorf("failed to generate upload URL: %w", err)
	}

	return &model.UploadResponseDTO{
		FileID:    file.ID,
		UploadURL: uploadURL,
		S3Key:     s3Key,
		ExpiresIn: int(DefaultUploadExpiration.Seconds()),
	}, nil
}

// RequestBatchUploadURL generates presigned upload URLs for multiple files (max 30)
func (u *CloudRepositoryUsecase) RequestBatchUploadURL(ctx context.Context, userID uint, req *model.BatchUploadRequestDTO) (*model.BatchUploadResponseDTO, error) {
	if len(req.Files) == 0 {
		return nil, fmt.Errorf("no files provided")
	}

	if len(req.Files) > MaxBatchSize {
		return nil, fmt.Errorf("maximum %d files allowed, got %d", MaxBatchSize, len(req.Files))
	}

	results := make([]model.UploadResponseDTO, 0, len(req.Files))
	successCount := 0
	failedCount := 0

	// Process each file
	for _, fileReq := range req.Files {
		uploadResp, err := u.RequestUploadURL(ctx, userID, &fileReq)
		if err != nil {
			// Log error but continue processing other files
			fmt.Printf("Failed to process file %s: %v\n", fileReq.FileName, err)
			failedCount++
			continue
		}

		results = append(results, *uploadResp)
		successCount++
	}

	return &model.BatchUploadResponseDTO{
		Results:      results,
		TotalCount:   len(req.Files),
		SuccessCount: successCount,
		FailedCount:  failedCount,
	}, nil
}
