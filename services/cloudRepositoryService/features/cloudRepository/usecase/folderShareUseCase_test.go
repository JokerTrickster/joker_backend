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

// MockFolderRepositoryForShare mocks IFolderRepository for folder share usecase
type MockFolderRepositoryForShare struct {
	mock.Mock
}

func (m *MockFolderRepositoryForShare) CreateFolder(ctx context.Context, folder *entity.Folder) error {
	args := m.Called(ctx, folder)
	return args.Error(0)
}

func (m *MockFolderRepositoryForShare) GetFolderByID(ctx context.Context, id uint, userID int32) (*entity.Folder, error) {
	args := m.Called(ctx, id, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Folder), args.Error(1)
}

func (m *MockFolderRepositoryForShare) GetFolderByIDWithoutUserCheck(ctx context.Context, id uint) (*entity.Folder, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Folder), args.Error(1)
}

func (m *MockFolderRepositoryForShare) GetFoldersByUserID(ctx context.Context, userID int32) ([]entity.Folder, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]entity.Folder), args.Error(1)
}

func (m *MockFolderRepositoryForShare) UpdateFolder(ctx context.Context, folder *entity.Folder) error {
	args := m.Called(ctx, folder)
	return args.Error(0)
}

func (m *MockFolderRepositoryForShare) DeleteFolder(ctx context.Context, id uint, userID int32) error {
	args := m.Called(ctx, id, userID)
	return args.Error(0)
}

func (m *MockFolderRepositoryForShare) GetFolderFileCount(ctx context.Context, folderID uint, userID int32) (int, error) {
	args := m.Called(ctx, folderID, userID)
	return args.Int(0), args.Error(1)
}

func (m *MockFolderRepositoryForShare) GetFilesByFolderID(ctx context.Context, folderID *uint, userID int32) ([]entity.CloudFile, error) {
	args := m.Called(ctx, folderID, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]entity.CloudFile), args.Error(1)
}

func (m *MockFolderRepositoryForShare) GetFilesByFolderIDWithoutUserCheck(ctx context.Context, folderID *uint) ([]entity.CloudFile, error) {
	args := m.Called(ctx, folderID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]entity.CloudFile), args.Error(1)
}

func (m *MockFolderRepositoryForShare) MoveFilesToFolder(ctx context.Context, fileIDs []uint, folderID *uint, userID int32) (int, error) {
	args := m.Called(ctx, fileIDs, folderID, userID)
	return args.Int(0), args.Error(1)
}

// MockFolderShareRepositoryForFolderShare mocks IFolderShareRepository
type MockFolderShareRepositoryForFolderShare struct {
	mock.Mock
}

func (m *MockFolderShareRepositoryForFolderShare) CreateFolderShare(ctx context.Context, share *entity.FolderShare) error {
	args := m.Called(ctx, share)
	return args.Error(0)
}

func (m *MockFolderShareRepositoryForFolderShare) GetFolderSharesByFolderID(ctx context.Context, folderID uint) ([]entity.FolderShare, error) {
	args := m.Called(ctx, folderID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]entity.FolderShare), args.Error(1)
}

func (m *MockFolderShareRepositoryForFolderShare) GetSharedFoldersByUserID(ctx context.Context, userID int32) ([]entity.FolderShare, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]entity.FolderShare), args.Error(1)
}

func (m *MockFolderShareRepositoryForFolderShare) GetFoldersSharedByUserID(ctx context.Context, ownerID int32) ([]entity.FolderShare, error) {
	args := m.Called(ctx, ownerID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]entity.FolderShare), args.Error(1)
}

func (m *MockFolderShareRepositoryForFolderShare) DeleteFolderShare(ctx context.Context, folderID uint, sharedWithID int32, ownerID int32) error {
	args := m.Called(ctx, folderID, sharedWithID, ownerID)
	return args.Error(0)
}

func (m *MockFolderShareRepositoryForFolderShare) HasFolderAccess(ctx context.Context, userID int32, folderID uint) (bool, error) {
	args := m.Called(ctx, userID, folderID)
	return args.Bool(0), args.Error(1)
}

func (m *MockFolderShareRepositoryForFolderShare) HasFolderWritePermission(ctx context.Context, userID int32, folderID uint) (bool, error) {
	args := m.Called(ctx, userID, folderID)
	return args.Bool(0), args.Error(1)
}

func (m *MockFolderShareRepositoryForFolderShare) GetUsersByEmails(ctx context.Context, emails []string) ([]entity.User, error) {
	args := m.Called(ctx, emails)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]entity.User), args.Error(1)
}

func TestShareFolder_Success(t *testing.T) {
	t.Logf("TestShareFolder_Success: verifying ShareFolder creates share for valid user")

	mockFolderRepo := new(MockFolderRepositoryForShare)
	mockFolderShareRepo := new(MockFolderShareRepositoryForFolderShare)

	uc := NewFolderShareUseCase(mockFolderRepo, mockFolderShareRepo, 5*time.Second)

	ctx := context.Background()
	folderID := uint(100)
	ownerID := int32(1)
	req := &request.ShareFolderRequestDTO{UserEmails: []string{"shared@example.com"}, Permission: "read"}

	mockFolder := &entity.Folder{ID: folderID, UserID: uint(ownerID), FolderName: "My Folder"}
	mockFolderRepo.On("GetFolderByID", mock.Anything, folderID, ownerID).Return(mockFolder, nil)
	mockFolderShareRepo.On("GetUsersByEmails", mock.Anything, []string{"shared@example.com"}).
		Return([]entity.User{{ID: 2, Email: "shared@example.com", Name: "Shared User"}}, nil)
	mockFolderShareRepo.On("GetFolderSharesByFolderID", mock.Anything, folderID).Return([]entity.FolderShare{}, nil)
	mockFolderShareRepo.On("CreateFolderShare", mock.Anything, mock.AnythingOfType("*entity.FolderShare")).Return(nil)

	result, err := uc.ShareFolder(ctx, folderID, ownerID, req)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result.SharedUsers, 1)
	t.Logf("Success: folder shared")

	mockFolderRepo.AssertExpectations(t)
	mockFolderShareRepo.AssertExpectations(t)
}

func TestShareFolder_FolderNotFound(t *testing.T) {
	t.Logf("TestShareFolder_FolderNotFound: verifying error when folder not found")

	mockFolderRepo := new(MockFolderRepositoryForShare)
	mockFolderShareRepo := new(MockFolderShareRepositoryForFolderShare)

	uc := NewFolderShareUseCase(mockFolderRepo, mockFolderShareRepo, 5*time.Second)

	ctx := context.Background()
	folderID := uint(999)
	ownerID := int32(1)
	req := &request.ShareFolderRequestDTO{UserEmails: []string{"user@example.com"}}

	mockFolderRepo.On("GetFolderByID", mock.Anything, folderID, ownerID).Return(nil, errors.New("not found"))

	result, err := uc.ShareFolder(ctx, folderID, ownerID, req)

	assert.Error(t, err)
	assert.Nil(t, result)
	t.Logf("Expected: GetUsersByEmails should not be called")

	mockFolderShareRepo.AssertNotCalled(t, "GetUsersByEmails")
}

func TestGetFolderShares_Success(t *testing.T) {
	t.Logf("TestGetFolderShares_Success: verifying GetFolderShares returns share list")

	mockFolderRepo := new(MockFolderRepositoryForShare)
	mockFolderShareRepo := new(MockFolderShareRepositoryForFolderShare)

	uc := NewFolderShareUseCase(mockFolderRepo, mockFolderShareRepo, 5*time.Second)

	ctx := context.Background()
	folderID := uint(100)
	ownerID := int32(1)

	mockFolder := &entity.Folder{ID: folderID, FolderName: "Shared"}
	mockShares := []entity.FolderShare{
		{FolderID: folderID, OwnerID: ownerID, SharedWithID: 2, SharedWith: &entity.User{ID: 2, Email: "u@ex.com", Name: "U"}},
	}

	mockFolderRepo.On("GetFolderByID", mock.Anything, folderID, ownerID).Return(mockFolder, nil)
	mockFolderShareRepo.On("GetFolderSharesByFolderID", mock.Anything, folderID).Return(mockShares, nil)

	result, err := uc.GetFolderShares(ctx, folderID, ownerID)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, folderID, result.FolderID)
	assert.Len(t, result.SharedUsers, 1)
	t.Logf("Success: folder shares retrieved")

	mockFolderRepo.AssertExpectations(t)
	mockFolderShareRepo.AssertExpectations(t)
}

func TestRevokeFolderShare_Success(t *testing.T) {
	t.Logf("TestRevokeFolderShare_Success: verifying RevokeFolderShare deletes share")

	mockFolderRepo := new(MockFolderRepositoryForShare)
	mockFolderShareRepo := new(MockFolderShareRepositoryForFolderShare)

	uc := NewFolderShareUseCase(mockFolderRepo, mockFolderShareRepo, 5*time.Second)

	ctx := context.Background()
	folderID := uint(100)
	sharedWithID := int32(2)
	ownerID := int32(1)

	mockFolder := &entity.Folder{ID: folderID}
	mockFolderRepo.On("GetFolderByID", mock.Anything, folderID, ownerID).Return(mockFolder, nil)
	mockFolderShareRepo.On("DeleteFolderShare", mock.Anything, folderID, sharedWithID, ownerID).Return(nil)

	result, err := uc.RevokeFolderShare(ctx, folderID, sharedWithID, ownerID)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	t.Logf("Success: folder share revoked")

	mockFolderRepo.AssertExpectations(t)
	mockFolderShareRepo.AssertExpectations(t)
}

func TestGetSharedWithMeFolders_Success(t *testing.T) {
	t.Skip("Skipping: GetSharedWithMeFolders has complex structure - integration test preferred")
}

func TestGetFoldersSharedByMe_Success(t *testing.T) {
	t.Skip("Skipping: GetFoldersSharedByMe has complex grouping - integration test preferred")
}
