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
	"gorm.io/gorm"
)

// MockUploadRepository is a mock implementation of IUploadCloudRepositoryRepository
type MockUploadRepository struct {
	mock.Mock
}

func (m *MockUploadRepository) CreateFile(ctx context.Context, file *entity.CloudFile) error {
	args := m.Called(ctx, file)
	return args.Error(0)
}

func (m *MockUploadRepository) GeneratePresignedUploadURL(ctx context.Context, key, contentType string, expiration time.Duration) (string, error) {
	args := m.Called(ctx, key, contentType, expiration)
	return args.String(0), args.Error(1)
}

func (m *MockUploadRepository) DeleteFile(ctx context.Context, fileID uint) error {
	args := m.Called(ctx, fileID)
	return args.Error(0)
}

func (m *MockUploadRepository) GetFileByID(ctx context.Context, fileID uint) (*entity.CloudFile, error) {
	args := m.Called(ctx, fileID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.CloudFile), args.Error(1)
}

// MockUserStatsRepository is a mock implementation of IUserStatsCloudRepositoryRepository
type MockUserStatsRepository struct {
	mock.Mock
}

func (m *MockUserStatsRepository) GetTotalStorageUsed(ctx context.Context, userID uint) (int64, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockUserStatsRepository) GetMonthlyUploadCount(ctx context.Context, userID uint, year int, month int) (int, error) {
	args := m.Called(ctx, userID, year, month)
	return args.Int(0), args.Error(1)
}

func (m *MockUserStatsRepository) GetMonthlyDownloadCount(ctx context.Context, userID uint, year int, month int) (int, error) {
	args := m.Called(ctx, userID, year, month)
	return args.Int(0), args.Error(1)
}

func (m *MockUserStatsRepository) GetMonthlyTagsCreatedCount(ctx context.Context, userID uint, year int, month int) (int, error) {
	args := m.Called(ctx, userID, year, month)
	return args.Int(0), args.Error(1)
}

func (m *MockUserStatsRepository) LogActivity(ctx context.Context, activity *entity.ActivityLog) error {
	args := m.Called(ctx, activity)
	return args.Error(0)
}

// MockDB is a mock implementation of gorm.DB for testing
type MockDB struct {
	*gorm.DB
	mock.Mock
}

func (m *MockDB) WithContext(ctx context.Context) *gorm.DB {
	args := m.Called(ctx)
	return args.Get(0).(*gorm.DB)
}

func TestRequestUploadURL_MIMETypeValidation(t *testing.T) {
	tests := []struct {
		name        string
		fileType    string
		contentType string
		fileName    string
		fileSize    int64
		expectError bool
		errorMsg    string
	}{
		// iPhone Image Formats
		{
			name:        "HEIC image should be accepted",
			fileType:    "image",
			contentType: "image/heic",
			fileName:    "photo.heic",
			fileSize:    1024000,
			expectError: false,
		},
		{
			name:        "HEIF image should be accepted",
			fileType:    "image",
			contentType: "image/heif",
			fileName:    "photo.heif",
			fileSize:    1024000,
			expectError: false,
		},
		// iPhone Video Formats
		{
			name:        "MOV video should be accepted",
			fileType:    "video",
			contentType: "video/quicktime",
			fileName:    "video.mov",
			fileSize:    5120000,
			expectError: false,
		},
		{
			name:        "M4V video should be accepted",
			fileType:    "video",
			contentType: "video/x-m4v",
			fileName:    "video.m4v",
			fileSize:    5120000,
			expectError: false,
		},
		// Existing Formats (Regression Tests)
		{
			name:        "JPEG image should be accepted",
			fileType:    "image",
			contentType: "image/jpeg",
			fileName:    "photo.jpg",
			fileSize:    1024000,
			expectError: false,
		},
		{
			name:        "PNG image should be accepted",
			fileType:    "image",
			contentType: "image/png",
			fileName:    "photo.png",
			fileSize:    1024000,
			expectError: false,
		},
		{
			name:        "WebP image should be accepted",
			fileType:    "image",
			contentType: "image/webp",
			fileName:    "photo.webp",
			fileSize:    1024000,
			expectError: false,
		},
		{
			name:        "MP4 video should be accepted",
			fileType:    "video",
			contentType: "video/mp4",
			fileName:    "video.mp4",
			fileSize:    5120000,
			expectError: false,
		},
		{
			name:        "WebM video should be accepted",
			fileType:    "video",
			contentType: "video/webm",
			fileName:    "video.webm",
			fileSize:    5120000,
			expectError: false,
		},
		{
			name:        "AVI video should be accepted",
			fileType:    "video",
			contentType: "video/x-msvideo",
			fileName:    "video.avi",
			fileSize:    5120000,
			expectError: false,
		},
		// Invalid MIME Types
		{
			name:        "PDF should be rejected for image type",
			fileType:    "image",
			contentType: "application/pdf",
			fileName:    "document.pdf",
			fileSize:    1024000,
			expectError: true,
			errorMsg:    "이미지 파일만 업로드할 수 있습니다",
		},
		{
			name:        "Text file should be rejected for image type",
			fileType:    "image",
			contentType: "text/plain",
			fileName:    "document.txt",
			fileSize:    1024,
			expectError: true,
			errorMsg:    "이미지 파일만 업로드할 수 있습니다",
		},
		{
			name:        "JSON should be rejected for video type",
			fileType:    "video",
			contentType: "application/json",
			fileName:    "data.json",
			fileSize:    1024,
			expectError: true,
			errorMsg:    "동영상 파일만 업로드할 수 있습니다",
		},
		{
			name:        "HTML should be rejected for video type",
			fileType:    "video",
			contentType: "text/html",
			fileName:    "page.html",
			fileSize:    1024,
			expectError: true,
			errorMsg:    "동영상 파일만 업로드할 수 있습니다",
		},
		// Edge Cases
		{
			name:        "Image with video content type should be rejected",
			fileType:    "image",
			contentType: "video/mp4",
			fileName:    "photo.jpg",
			fileSize:    1024000,
			expectError: true,
			errorMsg:    "이미지 파일만 업로드할 수 있습니다",
		},
		{
			name:        "Video with image content type should be rejected",
			fileType:    "video",
			contentType: "image/jpeg",
			fileName:    "video.mp4",
			fileSize:    5120000,
			expectError: true,
			errorMsg:    "동영상 파일만 업로드할 수 있습니다",
		},
		// Additional iPhone formats
		{
			name:        "HEIC-sequence should be accepted",
			fileType:    "image",
			contentType: "image/heic-sequence",
			fileName:    "burst.heic",
			fileSize:    2048000,
			expectError: false,
		},
		{
			name:        "HEIF-sequence should be accepted",
			fileType:    "image",
			contentType: "image/heif-sequence",
			fileName:    "burst.heif",
			fileSize:    2048000,
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			mockRepo := new(MockUploadRepository)
			mockStatsRepo := new(MockUserStatsRepository)
			mockDB := &gorm.DB{}

			useCase := &UploadCloudRepositoryUseCase{
				Repo:           mockRepo,
				StatsRepo:      mockStatsRepo,
				DB:             mockDB,
				Bucket:         "test-bucket",
				ContextTimeout: 5 * time.Second,
			}

			// Prepare request
			req := &request.UploadRequestDTO{
				FileName:    tt.fileName,
				FileType:    tt.fileType,
				ContentType: tt.contentType,
				FileSize:    tt.fileSize,
			}

			// Only mock successful operations if we expect success
			if !tt.expectError {
				// Mock CreateFile to succeed
				mockRepo.On("CreateFile", mock.Anything, mock.AnythingOfType("*entity.CloudFile")).Return(nil)

				// Mock GeneratePresignedUploadURL to return a valid URL
				mockRepo.On("GeneratePresignedUploadURL",
					mock.Anything,
					mock.AnythingOfType("string"),
					tt.contentType,
					DefaultUploadExpiration,
				).Return("https://test-bucket.s3.amazonaws.com/presigned-url", nil)

				// Mock stats logging (should not fail the request)
				mockStatsRepo.On("LogActivity", mock.Anything, mock.AnythingOfType("*entity.ActivityLog")).Return(nil)
			}

			// Execute
			ctx := context.Background()
			result, err := useCase.RequestUploadURL(ctx, 1, req)

			// Assert
			if tt.expectError {
				assert.Error(t, err, "Expected error for %s", tt.name)
				assert.Nil(t, result, "Result should be nil on error")
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg, "Error message should contain expected text")
				}
			} else {
				assert.NoError(t, err, "Expected no error for %s", tt.name)
				assert.NotNil(t, result, "Result should not be nil")
				assert.NotEmpty(t, result.UploadURL, "Upload URL should not be empty")
				assert.NotEmpty(t, result.S3Key, "S3 key should not be empty")
				// Note: FileID is 0 in tests because we're using mocks, but the important thing is no error occurred
			}

			// Verify mocks
			mockRepo.AssertExpectations(t)
			mockStatsRepo.AssertExpectations(t)
		})
	}
}

func TestRequestUploadURL_ThumbnailGeneration(t *testing.T) {
	tests := []struct {
		name               string
		fileType           string
		contentType        string
		expectThumbnailURL bool
	}{
		{
			name:               "Image should generate thumbnail URL",
			fileType:           "image",
			contentType:        "image/jpeg",
			expectThumbnailURL: true,
		},
		{
			name:               "Video should generate thumbnail URL",
			fileType:           "video",
			contentType:        "video/mp4",
			expectThumbnailURL: true,
		},
		{
			name:               "HEIC should generate thumbnail URL",
			fileType:           "image",
			contentType:        "image/heic",
			expectThumbnailURL: true,
		},
		{
			name:               "MOV should generate thumbnail URL",
			fileType:           "video",
			contentType:        "video/quicktime",
			expectThumbnailURL: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			mockRepo := new(MockUploadRepository)
			mockStatsRepo := new(MockUserStatsRepository)
			mockDB := &gorm.DB{}

			useCase := &UploadCloudRepositoryUseCase{
				Repo:           mockRepo,
				StatsRepo:      mockStatsRepo,
				DB:             mockDB,
				Bucket:         "test-bucket",
				ContextTimeout: 5 * time.Second,
			}

			req := &request.UploadRequestDTO{
				FileName:    "test." + tt.fileType,
				FileType:    tt.fileType,
				ContentType: tt.contentType,
				FileSize:    1024000,
			}

			// Mock operations
			mockRepo.On("CreateFile", mock.Anything, mock.AnythingOfType("*entity.CloudFile")).Return(nil)
			mockRepo.On("GeneratePresignedUploadURL",
				mock.Anything,
				mock.AnythingOfType("string"),
				tt.contentType,
				DefaultUploadExpiration,
			).Return("https://test-bucket.s3.amazonaws.com/presigned-url", nil)
			mockStatsRepo.On("LogActivity", mock.Anything, mock.AnythingOfType("*entity.ActivityLog")).Return(nil)

			// Execute
			ctx := context.Background()
			result, err := useCase.RequestUploadURL(ctx, 1, req)

			// Assert
			assert.NoError(t, err)
			assert.NotNil(t, result)

			if tt.expectThumbnailURL {
				assert.NotEmpty(t, result.ThumbnailURL, "Thumbnail URL should be generated")
				assert.NotEmpty(t, result.ThumbnailKey, "Thumbnail key should be generated")
			}

			mockRepo.AssertExpectations(t)
			mockStatsRepo.AssertExpectations(t)
		})
	}
}

func TestRequestUploadURL_ErrorHandling(t *testing.T) {
	tests := []struct {
		name           string
		setupMock      func(*MockUploadRepository, *MockUserStatsRepository)
		expectError    bool
		errorSubstring string
	}{
		{
			name: "Database error on CreateFile",
			setupMock: func(repo *MockUploadRepository, statsRepo *MockUserStatsRepository) {
				repo.On("CreateFile", mock.Anything, mock.AnythingOfType("*entity.CloudFile")).
					Return(errors.New("database error"))
			},
			expectError:    true,
			errorSubstring: "failed to create file record",
		},
		{
			name: "S3 error on GeneratePresignedUploadURL",
			setupMock: func(repo *MockUploadRepository, statsRepo *MockUserStatsRepository) {
				repo.On("CreateFile", mock.Anything, mock.AnythingOfType("*entity.CloudFile")).Return(nil)
				repo.On("GeneratePresignedUploadURL",
					mock.Anything,
					mock.AnythingOfType("string"),
					"image/jpeg",
					DefaultUploadExpiration,
				).Return("", errors.New("S3 error"))
				// LogActivity is called after CreateFile succeeds
				statsRepo.On("LogActivity", mock.Anything, mock.AnythingOfType("*entity.ActivityLog")).Return(nil)
			},
			expectError:    true,
			errorSubstring: "failed to generate upload URL",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			mockRepo := new(MockUploadRepository)
			mockStatsRepo := new(MockUserStatsRepository)
			mockDB := &gorm.DB{}

			tt.setupMock(mockRepo, mockStatsRepo)

			useCase := &UploadCloudRepositoryUseCase{
				Repo:           mockRepo,
				StatsRepo:      mockStatsRepo,
				DB:             mockDB,
				Bucket:         "test-bucket",
				ContextTimeout: 5 * time.Second,
			}

			req := &request.UploadRequestDTO{
				FileName:    "test.jpg",
				FileType:    "image",
				ContentType: "image/jpeg",
				FileSize:    1024000,
			}

			// Execute
			ctx := context.Background()
			result, err := useCase.RequestUploadURL(ctx, 1, req)

			// Assert
			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, result)
				if tt.errorSubstring != "" {
					assert.Contains(t, err.Error(), tt.errorSubstring)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}
