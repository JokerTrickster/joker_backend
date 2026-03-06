package usecase

import (
	"context"
	"testing"
	"time"

	"github.com/JokerTrickster/joker_backend/services/molandolanService/features/gallery/model/entity"
	_interface "github.com/JokerTrickster/joker_backend/services/molandolanService/features/gallery/model/interface"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockLikeGalleryRepository struct {
	toggleLikeFunc func(ctx context.Context, userID, galleryID uint) (bool, int, error)
}

func (m *mockLikeGalleryRepository) List(ctx context.Context, page, limit int) ([]entity.GalleryPost, int64, error) {
	return nil, 0, nil
}
func (m *mockLikeGalleryRepository) FindByID(ctx context.Context, id uint) (*entity.GalleryPost, error) {
	return nil, nil
}
func (m *mockLikeGalleryRepository) Create(ctx context.Context, post *entity.GalleryPost) (*entity.GalleryPost, error) {
	return nil, nil
}
func (m *mockLikeGalleryRepository) Delete(ctx context.Context, id uint) error {
	return nil
}
func (m *mockLikeGalleryRepository) IsLiked(ctx context.Context, userID, galleryID uint) (bool, error) {
	return false, nil
}
func (m *mockLikeGalleryRepository) ToggleLike(ctx context.Context, userID, galleryID uint) (bool, int, error) {
	if m.toggleLikeFunc != nil {
		return m.toggleLikeFunc(ctx, userID, galleryID)
	}
	return false, 0, nil
}
func (m *mockLikeGalleryRepository) ListComments(ctx context.Context, galleryID uint, page, limit int) ([]entity.GalleryComment, int64, error) {
	return nil, 0, nil
}
func (m *mockLikeGalleryRepository) FindCommentByID(ctx context.Context, id uint) (*entity.GalleryComment, error) {
	return nil, nil
}
func (m *mockLikeGalleryRepository) CreateComment(ctx context.Context, comment *entity.GalleryComment) (*entity.GalleryComment, error) {
	return nil, nil
}
func (m *mockLikeGalleryRepository) DeleteComment(ctx context.Context, id uint) error {
	return nil
}
func (m *mockLikeGalleryRepository) GetAuthorNickname(ctx context.Context, userID uint) (string, error) {
	return "", nil
}
func (m *mockLikeGalleryRepository) GetAuthorInfo(ctx context.Context, userID uint) (string, *string, error) {
	return "", nil, nil
}
func (m *mockLikeGalleryRepository) IsLikedBatch(ctx context.Context, userID uint, galleryIDs []uint) (map[uint]bool, error) {
	return map[uint]bool{}, nil
}
func (m *mockLikeGalleryRepository) GetUserRole(ctx context.Context, userID uint) (string, error) {
	return "user", nil
}

var _ _interface.IGalleryRepository = (*mockLikeGalleryRepository)(nil)

func TestLikeUseCase_ToggleLike_LikeSuccess(t *testing.T) {
	t.Log("TestLikeUseCase_ToggleLike_LikeSuccess: ToggleLike returns (true, 5, nil) -> verify")
	mockRepo := &mockLikeGalleryRepository{
		toggleLikeFunc: func(ctx context.Context, userID, galleryID uint) (bool, int, error) {
			return true, 5, nil
		},
	}
	uc := NewLikeUseCase(mockRepo, 10*time.Second)
	ctx := context.Background()

	res, err := uc.ToggleLike(ctx, 1, 10)
	require.NoError(t, err)
	require.NotNil(t, res)
	assert.True(t, res.IsLiked)
	assert.Equal(t, 5, res.LikeCount)
	t.Logf("ToggleLike like success: isLiked=%v likeCount=%d", res.IsLiked, res.LikeCount)
}

func TestLikeUseCase_ToggleLike_UnlikeSuccess(t *testing.T) {
	t.Log("TestLikeUseCase_ToggleLike_UnlikeSuccess: ToggleLike returns (false, 4, nil) -> verify")
	mockRepo := &mockLikeGalleryRepository{
		toggleLikeFunc: func(ctx context.Context, userID, galleryID uint) (bool, int, error) {
			return false, 4, nil
		},
	}
	uc := NewLikeUseCase(mockRepo, 10*time.Second)
	ctx := context.Background()

	res, err := uc.ToggleLike(ctx, 1, 10)
	require.NoError(t, err)
	require.NotNil(t, res)
	assert.False(t, res.IsLiked)
	assert.Equal(t, 4, res.LikeCount)
	t.Logf("ToggleLike unlike success: isLiked=%v likeCount=%d", res.IsLiked, res.LikeCount)
}

func TestLikeUseCase_ToggleLike_RepoError(t *testing.T) {
	t.Log("TestLikeUseCase_ToggleLike_RepoError: ToggleLike returns error -> error")
	mockRepo := &mockLikeGalleryRepository{
		toggleLikeFunc: func(ctx context.Context, userID, galleryID uint) (bool, int, error) {
			return false, 0, assert.AnError
		},
	}
	uc := NewLikeUseCase(mockRepo, 10*time.Second)
	ctx := context.Background()

	res, err := uc.ToggleLike(ctx, 1, 10)
	require.Error(t, err)
	assert.Nil(t, res)
	assert.Equal(t, assert.AnError, err)
	t.Logf("ToggleLike repo error: %v", err)
}
