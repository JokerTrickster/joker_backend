package handler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/JokerTrickster/joker_backend/services/molandolanService/features/gallery/model/entity"
	_interface "github.com/JokerTrickster/joker_backend/services/molandolanService/features/gallery/model/interface"
	"github.com/JokerTrickster/joker_backend/services/molandolanService/features/gallery/usecase"
	"github.com/JokerTrickster/joker_backend/shared/aws"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockGalleryCreateRepository struct {
	createFunc        func(ctx context.Context, post *entity.GalleryPost) (*entity.GalleryPost, error)
	getAuthorNickname func(ctx context.Context, userID uint) (string, error)
}

func (m *mockGalleryCreateRepository) List(ctx context.Context, page, limit int) ([]entity.GalleryPost, int64, error) {
	return nil, 0, nil
}
func (m *mockGalleryCreateRepository) FindByID(ctx context.Context, id uint) (*entity.GalleryPost, error) {
	return nil, nil
}
func (m *mockGalleryCreateRepository) Create(ctx context.Context, post *entity.GalleryPost) (*entity.GalleryPost, error) {
	if m.createFunc != nil {
		return m.createFunc(ctx, post)
	}
	return post, nil
}
func (m *mockGalleryCreateRepository) Delete(ctx context.Context, id uint) error {
	return nil
}
func (m *mockGalleryCreateRepository) IsLiked(ctx context.Context, userID, galleryID uint) (bool, error) {
	return false, nil
}
func (m *mockGalleryCreateRepository) ToggleLike(ctx context.Context, userID, galleryID uint) (bool, int, error) {
	return false, 0, nil
}
func (m *mockGalleryCreateRepository) ListComments(ctx context.Context, galleryID uint, page, limit int) ([]entity.GalleryComment, int64, error) {
	return nil, 0, nil
}
func (m *mockGalleryCreateRepository) FindCommentByID(ctx context.Context, id uint) (*entity.GalleryComment, error) {
	return nil, nil
}
func (m *mockGalleryCreateRepository) CreateComment(ctx context.Context, comment *entity.GalleryComment) (*entity.GalleryComment, error) {
	return nil, nil
}
func (m *mockGalleryCreateRepository) DeleteComment(ctx context.Context, id uint) error {
	return nil
}
func (m *mockGalleryCreateRepository) GetAuthorNickname(ctx context.Context, userID uint) (string, error) {
	if m.getAuthorNickname != nil {
		return m.getAuthorNickname(ctx, userID)
	}
	return "testuser", nil
}
func (m *mockGalleryCreateRepository) GetAuthorInfo(ctx context.Context, userID uint) (string, *string, error) {
	if m.getAuthorNickname != nil {
		nick, _ := m.getAuthorNickname(ctx, userID)
		return nick, nil, nil
	}
	return "testuser", nil, nil
}
func (m *mockGalleryCreateRepository) IsLikedBatch(ctx context.Context, userID uint, galleryIDs []uint) (map[uint]bool, error) {
	return map[uint]bool{}, nil
}
func (m *mockGalleryCreateRepository) GetUserRole(ctx context.Context, userID uint) (string, error) {
	return "user", nil
}

var _ _interface.IGalleryRepository = (*mockGalleryCreateRepository)(nil)

// minimalValidJPEG is the smallest valid JPEG (107 bytes)
var minimalValidJPEG = []byte{
	0xff, 0xd8, 0xff, 0xe0, 0x00, 0x10, 0x4a, 0x46, 0x49, 0x46, 0x00, 0x01, 0x01, 0x00, 0x00, 0x01,
	0x00, 0x01, 0x00, 0x00, 0xff, 0xdb, 0x00, 0x43, 0x00, 0x08, 0x06, 0x06, 0x07, 0x06, 0x05, 0x08,
	0x07, 0x07, 0x07, 0x09, 0x09, 0x08, 0x0a, 0x0c, 0x14, 0x0d, 0x0c, 0x0b, 0x0b, 0x0c, 0x19, 0x12,
	0x13, 0x0f, 0x14, 0x1d, 0x1a, 0x1f, 0x1e, 0x1d, 0x1a, 0x1c, 0x1c, 0x20, 0x24, 0x2e, 0x27, 0x24,
	0x22, 0x2e, 0x1b, 0x1c, 0x1c, 0x28, 0x37, 0x29, 0x2c, 0x30, 0x31, 0x34, 0x34, 0x34, 0x1f, 0x27,
	0x39, 0x3d, 0x38, 0x32, 0x3c, 0x2e, 0x33, 0x34, 0x32, 0xff, 0xc0, 0x00, 0x0b, 0x08, 0x00, 0x01,
	0x00, 0x01, 0x01, 0x01, 0x11, 0x00, 0xff, 0xc4, 0x00, 0x1f, 0x00, 0x00, 0x01, 0x05, 0x01, 0x01,
	0x01, 0x01, 0x01, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01, 0x02, 0x03, 0x04,
	0x05, 0x06, 0x07, 0x08, 0x09, 0x0a, 0x0b, 0xff, 0xc4, 0x00, 0xb5, 0x10, 0x00, 0x02, 0x01, 0x03,
	0x03, 0x02, 0x04, 0x03, 0x05, 0x05, 0x04, 0x04, 0x00, 0x00, 0x01, 0x7d, 0x01, 0x02, 0x03, 0x00,
	0x04, 0x11, 0x05, 0x12, 0x21, 0x31, 0x41, 0x06, 0x13, 0x51, 0x61, 0x07, 0x22, 0x71, 0x14, 0x32,
	0x81, 0x91, 0xa1, 0x08, 0x23, 0x42, 0xb1, 0xc1, 0x15, 0x52, 0xd1, 0xf0, 0x24, 0x33, 0x62, 0x72,
	0x82, 0x09, 0x0a, 0x16, 0x17, 0x18, 0x19, 0x1a, 0x25, 0x26, 0x27, 0x28, 0x29, 0x2a, 0x34, 0x35,
	0x36, 0x37, 0x38, 0x39, 0x3a, 0x43, 0x44, 0x45, 0x46, 0x47, 0x48, 0x49, 0x4a, 0x53, 0x54, 0x55,
	0x56, 0x57, 0x58, 0x59, 0x5a, 0x63, 0x64, 0x65, 0x66, 0x67, 0x68, 0x69, 0x6a, 0x73, 0x74, 0x75,
	0x76, 0x77, 0x78, 0x79, 0x7a, 0x83, 0x84, 0x85, 0x86, 0x87, 0x88, 0x89, 0x8a, 0x92, 0x93, 0x94,
	0x95, 0x96, 0x97, 0x98, 0x99, 0x9a, 0xa2, 0xa3, 0xa4, 0xa5, 0xa6, 0xa7, 0xa8, 0xa9, 0xaa, 0xb2,
	0xb3, 0xb4, 0xb5, 0xb6, 0xb7, 0xb8, 0xb9, 0xba, 0xc2, 0xc3, 0xc4, 0xc5, 0xc6, 0xc7, 0xc8, 0xc9,
	0xca, 0xd2, 0xd3, 0xd4, 0xd5, 0xd6, 0xd7, 0xd8, 0xd9, 0xda, 0xe1, 0xe2, 0xe3, 0xe4, 0xe5, 0xe6,
	0xe7, 0xe8, 0xe9, 0xea, 0xf1, 0xf2, 0xf3, 0xf4, 0xf5, 0xf6, 0xf7, 0xf8, 0xf9, 0xfa, 0xff, 0xda,
	0x00, 0x0c, 0x03, 0x01, 0x00, 0x02, 0x11, 0x03, 0x11, 0x00, 0x3f, 0x00, 0xfb, 0x28, 0xff, 0xd9,
}

func TestCreateHandler_Create(t *testing.T) {
	tests := []struct {
		name           string
		req            *http.Request
		setUserID      bool
		needsS3        bool
		mockCreate     func(ctx context.Context, post *entity.GalleryPost) (*entity.GalleryPost, error)
		mockGetNickname func(ctx context.Context, userID uint) (string, error)
		wantStatus     int
		wantBodyHas    string
		checkErrReturn bool
	}{
		{
			name: "success: valid multipart request creates gallery returns 201",
			req:  newMultipartGalleryRequest(t, http.MethodPost, "/api/gallery", "test.jpg", minimalValidJPEG, "test caption"),
			setUserID: true,
			needsS3:    true,
			mockCreate: func(ctx context.Context, post *entity.GalleryPost) (*entity.GalleryPost, error) {
				post.ID = 1
				return post, nil
			},
			mockGetNickname: func(ctx context.Context, userID uint) (string, error) {
				return "testuser", nil
			},
			wantStatus:  http.StatusCreated,
			wantBodyHas: "gallery-001",
		},
		{
			name:           "no userID in context returns 401",
			req:            newMultipartGalleryRequest(t, http.MethodPost, "/api/gallery", "test.jpg", minimalValidJPEG, ""),
			setUserID:      false,
			wantStatus:     http.StatusUnauthorized,
			checkErrReturn: true,
		},
		{
			name:           "validation error: file is required returns 400",
			req:            newMultipartGalleryRequest(t, http.MethodPost, "/api/gallery", "", nil, "caption only"),
			setUserID:      true,
			wantStatus:     http.StatusBadRequest,
			checkErrReturn: true,
		},
		{
			name:           "validation error: unsupported file format returns 400",
			req:            newMultipartGalleryRequest(t, http.MethodPost, "/api/gallery", "file.txt", []byte("not an image"), ""),
			setUserID:      true,
			wantStatus:     http.StatusBadRequest,
			checkErrReturn: true,
		},
		{
			name:           "validation error: caption too long returns 400",
			req:            newMultipartGalleryRequest(t, http.MethodPost, "/api/gallery", "test.jpg", minimalValidJPEG, string(make([]byte, 501))),
			setUserID:      true,
			wantStatus:     http.StatusBadRequest,
			checkErrReturn: true,
		},
		{
			name: "usecase error: create fails returns 500",
			req:  newMultipartGalleryRequest(t, http.MethodPost, "/api/gallery", "test.jpg", minimalValidJPEG, ""),
			setUserID: true,
			needsS3:    true,
			mockCreate: func(ctx context.Context, post *entity.GalleryPost) (*entity.GalleryPost, error) {
				return nil, assert.AnError
			},
			wantStatus:     http.StatusInternalServerError,
			checkErrReturn: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Logf("Running case: %s", tt.name)
			if tt.needsS3 && aws.S3Client == nil {
				t.Skip("Skipping: AWS S3 not initialized - gallery create with file upload requires S3")
			}
			e := setupGalleryTestEcho()
			mockRepo := &mockGalleryCreateRepository{}
			if tt.mockCreate != nil {
				mockRepo.createFunc = tt.mockCreate
			}
			if tt.mockGetNickname != nil {
				mockRepo.getAuthorNickname = tt.mockGetNickname
			}
			uc := usecase.NewCreateUseCase(mockRepo, 10*time.Second)
			h := NewCreateHandler(uc)

			rec := httptest.NewRecorder()
			c := e.NewContext(tt.req, rec)
			if tt.setUserID {
				setupGalleryAuthContext(c)
			}

			err := h.Create(c)
			if tt.checkErrReturn {
				require.Error(t, err)
				if he, ok := err.(*echo.HTTPError); ok {
					assert.Equal(t, tt.wantStatus, he.Code)
					t.Logf("Error: code=%d message=%v", he.Code, he.Message)
				}
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.wantStatus, rec.Code)
			if tt.wantBodyHas != "" {
				assert.Contains(t, rec.Body.String(), tt.wantBodyHas)
			}
			t.Logf("Response: status=%d body=%s", rec.Code, rec.Body.String())
		})
	}
}
