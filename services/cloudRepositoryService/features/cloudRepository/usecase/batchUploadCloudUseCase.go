package usecase

import (
	"context"
	"fmt"
	"time"

	_interface "github.com/JokerTrickster/joker_backend/services/cloudRepositoryService/features/cloudRepository/model/interface"
	"github.com/JokerTrickster/joker_backend/services/cloudRepositoryService/features/cloudRepository/model/request"
	"github.com/JokerTrickster/joker_backend/services/cloudRepositoryService/features/cloudRepository/model/response"
)

type BatchUploadCloudRepositoryUseCase struct {
	UploadUseCase  _interface.IUploadCloudRepositoryUseCase
	ContextTimeout time.Duration
}

func NewBatchUploadCloudRepositoryUseCase(uploadUseCase _interface.IUploadCloudRepositoryUseCase, timeout time.Duration) _interface.IBatchUploadCloudRepositoryUseCase {
	return &BatchUploadCloudRepositoryUseCase{
		UploadUseCase:  uploadUseCase,
		ContextTimeout: timeout,
	}
}

// RequestBatchUploadURL generates presigned upload URLs for multiple files (max 30)
func (u *BatchUploadCloudRepositoryUseCase) RequestBatchUploadURL(c context.Context, userID uint, req *request.BatchUploadRequestDTO) (*response.BatchUploadResponseDTO, error) {
	ctx, cancel := context.WithTimeout(c, u.ContextTimeout)
	defer cancel()
	maxCount := 30
	if len(req.Files) == 0 {
		return nil, fmt.Errorf("no files provided")
	}

	if len(req.Files) > maxCount {
		return nil, fmt.Errorf("maximum 30 files allowed, got %d", len(req.Files))
	}

	results := make([]response.UploadResponseDTO, 0, len(req.Files))
	successCount := 0
	failedCount := 0

	// Process each file
	for _, fileReq := range req.Files {
		uploadResp, err := u.UploadUseCase.RequestUploadURL(ctx, userID, &fileReq)
		if err != nil {
			// Log error but continue processing other files
			fmt.Printf("Failed to process file %s: %v\n", fileReq.FileName, err)
			failedCount++
			continue
		}

		results = append(results, *uploadResp)
		successCount++
	}

	return &response.BatchUploadResponseDTO{
		Results:      results,
		TotalCount:   len(req.Files),
		SuccessCount: successCount,
		FailedCount:  failedCount,
	}, nil
}
