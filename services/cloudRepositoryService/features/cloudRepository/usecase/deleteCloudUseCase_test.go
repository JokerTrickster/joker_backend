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

// MockDeleteRepository mocks IDeleteCloudRepositoryRepository
type MockDeleteRepository struct {
	mock.Mock
}

func (m *MockDeleteRepository) DeleteFromS3(ctx context.Context, s3Key string) error {
	args := m.Called(ctx, s3Key)
	return args.Error(0)
}

func (m *MockDeleteRepository) SoftDeleteFile(ctx context.Context, id uint, userID uint) error {
	args := m.Called(ctx, id, userID)
	return args.Error(0)
}

func (m *MockDeleteRepository) GetFileByID(ctx context.Context, id uint) (*entity.CloudFile, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.CloudFile), args.Error(1)
}

func TestDeleteFile_Success(t *testing.T) {
	t.Logf("TestDeleteFile_Success: verifying soft delete and S3 delete on success")

	mockRepo := new(MockDeleteRepository)
	uc := NewDeleteCloudRepositoryUseCase(mockRepo, 5*time.Second)

	ctx := context.Background()
	userID := uint(1)
	fileID := uint(100)

	mockFile := &entity.CloudFile{
		ID:      fileID,
		UserID:  userID,
		S3Key:   "user1/test.jpg",
		FileName: "test.jpg",
	}

	mockRepo.On("GetFileByID", mock.Anything, fileID).Return(mockFile, nil)
	mockRepo.On("SoftDeleteFile", mock.Anything, fileID, userID).Return(nil)
	mockRepo.On("DeleteFromS3", mock.Anything, "user1/test.jpg").Return(nil)

	err := uc.DeleteFile(ctx, userID, fileID)

	assert.NoError(t, err)
	t.Logf("Success: file soft deleted and S3 delete invoked")

	mockRepo.AssertExpectations(t)
}

func TestDeleteFile_FileNotFound(t *testing.T) {
	t.Logf("TestDeleteFile_FileNotFound: verifying error when file does not exist")

	mockRepo := new(MockDeleteRepository)
	uc := NewDeleteCloudRepositoryUseCase(mockRepo, 5*time.Second)

	ctx := context.Background()
	userID := uint(1)
	fileID := uint(999)

	mockRepo.On("GetFileByID", mock.Anything, fileID).Return(nil, errors.New("record not found"))

	err := uc.DeleteFile(ctx, userID, fileID)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "file not found")
	t.Logf("Expected: SoftDeleteFile and DeleteFromS3 should not be called")

	mockRepo.AssertExpectations(t)
	mockRepo.AssertNotCalled(t, "SoftDeleteFile")
	mockRepo.AssertNotCalled(t, "DeleteFromS3")
}

func TestDeleteFile_UnauthorizedWrongUserID(t *testing.T) {
	t.Logf("TestDeleteFile_UnauthorizedWrongUserID: verifying error when user does not own file")

	mockRepo := new(MockDeleteRepository)
	uc := NewDeleteCloudRepositoryUseCase(mockRepo, 5*time.Second)

	ctx := context.Background()
	requesterID := uint(2)
	ownerID := uint(1)
	fileID := uint(100)

	mockFile := &entity.CloudFile{
		ID:      fileID,
		UserID:  ownerID,
		S3Key:   "user1/test.jpg",
	}

	mockRepo.On("GetFileByID", mock.Anything, fileID).Return(mockFile, nil)

	err := uc.DeleteFile(ctx, requesterID, fileID)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unauthorized")
	t.Logf("Expected: SoftDeleteFile and DeleteFromS3 should not be called when unauthorized")

	mockRepo.AssertExpectations(t)
	mockRepo.AssertNotCalled(t, "SoftDeleteFile")
	mockRepo.AssertNotCalled(t, "DeleteFromS3")
}

func TestDeleteFile_S3DeleteFailureStillSucceeds(t *testing.T) {
	t.Logf("TestDeleteFile_S3DeleteFailureStillSucceeds: soft delete succeeds even when S3 delete fails")

	mockRepo := new(MockDeleteRepository)
	uc := NewDeleteCloudRepositoryUseCase(mockRepo, 5*time.Second)

	ctx := context.Background()
	userID := uint(1)
	fileID := uint(100)

	mockFile := &entity.CloudFile{
		ID:      fileID,
		UserID:  userID,
		S3Key:   "user1/test.jpg",
	}

	mockRepo.On("GetFileByID", mock.Anything, fileID).Return(mockFile, nil)
	mockRepo.On("SoftDeleteFile", mock.Anything, fileID, userID).Return(nil)
	mockRepo.On("DeleteFromS3", mock.Anything, "user1/test.jpg").Return(errors.New("S3 connection error"))

	err := uc.DeleteFile(ctx, userID, fileID)

	assert.NoError(t, err)
	t.Logf("Success: UseCase returns nil even when S3 delete fails (error is logged only)")

	mockRepo.AssertExpectations(t)
}
