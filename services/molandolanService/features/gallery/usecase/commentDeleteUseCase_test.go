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

type mockCommentDeleteGalleryRepository struct {
	findCommentFunc func(ctx context.Context, id uint) (*entity.GalleryComment, error)
	deleteFunc      func(ctx context.Context, id uint) error
}

func (m *mockCommentDeleteGalleryRepository) List(ctx context.Context, page, limit int) ([]entity.GalleryPost, int64, error) {
	return nil, 0, nil
}
func (m *mockCommentDeleteGalleryRepository) FindByID(ctx context.Context, id uint) (*entity.GalleryPost, error) {
	return nil, nil
}
func (m *mockCommentDeleteGalleryRepository) Create(ctx context.Context, post *entity.GalleryPost) (*entity.GalleryPost, error) {
	return nil, nil
}
func (m *mockCommentDeleteGalleryRepository) Delete(ctx context.Context, id uint) error {
	return nil
}
func (m *mockCommentDeleteGalleryRepository) IsLiked(ctx context.Context, userID, galleryID uint) (bool, error) {
	return false, nil
}
func (m *mockCommentDeleteGalleryRepository) ToggleLike(ctx context.Context, userID, galleryID uint) (bool, int, error) {
	return false, 0, nil
}
func (m *mockCommentDeleteGalleryRepository) ListComments(ctx context.Context, galleryID uint, page, limit int) ([]entity.GalleryComment, int64, error) {
	return nil, 0, nil
}
func (m *mockCommentDeleteGalleryRepository) FindCommentByID(ctx context.Context, id uint) (*entity.GalleryComment, error) {
	if m.findCommentFunc != nil {
		return m.findCommentFunc(ctx, id)
	}
	return nil, nil
}
func (m *mockCommentDeleteGalleryRepository) CreateComment(ctx context.Context, comment *entity.GalleryComment) (*entity.GalleryComment, error) {
	return nil, nil
}
func (m *mockCommentDeleteGalleryRepository) DeleteComment(ctx context.Context, id uint) error {
	if m.deleteFunc != nil {
		return m.deleteFunc(ctx, id)
	}
	return nil
}
func (m *mockCommentDeleteGalleryRepository) GetAuthorNickname(ctx context.Context, userID uint) (string, error) {
	return "", nil
}
func (m *mockCommentDeleteGalleryRepository) GetAuthorInfo(ctx context.Context, userID uint) (string, *string, error) {
	return "", nil, nil
}
func (m *mockCommentDeleteGalleryRepository) IsLikedBatch(ctx context.Context, userID uint, galleryIDs []uint) (map[uint]bool, error) {
	return map[uint]bool{}, nil
}
func (m *mockCommentDeleteGalleryRepository) GetUserRole(ctx context.Context, userID uint) (string, error) {
	return "user", nil
}

var _ _interface.IGalleryRepository = (*mockCommentDeleteGalleryRepository)(nil)

func TestCommentDeleteUseCase_Delete_Success_Author(t *testing.T) {
	mockRepo := &mockCommentDeleteGalleryRepository{
		findCommentFunc: func(ctx context.Context, id uint) (*entity.GalleryComment, error) {
			return &entity.GalleryComment{ID: 1, AuthorID: 1, GalleryID: 1}, nil
		},
		deleteFunc: func(ctx context.Context, id uint) error {
			return nil
		},
	}
	uc := NewCommentDeleteUseCase(mockRepo, 10*time.Second)
	ctx := context.Background()

	err := uc.Delete(ctx, 1, 1, "user")
	require.NoError(t, err)
	t.Logf("CommentDelete success: author deletes own comment")
}

func TestCommentDeleteUseCase_Delete_Success_Admin(t *testing.T) {
	mockRepo := &mockCommentDeleteGalleryRepository{
		findCommentFunc: func(ctx context.Context, id uint) (*entity.GalleryComment, error) {
			return &entity.GalleryComment{ID: 1, AuthorID: 999, GalleryID: 1}, nil
		},
		deleteFunc: func(ctx context.Context, id uint) error {
			return nil
		},
	}
	uc := NewCommentDeleteUseCase(mockRepo, 10*time.Second)
	ctx := context.Background()

	err := uc.Delete(ctx, 1, 1, "admin")
	require.NoError(t, err)
	t.Logf("CommentDelete success: admin deletes other's comment")
}

func TestCommentDeleteUseCase_Delete_Forbidden(t *testing.T) {
	mockRepo := &mockCommentDeleteGalleryRepository{
		findCommentFunc: func(ctx context.Context, id uint) (*entity.GalleryComment, error) {
			return &entity.GalleryComment{ID: 1, AuthorID: 999, GalleryID: 1}, nil
		},
	}
	uc := NewCommentDeleteUseCase(mockRepo, 10*time.Second)
	ctx := context.Background()

	err := uc.Delete(ctx, 1, 1, "user")
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrForbidden)
	t.Logf("CommentDelete forbidden: user tries to delete other's comment")
}

func TestCommentDeleteUseCase_Delete_NotFound(t *testing.T) {
	mockRepo := &mockCommentDeleteGalleryRepository{
		findCommentFunc: func(ctx context.Context, id uint) (*entity.GalleryComment, error) {
			return nil, assert.AnError
		},
	}
	uc := NewCommentDeleteUseCase(mockRepo, 10*time.Second)
	ctx := context.Background()

	err := uc.Delete(ctx, 999, 1, "user")
	require.Error(t, err)
	t.Logf("CommentDelete not found: %v", err)
}
