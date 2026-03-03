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

// MockFavoriteRepository mocks IFavoriteRepository
type MockFavoriteRepository struct {
	mock.Mock
}

func (m *MockFavoriteRepository) AddFavorite(ctx context.Context, userID, fileID uint) (*entity.Favorite, error) {
	args := m.Called(ctx, userID, fileID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Favorite), args.Error(1)
}

func (m *MockFavoriteRepository) RemoveFavorite(ctx context.Context, userID, fileID uint) error {
	args := m.Called(ctx, userID, fileID)
	return args.Error(0)
}

func (m *MockFavoriteRepository) GetFavoritesByUserID(ctx context.Context, userID uint, filter request.ListFavoritesRequestDTO) ([]entity.CloudFile, int64, error) {
	args := m.Called(ctx, userID, filter)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]entity.CloudFile), args.Get(1).(int64), args.Error(2)
}

func (m *MockFavoriteRepository) CheckIsFavorited(ctx context.Context, userID, fileID uint) (bool, error) {
	args := m.Called(ctx, userID, fileID)
	return args.Bool(0), args.Error(1)
}

func TestAddFavorite_Success(t *testing.T) {
	t.Logf("TestAddFavorite_Success: verifying add favorite when user owns file")

	mockFavoriteRepo := new(MockFavoriteRepository)
	mockFileRepo := new(MockDownloadRepository)
	mockListRepo := new(MockListRepository)

	uc := NewFavoriteUseCase(mockFavoriteRepo, mockFileRepo, mockListRepo, 5*time.Second)

	ctx := context.Background()
	userID := uint(1)
	fileID := uint(100)

	mockFile := &entity.CloudFile{ID: fileID, UserID: userID, FileName: "test.jpg"}
	favoritedAt := time.Now()
	mockFavorite := &entity.Favorite{UserID: userID, FileID: fileID, FavoritedAt: favoritedAt}

	mockFileRepo.On("GetFileByID", mock.Anything, fileID).Return(mockFile, nil)
	mockFavoriteRepo.On("AddFavorite", mock.Anything, userID, fileID).Return(mockFavorite, nil)

	result, err := uc.AddFavorite(ctx, userID, fileID)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.True(t, result.Success)
	assert.Equal(t, favoritedAt, result.FavoritedAt)
	t.Logf("Success: favorite added")

	mockFavoriteRepo.AssertExpectations(t)
	mockFileRepo.AssertExpectations(t)
}

func TestAddFavorite_FileNotFound(t *testing.T) {
	t.Logf("TestAddFavorite_FileNotFound: verifying error when file does not exist")

	mockFavoriteRepo := new(MockFavoriteRepository)
	mockFileRepo := new(MockDownloadRepository)
	mockListRepo := new(MockListRepository)

	uc := NewFavoriteUseCase(mockFavoriteRepo, mockFileRepo, mockListRepo, 5*time.Second)

	ctx := context.Background()
	userID := uint(1)
	fileID := uint(999)

	mockFileRepo.On("GetFileByID", mock.Anything, fileID).Return(nil, errors.New("record not found"))

	result, err := uc.AddFavorite(ctx, userID, fileID)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "file not found")
	t.Logf("Expected: AddFavorite should not be called")

	mockFileRepo.AssertExpectations(t)
	mockFavoriteRepo.AssertNotCalled(t, "AddFavorite")
}

func TestAddFavorite_AccessDenied(t *testing.T) {
	t.Logf("TestAddFavorite_AccessDenied: verifying error when user does not own file")

	mockFavoriteRepo := new(MockFavoriteRepository)
	mockFileRepo := new(MockDownloadRepository)
	mockListRepo := new(MockListRepository)

	uc := NewFavoriteUseCase(mockFavoriteRepo, mockFileRepo, mockListRepo, 5*time.Second)

	ctx := context.Background()
	requesterID := uint(2)
	ownerID := uint(1)
	fileID := uint(100)

	mockFile := &entity.CloudFile{ID: fileID, UserID: ownerID}
	mockFileRepo.On("GetFileByID", mock.Anything, fileID).Return(mockFile, nil)

	result, err := uc.AddFavorite(ctx, requesterID, fileID)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "access denied")
	t.Logf("Expected: AddFavorite should not be called when access denied")

	mockFileRepo.AssertExpectations(t)
	mockFavoriteRepo.AssertNotCalled(t, "AddFavorite")
}

func TestRemoveFavorite_Success(t *testing.T) {
	t.Logf("TestRemoveFavorite_Success: verifying remove favorite (idempotent)")

	mockFavoriteRepo := new(MockFavoriteRepository)
	mockFileRepo := new(MockDownloadRepository)
	mockListRepo := new(MockListRepository)

	uc := NewFavoriteUseCase(mockFavoriteRepo, mockFileRepo, mockListRepo, 5*time.Second)

	ctx := context.Background()
	userID := uint(1)
	fileID := uint(100)

	mockFavoriteRepo.On("RemoveFavorite", mock.Anything, userID, fileID).Return(nil)

	err := uc.RemoveFavorite(ctx, userID, fileID)

	assert.NoError(t, err)
	t.Logf("Success: favorite removed")

	mockFavoriteRepo.AssertExpectations(t)
}

func TestListFavorites_Success(t *testing.T) {
	t.Logf("TestListFavorites_Success: verifying list favorites with presigned URLs")

	mockFavoriteRepo := new(MockFavoriteRepository)
	mockFileRepo := new(MockDownloadRepository)
	mockListRepo := new(MockListRepository)

	uc := NewFavoriteUseCase(mockFavoriteRepo, mockFileRepo, mockListRepo, 5*time.Second)

	ctx := context.Background()
	userID := uint(1)
	filter := request.ListFavoritesRequestDTO{Page: 1, Size: 20}

	mockFiles := []entity.CloudFile{
		{ID: 1, UserID: userID, FileName: "fav1.jpg", S3Key: "user1/fav1.jpg", FileType: entity.FileTypeImage},
	}

	mockFavoriteRepo.On("GetFavoritesByUserID", mock.Anything, userID, filter).Return(mockFiles, int64(1), nil)
	mockListRepo.On("GeneratePresignedDownloadURL", mock.Anything, "user1/fav1.jpg", 1*time.Hour).Return("https://s3.url/fav1", nil)

	result, err := uc.ListFavorites(ctx, userID, filter)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result.Data, 1)
	assert.Equal(t, "fav1.jpg", result.Data[0].FileName)
	assert.Equal(t, "https://s3.url/fav1", result.Data[0].DownloadURL)
	assert.Equal(t, int64(1), result.Pagination.Total)
	t.Logf("Success: favorites listed with presigned URLs")

	mockFavoriteRepo.AssertExpectations(t)
	mockListRepo.AssertExpectations(t)
}
