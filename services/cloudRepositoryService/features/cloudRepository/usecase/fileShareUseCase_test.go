package usecase

import (
	"context"
	"testing"
	"time"

	"github.com/JokerTrickster/joker_backend/services/cloudRepositoryService/features/cloudRepository/model/entity"
	"github.com/JokerTrickster/joker_backend/services/cloudRepositoryService/features/cloudRepository/model/request"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockFileShareRepository mocks IFileShareRepository for file share usecase
type MockFileShareRepositoryForShare struct {
	mock.Mock
}

func (m *MockFileShareRepositoryForShare) CreateFileShare(ctx context.Context, share *entity.FileShare) error {
	args := m.Called(ctx, share)
	return args.Error(0)
}

func (m *MockFileShareRepositoryForShare) GetFileSharesByFileID(ctx context.Context, fileID uint) ([]entity.FileShare, error) {
	args := m.Called(ctx, fileID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]entity.FileShare), args.Error(1)
}

func (m *MockFileShareRepositoryForShare) GetSharedFilesByUserID(ctx context.Context, userID int32) ([]entity.FileShare, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]entity.FileShare), args.Error(1)
}

func (m *MockFileShareRepositoryForShare) GetFilesSharedByUserID(ctx context.Context, ownerID int32) ([]entity.FileShare, error) {
	args := m.Called(ctx, ownerID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]entity.FileShare), args.Error(1)
}

func (m *MockFileShareRepositoryForShare) DeleteFileShare(ctx context.Context, fileID uint, sharedWithID int32, ownerID int32) error {
	args := m.Called(ctx, fileID, sharedWithID, ownerID)
	return args.Error(0)
}

func (m *MockFileShareRepositoryForShare) HasFileAccess(ctx context.Context, userID int32, fileID uint) (bool, error) {
	args := m.Called(ctx, userID, fileID)
	return args.Bool(0), args.Error(1)
}

// MockFolderShareRepositoryForShare mocks IFolderShareRepository (subset for file share)
type MockFolderShareRepositoryForShare struct {
	mock.Mock
}

func (m *MockFolderShareRepositoryForShare) CreateFolderShare(ctx context.Context, share *entity.FolderShare) error {
	args := m.Called(ctx, share)
	return args.Error(0)
}

func (m *MockFolderShareRepositoryForShare) GetFolderSharesByFolderID(ctx context.Context, folderID uint) ([]entity.FolderShare, error) {
	args := m.Called(ctx, folderID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]entity.FolderShare), args.Error(1)
}

func (m *MockFolderShareRepositoryForShare) GetSharedFoldersByUserID(ctx context.Context, userID int32) ([]entity.FolderShare, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]entity.FolderShare), args.Error(1)
}

func (m *MockFolderShareRepositoryForShare) GetFoldersSharedByUserID(ctx context.Context, ownerID int32) ([]entity.FolderShare, error) {
	args := m.Called(ctx, ownerID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]entity.FolderShare), args.Error(1)
}

func (m *MockFolderShareRepositoryForShare) DeleteFolderShare(ctx context.Context, folderID uint, sharedWithID int32, ownerID int32) error {
	args := m.Called(ctx, folderID, sharedWithID, ownerID)
	return args.Error(0)
}

func (m *MockFolderShareRepositoryForShare) HasFolderAccess(ctx context.Context, userID int32, folderID uint) (bool, error) {
	args := m.Called(ctx, userID, folderID)
	return args.Bool(0), args.Error(1)
}

func (m *MockFolderShareRepositoryForShare) HasFolderWritePermission(ctx context.Context, userID int32, folderID uint) (bool, error) {
	args := m.Called(ctx, userID, folderID)
	return args.Bool(0), args.Error(1)
}

func (m *MockFolderShareRepositoryForShare) GetUsersByEmails(ctx context.Context, emails []string) ([]entity.User, error) {
	args := m.Called(ctx, emails)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]entity.User), args.Error(1)
}

func TestShareFile_Success(t *testing.T) {
	t.Logf("TestShareFile_Success: verifying ShareFile creates share for valid user")

	mockFileShare := new(MockFileShareRepositoryForShare)
	mockFolderShare := new(MockFolderShareRepositoryForShare)
	mockS3 := new(MockListRepository)

	uc := NewFileShareUseCase(mockFileShare, mockFolderShare, mockS3, 5*time.Second)

	ctx := context.Background()
	fileID := uint(100)
	ownerID := int32(1)
	req := &request.ShareFileRequestDTO{UserEmails: []string{"shared@example.com"}, Permission: "read"}

	mockFolderShare.On("GetUsersByEmails", mock.Anything, []string{"shared@example.com"}).
		Return([]entity.User{{ID: 2, Email: "shared@example.com", Name: "Shared User"}}, nil)
	mockFileShare.On("HasFileAccess", mock.Anything, ownerID, fileID).Return(true, nil)
	mockFileShare.On("GetFileSharesByFileID", mock.Anything, fileID).Return([]entity.FileShare{}, nil)
	mockFileShare.On("CreateFileShare", mock.Anything, mock.AnythingOfType("*entity.FileShare")).Return(nil)

	result, err := uc.ShareFile(ctx, fileID, ownerID, req)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result.SharedUsers, 1)
	assert.Equal(t, "shared@example.com", result.SharedUsers[0].Email)
	t.Logf("Success: file shared with user")

	mockFileShare.AssertExpectations(t)
	mockFolderShare.AssertExpectations(t)
}

func TestShareFile_NoAccess(t *testing.T) {
	t.Logf("TestShareFile_NoAccess: verifying error when owner has no access to file")

	mockFileShare := new(MockFileShareRepositoryForShare)
	mockFolderShare := new(MockFolderShareRepositoryForShare)
	mockS3 := new(MockListRepository)

	uc := NewFileShareUseCase(mockFileShare, mockFolderShare, mockS3, 5*time.Second)

	ctx := context.Background()
	fileID := uint(100)
	ownerID := int32(1)
	req := &request.ShareFileRequestDTO{UserEmails: []string{"user@example.com"}}

	mockFileShare.On("HasFileAccess", mock.Anything, ownerID, fileID).Return(false, nil)

	result, err := uc.ShareFile(ctx, fileID, ownerID, req)

	assert.Error(t, err)
	assert.Nil(t, result)
	t.Logf("Expected: GetUsersByEmails should not be called when no access")

	mockFileShare.AssertExpectations(t)
	mockFolderShare.AssertNotCalled(t, "GetUsersByEmails")
}

func TestGetFileShares_Success(t *testing.T) {
	t.Logf("TestGetFileShares_Success: verifying GetFileShares returns share list")

	mockFileShare := new(MockFileShareRepositoryForShare)
	mockFolderShare := new(MockFolderShareRepositoryForShare)
	mockS3 := new(MockListRepository)

	uc := NewFileShareUseCase(mockFileShare, mockFolderShare, mockS3, 5*time.Second)

	ctx := context.Background()
	fileID := uint(100)
	ownerID := int32(1)

	mockShares := []entity.FileShare{
		{
			FileID:       fileID,
			OwnerID:      ownerID,
			SharedWithID: 2,
			SharedWith:   &entity.User{ID: 2, Email: "user@ex.com", Name: "User"},
			File:         &entity.CloudFile{FileName: "doc.pdf"},
		},
	}

	mockFileShare.On("HasFileAccess", mock.Anything, ownerID, fileID).Return(true, nil)
	mockFileShare.On("GetFileSharesByFileID", mock.Anything, fileID).Return(mockShares, nil)

	result, err := uc.GetFileShares(ctx, fileID, ownerID)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, fileID, result.FileID)
	assert.Len(t, result.SharedUsers, 1)
	t.Logf("Success: file shares retrieved")

	mockFileShare.AssertExpectations(t)
}

func TestRevokeFileShare_Success(t *testing.T) {
	t.Logf("TestRevokeFileShare_Success: verifying RevokeFileShare deletes share")

	mockFileShare := new(MockFileShareRepositoryForShare)
	mockFolderShare := new(MockFolderShareRepositoryForShare)
	mockS3 := new(MockListRepository)

	uc := NewFileShareUseCase(mockFileShare, mockFolderShare, mockS3, 5*time.Second)

	ctx := context.Background()
	fileID := uint(100)
	sharedWithID := int32(2)
	ownerID := int32(1)

	mockFileShare.On("HasFileAccess", mock.Anything, ownerID, fileID).Return(true, nil)
	mockFileShare.On("DeleteFileShare", mock.Anything, fileID, sharedWithID, ownerID).Return(nil)

	result, err := uc.RevokeFileShare(ctx, fileID, sharedWithID, ownerID)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	t.Logf("Success: file share revoked")

	mockFileShare.AssertExpectations(t)
}

func TestGetSharedWithMeFiles_Success(t *testing.T) {
	t.Skip("Skipping: GetSharedWithMeFiles has complex S3/file structure - integration test preferred")
}

func TestGetFilesSharedByMe_Success(t *testing.T) {
	t.Skip("Skipping: GetFilesSharedByMe has complex grouping - integration test preferred")
}
