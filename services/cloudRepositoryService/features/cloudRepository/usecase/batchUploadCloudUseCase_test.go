package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/JokerTrickster/joker_backend/services/cloudRepositoryService/features/cloudRepository/model/request"
	"github.com/JokerTrickster/joker_backend/services/cloudRepositoryService/features/cloudRepository/model/response"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockUploadUseCase mocks IUploadCloudRepositoryUseCase for batch upload
type MockUploadUseCase struct {
	mock.Mock
}

func (m *MockUploadUseCase) RequestUploadURL(ctx context.Context, userID uint, req *request.UploadRequestDTO) (*response.UploadResponseDTO, error) {
	args := m.Called(ctx, userID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*response.UploadResponseDTO), args.Error(1)
}

func TestRequestBatchUploadURL_Success(t *testing.T) {
	t.Logf("TestRequestBatchUploadURL_Success: verifying multiple files processed successfully")

	mockUploadUC := new(MockUploadUseCase)
	uc := NewBatchUploadCloudRepositoryUseCase(mockUploadUC, 5*time.Second)

	ctx := context.Background()
	userID := uint(1)
	req := &request.BatchUploadRequestDTO{
		Files: []request.UploadRequestDTO{
			{FileName: "file1.jpg", ContentType: "image/jpeg", FileType: "image", FileSize: 1024},
			{FileName: "file2.png", ContentType: "image/png", FileType: "image", FileSize: 2048},
		},
	}

	mockUploadUC.On("RequestUploadURL", mock.Anything, userID, mock.AnythingOfType("*request.UploadRequestDTO")).
		Return(&response.UploadResponseDTO{UploadURL: "https://url1", S3Key: "key1", FileID: 1}, nil).Times(2)

	result, err := uc.RequestBatchUploadURL(ctx, userID, req)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result.Results, 2)
	assert.Equal(t, 2, result.TotalCount)
	assert.Equal(t, 2, result.SuccessCount)
	assert.Equal(t, 0, result.FailedCount)
	t.Logf("Success: batch upload processed 2 files")

	mockUploadUC.AssertExpectations(t)
}

func TestRequestBatchUploadURL_EmptyFiles(t *testing.T) {
	t.Logf("TestRequestBatchUploadURL_EmptyFiles: verifying error when no files provided")

	mockUploadUC := new(MockUploadUseCase)
	uc := NewBatchUploadCloudRepositoryUseCase(mockUploadUC, 5*time.Second)

	ctx := context.Background()
	userID := uint(1)
	req := &request.BatchUploadRequestDTO{Files: []request.UploadRequestDTO{}}

	result, err := uc.RequestBatchUploadURL(ctx, userID, req)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "no files provided")
	t.Logf("Expected: RequestUploadURL should not be called")

	mockUploadUC.AssertNotCalled(t, "RequestUploadURL")
}

func TestRequestBatchUploadURL_ExceedsMax30(t *testing.T) {
	t.Logf("TestRequestBatchUploadURL_ExceedsMax30: verifying error when more than 30 files")

	mockUploadUC := new(MockUploadUseCase)
	uc := NewBatchUploadCloudRepositoryUseCase(mockUploadUC, 5*time.Second)

	ctx := context.Background()
	userID := uint(1)
	files := make([]request.UploadRequestDTO, 31)
	for i := 0; i < 31; i++ {
		files[i] = request.UploadRequestDTO{FileName: "file.jpg", ContentType: "image/jpeg", FileType: "image", FileSize: 1024}
	}
	req := &request.BatchUploadRequestDTO{Files: files}

	result, err := uc.RequestBatchUploadURL(ctx, userID, req)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "maximum 30 files allowed")
	t.Logf("Expected: RequestUploadURL should not be called")

	mockUploadUC.AssertNotCalled(t, "RequestUploadURL")
}

func TestRequestBatchUploadURL_PartialFailure(t *testing.T) {
	t.Logf("TestRequestBatchUploadURL_PartialFailure: verifying partial success when some uploads fail")

	mockUploadUC := new(MockUploadUseCase)
	uc := NewBatchUploadCloudRepositoryUseCase(mockUploadUC, 5*time.Second)

	ctx := context.Background()
	userID := uint(1)
	req := &request.BatchUploadRequestDTO{
		Files: []request.UploadRequestDTO{
			{FileName: "ok.jpg", ContentType: "image/jpeg", FileType: "image", FileSize: 1024},
			{FileName: "fail.jpg", ContentType: "image/jpeg", FileType: "image", FileSize: 1024},
			{FileName: "ok2.jpg", ContentType: "image/jpeg", FileType: "image", FileSize: 1024},
		},
	}

	mockUploadUC.On("RequestUploadURL", mock.Anything, userID, mock.AnythingOfType("*request.UploadRequestDTO")).
		Return(&response.UploadResponseDTO{UploadURL: "https://url1", S3Key: "key1", FileID: 1}, nil).Once()
	mockUploadUC.On("RequestUploadURL", mock.Anything, userID, mock.AnythingOfType("*request.UploadRequestDTO")).
		Return(nil, errors.New("validation failed")).Once()
	mockUploadUC.On("RequestUploadURL", mock.Anything, userID, mock.AnythingOfType("*request.UploadRequestDTO")).
		Return(&response.UploadResponseDTO{UploadURL: "https://url3", S3Key: "key3", FileID: 3}, nil).Once()

	result, err := uc.RequestBatchUploadURL(ctx, userID, req)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result.Results, 2)
	assert.Equal(t, 3, result.TotalCount)
	assert.Equal(t, 2, result.SuccessCount)
	assert.Equal(t, 1, result.FailedCount)
	t.Logf("Success: partial failure handled - 2 success, 1 failed")

	mockUploadUC.AssertExpectations(t)
}
