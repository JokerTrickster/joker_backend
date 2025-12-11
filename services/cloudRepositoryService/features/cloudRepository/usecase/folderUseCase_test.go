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

// MockFolderRepository is a mock implementation of IFolderRepository
type MockFolderRepository struct {
	mock.Mock
}

func (m *MockFolderRepository) CreateFolder(ctx context.Context, folder *entity.Folder) error {
	args := m.Called(ctx, folder)
	return args.Error(0)
}

func (m *MockFolderRepository) GetFolderByID(ctx context.Context, id uint, userID int32) (*entity.Folder, error) {
	args := m.Called(ctx, id, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Folder), args.Error(1)
}

func (m *MockFolderRepository) GetFolderByIDWithoutUserCheck(ctx context.Context, id uint) (*entity.Folder, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Folder), args.Error(1)
}

func (m *MockFolderRepository) GetFoldersByUserID(ctx context.Context, userID int32) ([]entity.Folder, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]entity.Folder), args.Error(1)
}

func (m *MockFolderRepository) UpdateFolder(ctx context.Context, folder *entity.Folder) error {
	args := m.Called(ctx, folder)
	return args.Error(0)
}

func (m *MockFolderRepository) DeleteFolder(ctx context.Context, id uint, userID int32) error {
	args := m.Called(ctx, id, userID)
	return args.Error(0)
}

func (m *MockFolderRepository) GetFolderFileCount(ctx context.Context, folderID uint, userID int32) (int, error) {
	args := m.Called(ctx, folderID, userID)
	return args.Int(0), args.Error(1)
}

func (m *MockFolderRepository) GetFilesByFolderID(ctx context.Context, folderID *uint, userID int32) ([]entity.CloudFile, error) {
	args := m.Called(ctx, folderID, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]entity.CloudFile), args.Error(1)
}

func (m *MockFolderRepository) GetFilesByFolderIDWithoutUserCheck(ctx context.Context, folderID *uint) ([]entity.CloudFile, error) {
	args := m.Called(ctx, folderID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]entity.CloudFile), args.Error(1)
}

func (m *MockFolderRepository) MoveFilesToFolder(ctx context.Context, fileIDs []uint, folderID *uint, userID int32) (int, error) {
	args := m.Called(ctx, fileIDs, folderID, userID)
	return args.Int(0), args.Error(1)
}

// MockFolderShareRepository is a mock implementation of IFolderShareRepository
type MockFolderShareRepository struct {
	mock.Mock
}

func (m *MockFolderShareRepository) CreateFolderShare(ctx context.Context, share *entity.FolderShare) error {
	args := m.Called(ctx, share)
	return args.Error(0)
}

func (m *MockFolderShareRepository) GetFolderSharesByFolderID(ctx context.Context, folderID uint) ([]entity.FolderShare, error) {
	args := m.Called(ctx, folderID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]entity.FolderShare), args.Error(1)
}

func (m *MockFolderShareRepository) GetSharedFoldersByUserID(ctx context.Context, userID int32) ([]entity.FolderShare, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]entity.FolderShare), args.Error(1)
}

func (m *MockFolderShareRepository) GetFoldersSharedByUserID(ctx context.Context, ownerID int32) ([]entity.FolderShare, error) {
	args := m.Called(ctx, ownerID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]entity.FolderShare), args.Error(1)
}

func (m *MockFolderShareRepository) DeleteFolderShare(ctx context.Context, folderID uint, sharedWithID int32, ownerID int32) error {
	args := m.Called(ctx, folderID, sharedWithID, ownerID)
	return args.Error(0)
}

func (m *MockFolderShareRepository) HasFolderAccess(ctx context.Context, userID int32, folderID uint) (bool, error) {
	args := m.Called(ctx, userID, folderID)
	return args.Bool(0), args.Error(1)
}

func (m *MockFolderShareRepository) HasFolderWritePermission(ctx context.Context, userID int32, folderID uint) (bool, error) {
	args := m.Called(ctx, userID, folderID)
	return args.Bool(0), args.Error(1)
}

func (m *MockFolderShareRepository) GetUsersByEmails(ctx context.Context, emails []string) ([]entity.User, error) {
	args := m.Called(ctx, emails)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]entity.User), args.Error(1)
}

// MockS3Repository is a mock implementation of IListCloudRepositoryRepository
type MockS3Repository struct {
	mock.Mock
}

func (m *MockS3Repository) GetFilesByUserID(ctx context.Context, userID uint, filter request.ListFilesRequestDTO) ([]entity.CloudFile, int64, error) {
	args := m.Called(ctx, userID, filter)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]entity.CloudFile), args.Get(1).(int64), args.Error(2)
}

func (m *MockS3Repository) GeneratePresignedDownloadURL(ctx context.Context, s3Key string, expiration time.Duration) (string, error) {
	args := m.Called(ctx, s3Key, expiration)
	return args.String(0), args.Error(1)
}

// TestGetFolderFiles_OwnerAccess tests that folder owner can retrieve files from their own folder
func TestGetFolderFiles_OwnerAccess(t *testing.T) {
	mockFolderRepo := new(MockFolderRepository)
	mockFolderShareRepo := new(MockFolderShareRepository)
	mockS3Repo := new(MockS3Repository)

	uc := NewFolderUseCase(mockFolderRepo, mockFolderShareRepo, mockS3Repo, 5*time.Second)

	ctx := context.Background()
	ownerUserID := uint(1)
	folderID := uint(100)

	// Mock folder owned by user 1
	mockFolder := &entity.Folder{
		ID:         folderID,
		UserID:     ownerUserID,
		FolderName: "My Folder",
	}

	// Mock files in the folder
	mockFiles := []entity.CloudFile{
		{
			ID:          1,
			UserID:      ownerUserID,
			FileName:    "file1.txt",
			FileType:    entity.FileTypeImage,
			ContentType: "text/plain",
			FileSize:    1024,
			S3Key:       "user1/file1.txt",
		},
		{
			ID:          2,
			UserID:      ownerUserID,
			FileName:    "file2.jpg",
			FileType:    entity.FileTypeImage,
			ContentType: "image/jpeg",
			FileSize:    2048,
			S3Key:       "user1/file2.jpg",
		},
	}

	// Setup expectations
	mockFolderShareRepo.On("HasFolderAccess", mock.Anything, int32(ownerUserID), folderID).Return(true, nil)
	mockFolderRepo.On("GetFolderByIDWithoutUserCheck", mock.Anything, folderID).Return(mockFolder, nil)
	mockFolderRepo.On("GetFilesByFolderID", mock.Anything, &folderID, int32(ownerUserID)).Return(mockFiles, nil)
	mockS3Repo.On("GeneratePresignedDownloadURL", mock.Anything, "user1/file1.txt", 1*time.Hour).Return("https://s3.url/file1", nil)
	mockS3Repo.On("GeneratePresignedDownloadURL", mock.Anything, "user1/file2.jpg", 1*time.Hour).Return("https://s3.url/file2", nil)

	// Execute
	files, err := uc.GetFolderFiles(ctx, folderID, ownerUserID)

	// Assert
	assert.NoError(t, err)
	assert.Len(t, files, 2)
	assert.Equal(t, "file1.txt", files[0].FileName)
	assert.Equal(t, "file2.jpg", files[1].FileName)
	assert.Equal(t, "https://s3.url/file1", files[0].DownloadURL)
	assert.Equal(t, "https://s3.url/file2", files[1].DownloadURL)

	mockFolderShareRepo.AssertExpectations(t)
	mockFolderRepo.AssertExpectations(t)
	mockS3Repo.AssertExpectations(t)
}

// TestGetFolderFiles_SharedUserReadAccess tests that a user with read permission can retrieve files
func TestGetFolderFiles_SharedUserReadAccess(t *testing.T) {
	mockFolderRepo := new(MockFolderRepository)
	mockFolderShareRepo := new(MockFolderShareRepository)
	mockS3Repo := new(MockS3Repository)

	uc := NewFolderUseCase(mockFolderRepo, mockFolderShareRepo, mockS3Repo, 5*time.Second)

	ctx := context.Background()
	ownerUserID := uint(1)
	sharedUserID := uint(2)
	folderID := uint(100)

	// Mock folder owned by user 1, shared with user 2
	mockFolder := &entity.Folder{
		ID:         folderID,
		UserID:     ownerUserID,
		FolderName: "Shared Folder",
	}

	// Mock files in the folder - IMPORTANT: Files belong to owner (user 1)
	mockFiles := []entity.CloudFile{
		{
			ID:          1,
			UserID:      ownerUserID, // File owned by user 1
			FileName:    "shared_file1.txt",
			FileType:    entity.FileTypeImage,
			ContentType: "text/plain",
			FileSize:    1024,
			S3Key:       "user1/shared_file1.txt",
		},
		{
			ID:          2,
			UserID:      ownerUserID, // File owned by user 1
			FileName:    "shared_file2.pdf",
			FileType:    entity.FileTypeImage,
			ContentType: "application/pdf",
			FileSize:    3072,
			S3Key:       "user1/shared_file2.pdf",
		},
	}

	// Setup expectations - User 2 accessing user 1's folder
	mockFolderShareRepo.On("HasFolderAccess", mock.Anything, int32(sharedUserID), folderID).Return(true, nil)
	mockFolderRepo.On("GetFolderByIDWithoutUserCheck", mock.Anything, folderID).Return(mockFolder, nil)
	// CRITICAL: GetFilesByFolderID should be called with OWNER's ID (user 1), not requester's ID (user 2)
	mockFolderRepo.On("GetFilesByFolderID", mock.Anything, &folderID, int32(ownerUserID)).Return(mockFiles, nil)
	mockS3Repo.On("GeneratePresignedDownloadURL", mock.Anything, "user1/shared_file1.txt", 1*time.Hour).Return("https://s3.url/shared1", nil)
	mockS3Repo.On("GeneratePresignedDownloadURL", mock.Anything, "user1/shared_file2.pdf", 1*time.Hour).Return("https://s3.url/shared2", nil)

	// Execute - User 2 requests files from user 1's folder
	files, err := uc.GetFolderFiles(ctx, folderID, sharedUserID)

	// Assert
	assert.NoError(t, err)
	assert.Len(t, files, 2)
	assert.Equal(t, "shared_file1.txt", files[0].FileName)
	assert.Equal(t, "shared_file2.pdf", files[1].FileName)
	assert.Equal(t, "https://s3.url/shared1", files[0].DownloadURL)
	assert.Equal(t, "https://s3.url/shared2", files[1].DownloadURL)

	mockFolderShareRepo.AssertExpectations(t)
	mockFolderRepo.AssertExpectations(t)
	mockS3Repo.AssertExpectations(t)
}

// TestGetFolderFiles_SharedUserWriteAccess tests that a user with write permission can retrieve files
func TestGetFolderFiles_SharedUserWriteAccess(t *testing.T) {
	mockFolderRepo := new(MockFolderRepository)
	mockFolderShareRepo := new(MockFolderShareRepository)
	mockS3Repo := new(MockS3Repository)

	uc := NewFolderUseCase(mockFolderRepo, mockFolderShareRepo, mockS3Repo, 5*time.Second)

	ctx := context.Background()
	ownerUserID := uint(1)
	sharedUserID := uint(3)
	folderID := uint(100)

	// Mock folder owned by user 1, shared with user 3 with write permission
	mockFolder := &entity.Folder{
		ID:         folderID,
		UserID:     ownerUserID,
		FolderName: "Shared Folder with Write Access",
	}

	// Mock files in the folder
	mockFiles := []entity.CloudFile{
		{
			ID:          1,
			UserID:      ownerUserID,
			FileName:    "writable_file.docx",
			FileType:    entity.FileTypeImage,
			ContentType: "application/vnd.openxmlformats-officedocument.wordprocessingml.document",
			FileSize:    5120,
			S3Key:       "user1/writable_file.docx",
		},
	}

	// Setup expectations
	mockFolderShareRepo.On("HasFolderAccess", mock.Anything, int32(sharedUserID), folderID).Return(true, nil)
	mockFolderRepo.On("GetFolderByIDWithoutUserCheck", mock.Anything, folderID).Return(mockFolder, nil)
	mockFolderRepo.On("GetFilesByFolderID", mock.Anything, &folderID, int32(ownerUserID)).Return(mockFiles, nil)
	mockS3Repo.On("GeneratePresignedDownloadURL", mock.Anything, "user1/writable_file.docx", 1*time.Hour).Return("https://s3.url/writable", nil)

	// Execute
	files, err := uc.GetFolderFiles(ctx, folderID, sharedUserID)

	// Assert
	assert.NoError(t, err)
	assert.Len(t, files, 1)
	assert.Equal(t, "writable_file.docx", files[0].FileName)
	assert.Equal(t, "https://s3.url/writable", files[0].DownloadURL)

	mockFolderShareRepo.AssertExpectations(t)
	mockFolderRepo.AssertExpectations(t)
	mockS3Repo.AssertExpectations(t)
}

// TestGetFolderFiles_NoAccess tests that a user without permission cannot retrieve files
func TestGetFolderFiles_NoAccess(t *testing.T) {
	mockFolderRepo := new(MockFolderRepository)
	mockFolderShareRepo := new(MockFolderShareRepository)
	mockS3Repo := new(MockS3Repository)

	uc := NewFolderUseCase(mockFolderRepo, mockFolderShareRepo, mockS3Repo, 5*time.Second)

	ctx := context.Background()
	unauthorizedUserID := uint(999)
	folderID := uint(100)

	// Setup expectations - User has no access
	mockFolderShareRepo.On("HasFolderAccess", mock.Anything, int32(unauthorizedUserID), folderID).Return(false, nil)

	// Execute
	files, err := uc.GetFolderFiles(ctx, folderID, unauthorizedUserID)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, files)
	assert.Contains(t, err.Error(), "folder not found or no access")

	mockFolderShareRepo.AssertExpectations(t)
	// Repository methods should NOT be called after access check fails
	mockFolderRepo.AssertNotCalled(t, "GetFolderByIDWithoutUserCheck")
	mockFolderRepo.AssertNotCalled(t, "GetFilesByFolderID")
}

// TestGetFolderFiles_AccessCheckError tests error handling when access check fails
func TestGetFolderFiles_AccessCheckError(t *testing.T) {
	mockFolderRepo := new(MockFolderRepository)
	mockFolderShareRepo := new(MockFolderShareRepository)
	mockS3Repo := new(MockS3Repository)

	uc := NewFolderUseCase(mockFolderRepo, mockFolderShareRepo, mockS3Repo, 5*time.Second)

	ctx := context.Background()
	userID := uint(1)
	folderID := uint(100)

	// Setup expectations - Access check returns error
	mockFolderShareRepo.On("HasFolderAccess", mock.Anything, int32(userID), folderID).Return(false, errors.New("database connection error"))

	// Execute
	files, err := uc.GetFolderFiles(ctx, folderID, userID)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, files)
	assert.Contains(t, err.Error(), "failed to check folder access")

	mockFolderShareRepo.AssertExpectations(t)
}

// TestGetFolderFiles_FolderNotFound tests error handling when folder doesn't exist
func TestGetFolderFiles_FolderNotFound(t *testing.T) {
	mockFolderRepo := new(MockFolderRepository)
	mockFolderShareRepo := new(MockFolderShareRepository)
	mockS3Repo := new(MockS3Repository)

	uc := NewFolderUseCase(mockFolderRepo, mockFolderShareRepo, mockS3Repo, 5*time.Second)

	ctx := context.Background()
	userID := uint(1)
	nonExistentFolderID := uint(999)

	// Setup expectations
	mockFolderShareRepo.On("HasFolderAccess", mock.Anything, int32(userID), nonExistentFolderID).Return(true, nil)
	mockFolderRepo.On("GetFolderByIDWithoutUserCheck", mock.Anything, nonExistentFolderID).Return(nil, errors.New("folder not found"))

	// Execute
	files, err := uc.GetFolderFiles(ctx, nonExistentFolderID, userID)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, files)
	assert.Contains(t, err.Error(), "folder not found")

	mockFolderShareRepo.AssertExpectations(t)
	mockFolderRepo.AssertExpectations(t)
}

// TestGetFolderFiles_GetFilesError tests error handling when file retrieval fails
func TestGetFolderFiles_GetFilesError(t *testing.T) {
	mockFolderRepo := new(MockFolderRepository)
	mockFolderShareRepo := new(MockFolderShareRepository)
	mockS3Repo := new(MockS3Repository)

	uc := NewFolderUseCase(mockFolderRepo, mockFolderShareRepo, mockS3Repo, 5*time.Second)

	ctx := context.Background()
	userID := uint(1)
	folderID := uint(100)

	mockFolder := &entity.Folder{
		ID:         folderID,
		UserID:     userID,
		FolderName: "Test Folder",
	}

	// Setup expectations
	mockFolderShareRepo.On("HasFolderAccess", mock.Anything, int32(userID), folderID).Return(true, nil)
	mockFolderRepo.On("GetFolderByIDWithoutUserCheck", mock.Anything, folderID).Return(mockFolder, nil)
	mockFolderRepo.On("GetFilesByFolderID", mock.Anything, &folderID, int32(userID)).Return(nil, errors.New("database query failed"))

	// Execute
	files, err := uc.GetFolderFiles(ctx, folderID, userID)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, files)
	assert.Contains(t, err.Error(), "failed to get folder files")

	mockFolderShareRepo.AssertExpectations(t)
	mockFolderRepo.AssertExpectations(t)
}

// TestGetFolderFiles_EmptyFolder tests retrieval from empty folder
func TestGetFolderFiles_EmptyFolder(t *testing.T) {
	mockFolderRepo := new(MockFolderRepository)
	mockFolderShareRepo := new(MockFolderShareRepository)
	mockS3Repo := new(MockS3Repository)

	uc := NewFolderUseCase(mockFolderRepo, mockFolderShareRepo, mockS3Repo, 5*time.Second)

	ctx := context.Background()
	userID := uint(1)
	folderID := uint(100)

	mockFolder := &entity.Folder{
		ID:         folderID,
		UserID:     userID,
		FolderName: "Empty Folder",
	}

	// Setup expectations - no files in folder
	mockFolderShareRepo.On("HasFolderAccess", mock.Anything, int32(userID), folderID).Return(true, nil)
	mockFolderRepo.On("GetFolderByIDWithoutUserCheck", mock.Anything, folderID).Return(mockFolder, nil)
	mockFolderRepo.On("GetFilesByFolderID", mock.Anything, &folderID, int32(userID)).Return([]entity.CloudFile{}, nil)

	// Execute
	files, err := uc.GetFolderFiles(ctx, folderID, userID)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, files)
	assert.Len(t, files, 0)

	mockFolderShareRepo.AssertExpectations(t)
	mockFolderRepo.AssertExpectations(t)
}

// TestGetFolderFiles_WithTags tests file retrieval with tags
func TestGetFolderFiles_WithTags(t *testing.T) {
	mockFolderRepo := new(MockFolderRepository)
	mockFolderShareRepo := new(MockFolderShareRepository)
	mockS3Repo := new(MockS3Repository)

	uc := NewFolderUseCase(mockFolderRepo, mockFolderShareRepo, mockS3Repo, 5*time.Second)

	ctx := context.Background()
	userID := uint(1)
	folderID := uint(100)

	mockFolder := &entity.Folder{
		ID:         folderID,
		UserID:     userID,
		FolderName: "Tagged Files Folder",
	}

	mockFiles := []entity.CloudFile{
		{
			ID:          1,
			UserID:      userID,
			FileName:    "tagged_file.txt",
			FileType:    entity.FileTypeImage,
			ContentType: "text/plain",
			FileSize:    1024,
			S3Key:       "user1/tagged_file.txt",
			Tags: []entity.Tag{
				{ID: 1, Name: "important"},
				{ID: 2, Name: "work"},
			},
		},
	}

	// Setup expectations
	mockFolderShareRepo.On("HasFolderAccess", mock.Anything, int32(userID), folderID).Return(true, nil)
	mockFolderRepo.On("GetFolderByIDWithoutUserCheck", mock.Anything, folderID).Return(mockFolder, nil)
	mockFolderRepo.On("GetFilesByFolderID", mock.Anything, &folderID, int32(userID)).Return(mockFiles, nil)
	mockS3Repo.On("GeneratePresignedDownloadURL", mock.Anything, "user1/tagged_file.txt", 1*time.Hour).Return("https://s3.url/tagged", nil)

	// Execute
	files, err := uc.GetFolderFiles(ctx, folderID, userID)

	// Assert
	assert.NoError(t, err)
	assert.Len(t, files, 1)
	assert.Equal(t, "tagged_file.txt", files[0].FileName)
	assert.Len(t, files[0].Tags, 2)
	assert.Equal(t, "important", files[0].Tags[0].Name)
	assert.Equal(t, "work", files[0].Tags[1].Name)

	mockFolderShareRepo.AssertExpectations(t)
	mockFolderRepo.AssertExpectations(t)
	mockS3Repo.AssertExpectations(t)
}

// TestGetFolderFiles_PresignedURLGenerationError tests handling of presigned URL generation failures
func TestGetFolderFiles_PresignedURLGenerationError(t *testing.T) {
	mockFolderRepo := new(MockFolderRepository)
	mockFolderShareRepo := new(MockFolderShareRepository)
	mockS3Repo := new(MockS3Repository)

	uc := NewFolderUseCase(mockFolderRepo, mockFolderShareRepo, mockS3Repo, 5*time.Second)

	ctx := context.Background()
	userID := uint(1)
	folderID := uint(100)

	mockFolder := &entity.Folder{
		ID:         folderID,
		UserID:     userID,
		FolderName: "Test Folder",
	}

	mockFiles := []entity.CloudFile{
		{
			ID:          1,
			UserID:      userID,
			FileName:    "file.txt",
			FileType:    entity.FileTypeImage,
			ContentType: "text/plain",
			FileSize:    1024,
			S3Key:       "user1/file.txt",
		},
	}

	// Setup expectations - presigned URL generation fails
	mockFolderShareRepo.On("HasFolderAccess", mock.Anything, int32(userID), folderID).Return(true, nil)
	mockFolderRepo.On("GetFolderByIDWithoutUserCheck", mock.Anything, folderID).Return(mockFolder, nil)
	mockFolderRepo.On("GetFilesByFolderID", mock.Anything, &folderID, int32(userID)).Return(mockFiles, nil)
	mockS3Repo.On("GeneratePresignedDownloadURL", mock.Anything, "user1/file.txt", 1*time.Hour).Return("", errors.New("S3 error"))

	// Execute - should still succeed but with empty URL
	files, err := uc.GetFolderFiles(ctx, folderID, userID)

	// Assert
	assert.NoError(t, err)
	assert.Len(t, files, 1)
	assert.Equal(t, "file.txt", files[0].FileName)
	assert.Equal(t, "", files[0].DownloadURL) // Empty URL on error

	mockFolderShareRepo.AssertExpectations(t)
	mockFolderRepo.AssertExpectations(t)
	mockS3Repo.AssertExpectations(t)
}
