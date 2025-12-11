package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/JokerTrickster/joker_backend/services/cloudRepositoryService/features/cloudRepository/model/entity"
	_interface "github.com/JokerTrickster/joker_backend/services/cloudRepositoryService/features/cloudRepository/model/interface"
	"github.com/JokerTrickster/joker_backend/services/cloudRepositoryService/features/cloudRepository/model/request"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

// MockMultipartUploadRepository is a mock implementation of IMultipartUploadRepository
type MockMultipartUploadRepository struct {
	mock.Mock
}

func (m *MockMultipartUploadRepository) CreateMultipartUpload(ctx context.Context, bucket, key, contentType string) (string, error) {
	args := m.Called(ctx, bucket, key, contentType)
	return args.String(0), args.Error(1)
}

func (m *MockMultipartUploadRepository) CreateMultipartUploadRecord(ctx context.Context, upload *entity.MultipartUpload) error {
	args := m.Called(ctx, upload)
	return args.Error(0)
}

func (m *MockMultipartUploadRepository) GetMultipartUpload(ctx context.Context, uploadID string, userID uint) (*entity.MultipartUpload, error) {
	args := m.Called(ctx, uploadID, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.MultipartUpload), args.Error(1)
}

func (m *MockMultipartUploadRepository) GeneratePresignedUploadPartURL(ctx context.Context, bucket, key, uploadID string, partNumber int, expiration time.Duration) (string, error) {
	args := m.Called(ctx, bucket, key, uploadID, partNumber, expiration)
	return args.String(0), args.Error(1)
}

func (m *MockMultipartUploadRepository) CompleteMultipartUpload(ctx context.Context, bucket, key, uploadID string, parts []_interface.CompletedPart) error {
	args := m.Called(ctx, bucket, key, uploadID, parts)
	return args.Error(0)
}

func (m *MockMultipartUploadRepository) AbortMultipartUpload(ctx context.Context, bucket, key, uploadID string) error {
	args := m.Called(ctx, bucket, key, uploadID)
	return args.Error(0)
}

func (m *MockMultipartUploadRepository) UpdateMultipartUploadStatus(ctx context.Context, uploadID string, status entity.MultipartUploadStatus) error {
	args := m.Called(ctx, uploadID, status)
	return args.Error(0)
}

func TestInitiateMultipartUpload_MIMETypeValidation(t *testing.T) {
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
			fileSize:    50 * 1024 * 1024, // 50MB
			expectError: false,
		},
		{
			name:        "HEIF image should be accepted",
			fileType:    "image",
			contentType: "image/heif",
			fileName:    "photo.heif",
			fileSize:    50 * 1024 * 1024,
			expectError: false,
		},
		// iPhone Video Formats
		{
			name:        "MOV video should be accepted",
			fileType:    "video",
			contentType: "video/quicktime",
			fileName:    "video.mov",
			fileSize:    100 * 1024 * 1024, // 100MB
			expectError: false,
		},
		{
			name:        "M4V video should be accepted",
			fileType:    "video",
			contentType: "video/x-m4v",
			fileName:    "video.m4v",
			fileSize:    100 * 1024 * 1024,
			expectError: false,
		},
		// Existing Formats (Regression Tests)
		{
			name:        "JPEG image should be accepted",
			fileType:    "image",
			contentType: "image/jpeg",
			fileName:    "photo.jpg",
			fileSize:    50 * 1024 * 1024,
			expectError: false,
		},
		{
			name:        "PNG image should be accepted",
			fileType:    "image",
			contentType: "image/png",
			fileName:    "photo.png",
			fileSize:    50 * 1024 * 1024,
			expectError: false,
		},
		{
			name:        "MP4 video should be accepted",
			fileType:    "video",
			contentType: "video/mp4",
			fileName:    "video.mp4",
			fileSize:    100 * 1024 * 1024,
			expectError: false,
		},
		{
			name:        "WebM video should be accepted",
			fileType:    "video",
			contentType: "video/webm",
			fileName:    "video.webm",
			fileSize:    100 * 1024 * 1024,
			expectError: false,
		},
		// Invalid MIME Types
		{
			name:        "PDF should be rejected for image type",
			fileType:    "image",
			contentType: "application/pdf",
			fileName:    "document.pdf",
			fileSize:    10 * 1024 * 1024,
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
			fileSize:    50 * 1024 * 1024,
			expectError: true,
			errorMsg:    "이미지 파일만 업로드할 수 있습니다",
		},
		{
			name:        "Video with image content type should be rejected",
			fileType:    "video",
			contentType: "image/jpeg",
			fileName:    "video.mp4",
			fileSize:    100 * 1024 * 1024,
			expectError: true,
			errorMsg:    "동영상 파일만 업로드할 수 있습니다",
		},
		// Additional iPhone formats
		{
			name:        "HEIC-sequence should be accepted",
			fileType:    "image",
			contentType: "image/heic-sequence",
			fileName:    "burst.heic",
			fileSize:    80 * 1024 * 1024,
			expectError: false,
		},
		{
			name:        "HEIF-sequence should be accepted",
			fileType:    "image",
			contentType: "image/heif-sequence",
			fileName:    "burst.heif",
			fileSize:    80 * 1024 * 1024,
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			mockRepo := new(MockMultipartUploadRepository)
			mockStatsRepo := new(MockUserStatsRepository)
			mockDB := &gorm.DB{}

			useCase := &MultipartUploadUseCase{
				Repo:           mockRepo,
				StatsRepo:      mockStatsRepo,
				DB:             mockDB,
				Bucket:         "test-bucket",
				ContextTimeout: 5 * time.Second,
			}

			// Prepare request
			req := &request.InitiateMultipartUploadRequestDTO{
				FileName:    tt.fileName,
				FileType:    tt.fileType,
				ContentType: tt.contentType,
				FileSize:    tt.fileSize,
			}

			// Only mock successful operations if we expect success
			if !tt.expectError {
				// Mock CreateMultipartUpload to return a valid upload ID
				mockRepo.On("CreateMultipartUpload",
					mock.Anything,
					"test-bucket",
					mock.AnythingOfType("string"),
					tt.contentType,
				).Return("test-upload-id-123", nil)

				// Mock CreateMultipartUploadRecord to succeed
				mockRepo.On("CreateMultipartUploadRecord",
					mock.Anything,
					mock.AnythingOfType("*entity.MultipartUpload"),
				).Return(nil)
			}

			// Execute
			ctx := context.Background()
			result, err := useCase.InitiateMultipartUpload(ctx, 1, req)

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
				assert.NotEmpty(t, result.UploadID, "Upload ID should not be empty")
				assert.NotEmpty(t, result.FileKey, "File key should not be empty")
				assert.Greater(t, result.PartSize, 0, "Part size should be greater than 0")
				assert.Greater(t, result.TotalParts, 0, "Total parts should be greater than 0")
			}

			// Verify mocks
			mockRepo.AssertExpectations(t)
		})
	}
}

// Note: CompleteMultipartUpload file type detection is tested indirectly through InitiateMultipartUpload
// since the content type is validated at initiation time and used during completion

func TestInitiateMultipartUpload_ErrorHandling(t *testing.T) {
	tests := []struct {
		name           string
		setupMock      func(*MockMultipartUploadRepository)
		expectError    bool
		errorSubstring string
	}{
		{
			name: "S3 error on CreateMultipartUpload",
			setupMock: func(repo *MockMultipartUploadRepository) {
				repo.On("CreateMultipartUpload",
					mock.Anything,
					"test-bucket",
					mock.AnythingOfType("string"),
					"image/jpeg",
				).Return("", errors.New("S3 error"))
			},
			expectError:    true,
			errorSubstring: "failed to initiate multipart upload",
		},
		{
			name: "Database error on CreateMultipartUploadRecord",
			setupMock: func(repo *MockMultipartUploadRepository) {
				repo.On("CreateMultipartUpload",
					mock.Anything,
					"test-bucket",
					mock.AnythingOfType("string"),
					"image/jpeg",
				).Return("test-upload-id", nil)
				repo.On("CreateMultipartUploadRecord",
					mock.Anything,
					mock.AnythingOfType("*entity.MultipartUpload"),
				).Return(errors.New("database error"))
				// Should attempt to abort the upload
				repo.On("AbortMultipartUpload",
					mock.Anything,
					"test-bucket",
					mock.AnythingOfType("string"),
					"test-upload-id",
				).Return(nil)
			},
			expectError:    true,
			errorSubstring: "failed to create upload record",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			mockRepo := new(MockMultipartUploadRepository)
			mockStatsRepo := new(MockUserStatsRepository)
			mockDB := &gorm.DB{}

			tt.setupMock(mockRepo)

			useCase := &MultipartUploadUseCase{
				Repo:           mockRepo,
				StatsRepo:      mockStatsRepo,
				DB:             mockDB,
				Bucket:         "test-bucket",
				ContextTimeout: 5 * time.Second,
			}

			req := &request.InitiateMultipartUploadRequestDTO{
				FileName:    "test.jpg",
				FileType:    "image",
				ContentType: "image/jpeg",
				FileSize:    50 * 1024 * 1024,
			}

			// Execute
			ctx := context.Background()
			result, err := useCase.InitiateMultipartUpload(ctx, 1, req)

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
