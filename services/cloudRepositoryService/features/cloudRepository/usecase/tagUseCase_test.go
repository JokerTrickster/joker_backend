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

// MockTagRepository mocks ITagRepository
type MockTagRepository struct {
	mock.Mock
}

func (m *MockTagRepository) GetFileByID(ctx context.Context, id uint, userID uint) (*entity.CloudFile, error) {
	args := m.Called(ctx, id, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.CloudFile), args.Error(1)
}

func (m *MockTagRepository) UpdateFileTags(ctx context.Context, fileID uint, userID uint, tags []entity.Tag) error {
	args := m.Called(ctx, fileID, userID, tags)
	return args.Error(0)
}

func (m *MockTagRepository) AddTagToFile(ctx context.Context, fileID uint, userID uint, tag entity.Tag) error {
	args := m.Called(ctx, fileID, userID, tag)
	return args.Error(0)
}

func (m *MockTagRepository) RemoveTagFromFile(ctx context.Context, fileID uint, userID uint, tagName string) error {
	args := m.Called(ctx, fileID, userID, tagName)
	return args.Error(0)
}

func (m *MockTagRepository) FindOrCreateTag(ctx context.Context, userID uint, tagName string) (*entity.Tag, error) {
	args := m.Called(ctx, userID, tagName)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Tag), args.Error(1)
}

// MockTagStatsRepository mocks IUserStatsCloudRepositoryRepository
type MockTagStatsRepository struct {
	mock.Mock
}

func (m *MockTagStatsRepository) GetTotalStorageUsed(ctx context.Context, userID uint) (int64, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockTagStatsRepository) GetMonthlyUploadCount(ctx context.Context, userID uint, year int, month int) (int, error) {
	args := m.Called(ctx, userID, year, month)
	return args.Int(0), args.Error(1)
}

func (m *MockTagStatsRepository) GetMonthlyDownloadCount(ctx context.Context, userID uint, year int, month int) (int, error) {
	args := m.Called(ctx, userID, year, month)
	return args.Int(0), args.Error(1)
}

func (m *MockTagStatsRepository) GetMonthlyTagsCreatedCount(ctx context.Context, userID uint, year int, month int) (int, error) {
	args := m.Called(ctx, userID, year, month)
	return args.Int(0), args.Error(1)
}

func (m *MockTagStatsRepository) LogActivity(ctx context.Context, activity *entity.ActivityLog) error {
	args := m.Called(ctx, activity)
	return args.Error(0)
}

func TestTagUseCase_UpdateFileTags_Success(t *testing.T) {
	t.Logf("TestTagUseCase_UpdateFileTags_Success: verifying tags are updated when file exists and user owns it")

	mockTagRepo := new(MockTagRepository)
	mockStatsRepo := new(MockTagStatsRepository)

	uc := NewTagUseCase(mockTagRepo, mockStatsRepo, 5*time.Second)

	ctx := context.Background()
	userID := uint(1)
	fileID := uint(100)
	req := &request.UpdateFileTagsRequestDTO{Tags: []string{"photo", "vacation"}}

	mockFile := &entity.CloudFile{ID: fileID, UserID: userID, FileName: "test.jpg"}
	tag1 := &entity.Tag{ID: 1, UserID: userID, Name: "photo", CreatedAt: time.Now()}
	tag2 := &entity.Tag{ID: 2, UserID: userID, Name: "vacation", CreatedAt: time.Now()}

	mockTagRepo.On("GetFileByID", mock.Anything, fileID, userID).Return(mockFile, nil)
	mockTagRepo.On("FindOrCreateTag", mock.Anything, userID, "photo").Return(tag1, nil)
	mockTagRepo.On("FindOrCreateTag", mock.Anything, userID, "vacation").Return(tag2, nil)
	mockStatsRepo.On("LogActivity", mock.Anything, mock.AnythingOfType("*entity.ActivityLog")).Return(nil).Times(2)
	mockTagRepo.On("UpdateFileTags", mock.Anything, fileID, userID, mock.AnythingOfType("[]entity.Tag")).Return(nil)

	resp, err := uc.UpdateFileTags(ctx, userID, fileID, req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, fileID, resp.FileID)
	assert.ElementsMatch(t, []string{"photo", "vacation"}, resp.Tags)
	t.Logf("Success: tags updated to [photo, vacation]")

	mockTagRepo.AssertExpectations(t)
	mockStatsRepo.AssertExpectations(t)
}

func TestTagUseCase_UpdateFileTags_FileNotFound(t *testing.T) {
	t.Logf("TestTagUseCase_UpdateFileTags_FileNotFound: verifying error when file does not exist or user lacks access")

	mockTagRepo := new(MockTagRepository)
	mockStatsRepo := new(MockTagStatsRepository)

	uc := NewTagUseCase(mockTagRepo, mockStatsRepo, 5*time.Second)

	ctx := context.Background()
	userID := uint(1)
	fileID := uint(999)
	req := &request.UpdateFileTagsRequestDTO{Tags: []string{"photo"}}

	mockTagRepo.On("GetFileByID", mock.Anything, fileID, userID).Return(nil, errors.New("record not found"))

	resp, err := uc.UpdateFileTags(ctx, userID, fileID, req)

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "file not found")
	t.Logf("Expected: UpdateFileTags and FindOrCreateTag should not be called")

	mockTagRepo.AssertExpectations(t)
	mockTagRepo.AssertNotCalled(t, "UpdateFileTags")
	mockTagRepo.AssertNotCalled(t, "FindOrCreateTag")
}

func TestTagUseCase_UpdateFileTags_SkipsEmptyTags(t *testing.T) {
	t.Logf("TestTagUseCase_UpdateFileTags_SkipsEmptyTags: verifying empty string tags are skipped and non-empty tags are processed")

	mockTagRepo := new(MockTagRepository)
	mockStatsRepo := new(MockTagStatsRepository)

	uc := NewTagUseCase(mockTagRepo, mockStatsRepo, 5*time.Second)

	ctx := context.Background()
	userID := uint(1)
	fileID := uint(100)
	req := &request.UpdateFileTagsRequestDTO{Tags: []string{"", "valid", "", "another"}}

	mockFile := &entity.CloudFile{ID: fileID, UserID: userID, FileName: "test.jpg"}
	tag1 := &entity.Tag{ID: 1, UserID: userID, Name: "valid", CreatedAt: time.Now()}
	tag2 := &entity.Tag{ID: 2, UserID: userID, Name: "another", CreatedAt: time.Now()}

	mockTagRepo.On("GetFileByID", mock.Anything, fileID, userID).Return(mockFile, nil)
	mockTagRepo.On("FindOrCreateTag", mock.Anything, userID, "valid").Return(tag1, nil)
	mockTagRepo.On("FindOrCreateTag", mock.Anything, userID, "another").Return(tag2, nil)
	mockStatsRepo.On("LogActivity", mock.Anything, mock.AnythingOfType("*entity.ActivityLog")).Return(nil).Times(2)
	mockTagRepo.On("UpdateFileTags", mock.Anything, fileID, userID, mock.AnythingOfType("[]entity.Tag")).Return(nil)

	resp, err := uc.UpdateFileTags(ctx, userID, fileID, req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.ElementsMatch(t, []string{"valid", "another"}, resp.Tags)
	t.Logf("Success: only non-empty tags processed")

	mockTagRepo.AssertExpectations(t)
	mockTagRepo.AssertNumberOfCalls(t, "FindOrCreateTag", 2)
}

func TestTagUseCase_AddTagToFile_Success(t *testing.T) {
	t.Logf("TestTagUseCase_AddTagToFile_Success: verifying single tag is added when file exists")

	mockTagRepo := new(MockTagRepository)
	mockStatsRepo := new(MockTagStatsRepository)

	uc := NewTagUseCase(mockTagRepo, mockStatsRepo, 5*time.Second)

	ctx := context.Background()
	userID := uint(1)
	fileID := uint(100)
	req := &request.AddFileTagRequestDTO{Tag: "landscape"}

	mockFile := &entity.CloudFile{ID: fileID, UserID: userID, FileName: "test.jpg"}
	tag := &entity.Tag{ID: 1, UserID: userID, Name: "landscape", CreatedAt: time.Now()}

	mockTagRepo.On("GetFileByID", mock.Anything, fileID, userID).Return(mockFile, nil)
	mockTagRepo.On("FindOrCreateTag", mock.Anything, userID, "landscape").Return(tag, nil)
	mockTagRepo.On("AddTagToFile", mock.Anything, fileID, userID, *tag).Return(nil)
	mockStatsRepo.On("LogActivity", mock.Anything, mock.AnythingOfType("*entity.ActivityLog")).Return(nil)

	resp, err := uc.AddTagToFile(ctx, userID, fileID, req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, fileID, resp.FileID)
	assert.Equal(t, "landscape", resp.Tag)
	t.Logf("Success: tag 'landscape' added")

	mockTagRepo.AssertExpectations(t)
	mockStatsRepo.AssertExpectations(t)
}

func TestTagUseCase_AddTagToFile_FileNotFound(t *testing.T) {
	t.Logf("TestTagUseCase_AddTagToFile_FileNotFound: verifying error when file does not exist")

	mockTagRepo := new(MockTagRepository)
	mockStatsRepo := new(MockTagStatsRepository)

	uc := NewTagUseCase(mockTagRepo, mockStatsRepo, 5*time.Second)

	ctx := context.Background()
	userID := uint(1)
	fileID := uint(999)
	req := &request.AddFileTagRequestDTO{Tag: "photo"}

	mockTagRepo.On("GetFileByID", mock.Anything, fileID, userID).Return(nil, errors.New("record not found"))

	resp, err := uc.AddTagToFile(ctx, userID, fileID, req)

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "file not found")
	t.Logf("Expected: FindOrCreateTag and AddTagToFile should not be called")

	mockTagRepo.AssertExpectations(t)
	mockTagRepo.AssertNotCalled(t, "FindOrCreateTag")
	mockTagRepo.AssertNotCalled(t, "AddTagToFile")
}

func TestTagUseCase_RemoveTagFromFile_Success(t *testing.T) {
	t.Logf("TestTagUseCase_RemoveTagFromFile_Success: verifying tag is removed when file exists and user owns it")

	mockTagRepo := new(MockTagRepository)
	mockStatsRepo := new(MockTagStatsRepository)

	uc := NewTagUseCase(mockTagRepo, mockStatsRepo, 5*time.Second)

	ctx := context.Background()
	userID := uint(1)
	fileID := uint(100)
	tagName := "obsolete"

	mockFile := &entity.CloudFile{ID: fileID, UserID: userID, FileName: "test.jpg"}

	mockTagRepo.On("GetFileByID", mock.Anything, fileID, userID).Return(mockFile, nil)
	mockTagRepo.On("RemoveTagFromFile", mock.Anything, fileID, userID, tagName).Return(nil)
	mockStatsRepo.On("LogActivity", mock.Anything, mock.AnythingOfType("*entity.ActivityLog")).Return(nil)

	err := uc.RemoveTagFromFile(ctx, userID, fileID, tagName)

	assert.NoError(t, err)
	t.Logf("Success: tag '%s' removed", tagName)

	mockTagRepo.AssertExpectations(t)
	mockStatsRepo.AssertExpectations(t)
}

func TestTagUseCase_RemoveTagFromFile_FileNotFound(t *testing.T) {
	t.Logf("TestTagUseCase_RemoveTagFromFile_FileNotFound: verifying error when file does not exist")

	mockTagRepo := new(MockTagRepository)
	mockStatsRepo := new(MockTagStatsRepository)

	uc := NewTagUseCase(mockTagRepo, mockStatsRepo, 5*time.Second)

	ctx := context.Background()
	userID := uint(1)
	fileID := uint(999)
	tagName := "photo"

	mockTagRepo.On("GetFileByID", mock.Anything, fileID, userID).Return(nil, errors.New("record not found"))

	err := uc.RemoveTagFromFile(ctx, userID, fileID, tagName)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "file not found")
	t.Logf("Expected: RemoveTagFromFile should not be called")

	mockTagRepo.AssertExpectations(t)
	mockTagRepo.AssertNotCalled(t, "RemoveTagFromFile")
}
