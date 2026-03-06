package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/JokerTrickster/joker_backend/services/cloudRepositoryService/features/cloudRepository/model/entity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockDownloadRepository mocks IDownloadCloudRepositoryRepository
type MockDownloadRepository struct {
	mock.Mock
}

func (m *MockDownloadRepository) GeneratePresignedDownloadURL(ctx context.Context, s3Key string, expiration time.Duration) (string, error) {
	args := m.Called(ctx, s3Key, expiration)
	return args.String(0), args.Error(1)
}

func (m *MockDownloadRepository) GeneratePresignedDownloadURLWithFilename(ctx context.Context, s3Key, filename string, expiration time.Duration) (string, error) {
	args := m.Called(ctx, s3Key, filename, expiration)
	return args.String(0), args.Error(1)
}

func (m *MockDownloadRepository) GetFileByID(ctx context.Context, id uint) (*entity.CloudFile, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.CloudFile), args.Error(1)
}

// MockFileShareRepository mocks IFileShareRepository for download usecase
type MockFileShareRepository struct {
	mock.Mock
}

func (m *MockFileShareRepository) CreateFileShare(ctx context.Context, share *entity.FileShare) error {
	args := m.Called(ctx, share)
	return args.Error(0)
}

func (m *MockFileShareRepository) GetFileSharesByFileID(ctx context.Context, fileID uint) ([]entity.FileShare, error) {
	args := m.Called(ctx, fileID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]entity.FileShare), args.Error(1)
}

func (m *MockFileShareRepository) GetSharedFilesByUserID(ctx context.Context, userID int32) ([]entity.FileShare, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]entity.FileShare), args.Error(1)
}

func (m *MockFileShareRepository) GetFilesSharedByUserID(ctx context.Context, ownerID int32) ([]entity.FileShare, error) {
	args := m.Called(ctx, ownerID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]entity.FileShare), args.Error(1)
}

func (m *MockFileShareRepository) DeleteFileShare(ctx context.Context, fileID uint, sharedWithID int32, ownerID int32) error {
	args := m.Called(ctx, fileID, sharedWithID, ownerID)
	return args.Error(0)
}

func (m *MockFileShareRepository) HasFileAccess(ctx context.Context, userID int32, fileID uint) (bool, error) {
	args := m.Called(ctx, userID, fileID)
	return args.Bool(0), args.Error(1)
}

func TestRequestDownloadURL_Success(t *testing.T) {
	t.Logf("TestRequestDownloadURL_Success: verifying presigned URL generation on success path")

	mockRepo := new(MockDownloadRepository)
	mockStatsRepo := new(MockUserStatsRepository)
	mockFileShareRepo := new(MockFileShareRepository)

	uc := NewDownloadCloudRepositoryUseCase(mockRepo, mockStatsRepo, mockFileShareRepo, 5*time.Second)

	ctx := context.Background()
	userID := uint(1)
	fileID := uint(100)

	mockFile := &entity.CloudFile{
		ID:       fileID,
		UserID:   userID,
		FileName:  "test.jpg",
		S3Key:    "user1/test.jpg",
		FileType:  entity.FileTypeImage,
		FileSize:  1024,
	}

	mockFileShareRepo.On("HasFileAccess", mock.Anything, int32(userID), fileID).Return(true, nil)
	mockRepo.On("GetFileByID", mock.Anything, fileID).Return(mockFile, nil)
	mockStatsRepo.On("LogActivity", mock.Anything, mock.AnythingOfType("*entity.ActivityLog")).Return(nil)
	mockRepo.On("GeneratePresignedDownloadURLWithFilename", mock.Anything, "user1/test.jpg", "test.jpg", 1*time.Hour).
		Return("https://s3.presigned.url/test.jpg", nil)

	result, err := uc.RequestDownloadURL(ctx, userID, fileID)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "https://s3.presigned.url/test.jpg", result.DownloadURL)
	assert.Equal(t, "test.jpg", result.FileName)
	assert.Equal(t, 3600, result.ExpiresIn)
	t.Logf("Success: presigned URL generated correctly")

	mockFileShareRepo.AssertExpectations(t)
	mockRepo.AssertExpectations(t)
}

func TestRequestDownloadURL_FileNotFound(t *testing.T) {
	t.Logf("TestRequestDownloadURL_FileNotFound: verifying error when file does not exist")

	mockRepo := new(MockDownloadRepository)
	mockStatsRepo := new(MockUserStatsRepository)
	mockFileShareRepo := new(MockFileShareRepository)

	uc := NewDownloadCloudRepositoryUseCase(mockRepo, mockStatsRepo, mockFileShareRepo, 5*time.Second)

	ctx := context.Background()
	userID := uint(1)
	fileID := uint(999)

	mockFileShareRepo.On("HasFileAccess", mock.Anything, int32(userID), fileID).Return(true, nil)
	mockRepo.On("GetFileByID", mock.Anything, fileID).Return(nil, errors.New("record not found"))

	result, err := uc.RequestDownloadURL(ctx, userID, fileID)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "file not found")
	t.Logf("Expected: file not found error returned")

	mockFileShareRepo.AssertExpectations(t)
	mockRepo.AssertExpectations(t)
}

func TestRequestDownloadURL_NoAccess(t *testing.T) {
	t.Logf("TestRequestDownloadURL_NoAccess: verifying error when HasFileAccess returns false")

	mockRepo := new(MockDownloadRepository)
	mockStatsRepo := new(MockUserStatsRepository)
	mockFileShareRepo := new(MockFileShareRepository)

	uc := NewDownloadCloudRepositoryUseCase(mockRepo, mockStatsRepo, mockFileShareRepo, 5*time.Second)

	ctx := context.Background()
	userID := uint(99)
	fileID := uint(100)

	mockFileShareRepo.On("HasFileAccess", mock.Anything, int32(userID), fileID).Return(false, nil)

	result, err := uc.RequestDownloadURL(ctx, userID, fileID)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "file not found or no access")
	t.Logf("Expected: access denied - GetFileByID should not be called")

	mockFileShareRepo.AssertExpectations(t)
	mockRepo.AssertNotCalled(t, "GetFileByID")
}

func TestRequestDownloadURL_HasFileAccessError(t *testing.T) {
	t.Logf("TestRequestDownloadURL_HasFileAccessError: verifying error when HasFileAccess fails")

	mockRepo := new(MockDownloadRepository)
	mockStatsRepo := new(MockUserStatsRepository)
	mockFileShareRepo := new(MockFileShareRepository)

	uc := NewDownloadCloudRepositoryUseCase(mockRepo, mockStatsRepo, mockFileShareRepo, 5*time.Second)

	ctx := context.Background()
	userID := uint(1)
	fileID := uint(100)

	mockFileShareRepo.On("HasFileAccess", mock.Anything, int32(userID), fileID).Return(false, errors.New("db error"))

	result, err := uc.RequestDownloadURL(ctx, userID, fileID)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to check file access")
	mockFileShareRepo.AssertExpectations(t)
	mockRepo.AssertNotCalled(t, "GetFileByID")
}

func TestRequestDownloadURL_GeneratePresignedURLError(t *testing.T) {
	t.Logf("TestRequestDownloadURL_GeneratePresignedURLError: verifying error when presigned URL generation fails")

	mockRepo := new(MockDownloadRepository)
	mockStatsRepo := new(MockUserStatsRepository)
	mockFileShareRepo := new(MockFileShareRepository)

	uc := NewDownloadCloudRepositoryUseCase(mockRepo, mockStatsRepo, mockFileShareRepo, 5*time.Second)

	ctx := context.Background()
	userID := uint(1)
	fileID := uint(100)

	mockFile := &entity.CloudFile{
		ID:       fileID,
		UserID:   userID,
		FileName: "test.jpg",
		S3Key:    "user1/test.jpg",
		FileType: entity.FileTypeImage,
		FileSize: 1024,
	}

	mockFileShareRepo.On("HasFileAccess", mock.Anything, int32(userID), fileID).Return(true, nil)
	mockRepo.On("GetFileByID", mock.Anything, fileID).Return(mockFile, nil)
	mockStatsRepo.On("LogActivity", mock.Anything, mock.AnythingOfType("*entity.ActivityLog")).Return(nil)
	mockRepo.On("GeneratePresignedDownloadURLWithFilename", mock.Anything, "user1/test.jpg", "test.jpg", 1*time.Hour).
		Return("", errors.New("S3 error"))

	result, err := uc.RequestDownloadURL(ctx, userID, fileID)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to generate download URL")
	mockFileShareRepo.AssertExpectations(t)
	mockRepo.AssertExpectations(t)
}
