package usecase

import (
	"context"
	"testing"
	"time"

	"github.com/JokerTrickster/joker_backend/services/molandolanService/features/gallery/model/entity"
	_interface "github.com/JokerTrickster/joker_backend/services/molandolanService/features/gallery/model/interface"
	"github.com/JokerTrickster/joker_backend/services/molandolanService/features/gallery/model/request"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockCreateGalleryRepository struct {
	createFunc        func(ctx context.Context, post *entity.GalleryPost) (*entity.GalleryPost, error)
	getAuthorNickname func(ctx context.Context, userID uint) (string, error)
}

func (m *mockCreateGalleryRepository) List(ctx context.Context, page, limit int) ([]entity.GalleryPost, int64, error) {
	return nil, 0, nil
}
func (m *mockCreateGalleryRepository) FindByID(ctx context.Context, id uint) (*entity.GalleryPost, error) {
	return nil, nil
}
func (m *mockCreateGalleryRepository) Create(ctx context.Context, post *entity.GalleryPost) (*entity.GalleryPost, error) {
	if m.createFunc != nil {
		return m.createFunc(ctx, post)
	}
	post.ID = 1
	return post, nil
}
func (m *mockCreateGalleryRepository) Delete(ctx context.Context, id uint) error {
	return nil
}
func (m *mockCreateGalleryRepository) IsLiked(ctx context.Context, userID, galleryID uint) (bool, error) {
	return false, nil
}
func (m *mockCreateGalleryRepository) ToggleLike(ctx context.Context, userID, galleryID uint) (bool, int, error) {
	return false, 0, nil
}
func (m *mockCreateGalleryRepository) ListComments(ctx context.Context, galleryID uint, page, limit int) ([]entity.GalleryComment, int64, error) {
	return nil, 0, nil
}
func (m *mockCreateGalleryRepository) FindCommentByID(ctx context.Context, id uint) (*entity.GalleryComment, error) {
	return nil, nil
}
func (m *mockCreateGalleryRepository) CreateComment(ctx context.Context, comment *entity.GalleryComment) (*entity.GalleryComment, error) {
	return nil, nil
}
func (m *mockCreateGalleryRepository) DeleteComment(ctx context.Context, id uint) error {
	return nil
}
func (m *mockCreateGalleryRepository) GetAuthorNickname(ctx context.Context, userID uint) (string, error) {
	if m.getAuthorNickname != nil {
		return m.getAuthorNickname(ctx, userID)
	}
	return "testuser", nil
}
func (m *mockCreateGalleryRepository) GetAuthorInfo(ctx context.Context, userID uint) (string, *string, error) {
	if m.getAuthorNickname != nil {
		nick, _ := m.getAuthorNickname(ctx, userID)
		return nick, nil, nil
	}
	return "testuser", nil, nil
}
func (m *mockCreateGalleryRepository) IsLikedBatch(ctx context.Context, userID uint, galleryIDs []uint) (map[uint]bool, error) {
	return map[uint]bool{}, nil
}

var _ _interface.IGalleryRepository = (*mockCreateGalleryRepository)(nil)

func TestCreateUseCase_Create_Success(t *testing.T) {
	mockRepo := &mockCreateGalleryRepository{
		createFunc: func(ctx context.Context, post *entity.GalleryPost) (*entity.GalleryPost, error) {
			post.ID = 1
			return post, nil
		},
		getAuthorNickname: func(ctx context.Context, userID uint) (string, error) {
			return "author1", nil
		},
	}
	uc := NewCreateUseCase(mockRepo, 10*time.Second)
	ctx := context.Background()
	req := &request.ReqCreateGallery{
		MediaURL:     "https://example.com/media.jpg",
		ThumbnailURL: "https://example.com/thumb.jpg",
		MediaType:    "image",
		Caption:      "test caption",
	}

	res, err := uc.Create(ctx, 1, req)
	require.NoError(t, err)
	require.NotNil(t, res)
	assert.Equal(t, "gallery-001", res.ID)
	assert.Equal(t, "author1", res.Author.Nickname)
	assert.False(t, res.IsLiked)
	assert.Equal(t, "image", res.MediaType)
	assert.Equal(t, "https://example.com/media.jpg", res.MediaURL)
	t.Logf("Create success: id=%s author=%s", res.ID, res.Author.Nickname)
}

func TestCreateUseCase_Create_RepoError(t *testing.T) {
	mockRepo := &mockCreateGalleryRepository{
		createFunc: func(ctx context.Context, post *entity.GalleryPost) (*entity.GalleryPost, error) {
			return nil, assert.AnError
		},
	}
	uc := NewCreateUseCase(mockRepo, 10*time.Second)
	ctx := context.Background()
	req := &request.ReqCreateGallery{
		MediaURL:     "https://example.com/media.jpg",
		ThumbnailURL: "https://example.com/thumb.jpg",
		MediaType:    "image",
	}

	res, err := uc.Create(ctx, 1, req)
	require.Error(t, err)
	assert.Nil(t, res)
	t.Logf("Create repo error: %v", err)
}
