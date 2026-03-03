package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/JokerTrickster/joker_backend/services/cloudRepositoryService/features/cloudRepository/model/entity"
	"github.com/JokerTrickster/joker_backend/services/cloudRepositoryService/features/cloudRepository/model/request"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockListRepository mocks IListCloudRepositoryRepository
type MockListRepository struct {
	mock.Mock
}

func (m *MockListRepository) GetFilesByUserID(ctx context.Context, userID uint, filter request.ListFilesRequestDTO) ([]entity.CloudFile, int64, error) {
	args := m.Called(ctx, userID, filter)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]entity.CloudFile), args.Get(1).(int64), args.Error(2)
}

func (m *MockListRepository) GeneratePresignedDownloadURL(ctx context.Context, s3Key string, expiration time.Duration) (string, error) {
	args := m.Called(ctx, s3Key, expiration)
	return args.String(0), args.Error(1)
}

func TestListFiles_SuccessWithFiles(t *testing.T) {
	t.Logf("TestListFiles_SuccessWithFiles: verifying list returns files with presigned URLs")

	mockRepo := new(MockListRepository)
	uc := NewListCloudRepositoryUseCase(mockRepo, 5*time.Second)

	ctx := context.Background()
	userID := uint(1)
	req := request.ListFilesRequestDTO{Page: 1, PageSize: 20}

	mockFiles := []entity.CloudFile{
		{
			ID:        1,
			UserID:    userID,
			FileName:  "file1.jpg",
			S3Key:     "user1/file1.jpg",
			FileType:  entity.FileTypeImage,
			ContentType: "image/jpeg",
			FileSize:   1024,
		},
	}

	mockRepo.On("GetFilesByUserID", mock.Anything, userID, req).Return(mockFiles, int64(1), nil)
	mockRepo.On("GeneratePresignedDownloadURL", mock.Anything, "user1/file1.jpg", 1*time.Hour).Return("https://s3.url/file1", nil)

	result, err := uc.ListFiles(ctx, userID, req)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result.Files, 1)
	assert.Equal(t, "file1.jpg", result.Files[0].FileName)
	assert.Equal(t, "https://s3.url/file1", result.Files[0].DownloadURL)
	assert.Equal(t, int64(1), result.TotalCount)
	t.Logf("Success: files listed with presigned URLs")

	mockRepo.AssertExpectations(t)
}

func TestListFiles_EmptyResult(t *testing.T) {
	t.Logf("TestListFiles_EmptyResult: verifying empty list returns empty slice")

	mockRepo := new(MockListRepository)
	uc := NewListCloudRepositoryUseCase(mockRepo, 5*time.Second)

	ctx := context.Background()
	userID := uint(1)
	req := request.ListFilesRequestDTO{Page: 1, PageSize: 20}

	mockRepo.On("GetFilesByUserID", mock.Anything, userID, req).Return([]entity.CloudFile{}, int64(0), nil)

	result, err := uc.ListFiles(ctx, userID, req)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result.Files, 0)
	assert.Equal(t, int64(0), result.TotalCount)
	t.Logf("Success: empty list returned correctly")

	mockRepo.AssertExpectations(t)
}

func TestListFiles_DefaultPagination(t *testing.T) {
	t.Logf("TestListFiles_DefaultPagination: verifying defaults applied when page/size invalid")

	mockRepo := new(MockListRepository)
	uc := NewListCloudRepositoryUseCase(mockRepo, 5*time.Second)

	ctx := context.Background()
	userID := uint(1)
	req := request.ListFilesRequestDTO{Page: 0, PageSize: 0}

	mockRepo.On("GetFilesByUserID", mock.Anything, userID, mock.MatchedBy(func(f request.ListFilesRequestDTO) bool {
		return f.Page == 1 && f.PageSize == 20
	})).Return([]entity.CloudFile{}, int64(0), nil)

	result, err := uc.ListFiles(ctx, userID, req)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, 1, result.Page)
	assert.Equal(t, 20, result.PageSize)
	t.Logf("Success: defaults applied (page=1, pageSize=20)")

	mockRepo.AssertExpectations(t)
}

func TestListFiles_InvalidFileType(t *testing.T) {
	t.Logf("TestListFiles_InvalidFileType: verifying error for invalid file type")

	mockRepo := new(MockListRepository)
	uc := NewListCloudRepositoryUseCase(mockRepo, 5*time.Second)

	ctx := context.Background()
	userID := uint(1)
	req := request.ListFilesRequestDTO{Page: 1, PageSize: 20, FileType: "invalid"}

	result, err := uc.ListFiles(ctx, userID, req)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "invalid file type")
	t.Logf("Expected: GetFilesByUserID should not be called for invalid file type")

	mockRepo.AssertNotCalled(t, "GetFilesByUserID")
}

func TestListFiles_RepoError(t *testing.T) {
	t.Logf("TestListFiles_RepoError: verifying error propagation from repository")

	mockRepo := new(MockListRepository)
	uc := NewListCloudRepositoryUseCase(mockRepo, 5*time.Second)

	ctx := context.Background()
	userID := uint(1)
	req := request.ListFilesRequestDTO{Page: 1, PageSize: 20}

	mockRepo.On("GetFilesByUserID", mock.Anything, userID, req).Return(nil, int64(0), errors.New("database error"))

	result, err := uc.ListFiles(ctx, userID, req)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to list files")
	t.Logf("Expected: repository error propagated")

	mockRepo.AssertExpectations(t)
}
