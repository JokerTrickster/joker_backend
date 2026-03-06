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

type mockDeleteGalleryRepository struct {
	findByIDFunc func(ctx context.Context, id uint) (*entity.GalleryPost, error)
	deleteFunc   func(ctx context.Context, id uint) error
}

func (m *mockDeleteGalleryRepository) List(ctx context.Context, page, limit int) ([]entity.GalleryPost, int64, error) {
	return nil, 0, nil
}
func (m *mockDeleteGalleryRepository) FindByID(ctx context.Context, id uint) (*entity.GalleryPost, error) {
	if m.findByIDFunc != nil {
		return m.findByIDFunc(ctx, id)
	}
	return nil, nil
}
func (m *mockDeleteGalleryRepository) Create(ctx context.Context, post *entity.GalleryPost) (*entity.GalleryPost, error) {
	return nil, nil
}
func (m *mockDeleteGalleryRepository) Delete(ctx context.Context, id uint) error {
	if m.deleteFunc != nil {
		return m.deleteFunc(ctx, id)
	}
	return nil
}
func (m *mockDeleteGalleryRepository) IsLiked(ctx context.Context, userID, galleryID uint) (bool, error) {
	return false, nil
}
func (m *mockDeleteGalleryRepository) ToggleLike(ctx context.Context, userID, galleryID uint) (bool, int, error) {
	return false, 0, nil
}
func (m *mockDeleteGalleryRepository) ListComments(ctx context.Context, galleryID uint, page, limit int) ([]entity.GalleryComment, int64, error) {
	return nil, 0, nil
}
func (m *mockDeleteGalleryRepository) FindCommentByID(ctx context.Context, id uint) (*entity.GalleryComment, error) {
	return nil, nil
}
func (m *mockDeleteGalleryRepository) CreateComment(ctx context.Context, comment *entity.GalleryComment) (*entity.GalleryComment, error) {
	return nil, nil
}
func (m *mockDeleteGalleryRepository) DeleteComment(ctx context.Context, id uint) error {
	return nil
}
func (m *mockDeleteGalleryRepository) GetAuthorNickname(ctx context.Context, userID uint) (string, error) {
	return "", nil
}
func (m *mockDeleteGalleryRepository) GetAuthorInfo(ctx context.Context, userID uint) (string, *string, error) {
	return "", nil, nil
}
func (m *mockDeleteGalleryRepository) IsLikedBatch(ctx context.Context, userID uint, galleryIDs []uint) (map[uint]bool, error) {
	return map[uint]bool{}, nil
}
func (m *mockDeleteGalleryRepository) GetUserRole(ctx context.Context, userID uint) (string, error) {
	return "user", nil
}

var _ _interface.IGalleryRepository = (*mockDeleteGalleryRepository)(nil)

func TestDeleteUseCase_Delete_Success_Author(t *testing.T) {
	mockRepo := &mockDeleteGalleryRepository{
		findByIDFunc: func(ctx context.Context, id uint) (*entity.GalleryPost, error) {
			return &entity.GalleryPost{ID: 1, AuthorID: 1}, nil
		},
		deleteFunc: func(ctx context.Context, id uint) error {
			return nil
		},
	}
	uc := NewDeleteUseCase(mockRepo, 10*time.Second)
	ctx := context.Background()

	err := uc.Delete(ctx, 1, 1, "user")
	require.NoError(t, err)
	t.Logf("Delete success: author deletes own post")
}

func TestDeleteUseCase_Delete_Success_Admin(t *testing.T) {
	mockRepo := &mockDeleteGalleryRepository{
		findByIDFunc: func(ctx context.Context, id uint) (*entity.GalleryPost, error) {
			return &entity.GalleryPost{ID: 1, AuthorID: 999}, nil
		},
		deleteFunc: func(ctx context.Context, id uint) error {
			return nil
		},
	}
	uc := NewDeleteUseCase(mockRepo, 10*time.Second)
	ctx := context.Background()

	err := uc.Delete(ctx, 1, 1, "admin")
	require.NoError(t, err)
	t.Logf("Delete success: admin deletes other's post")
}

func TestDeleteUseCase_Delete_Forbidden(t *testing.T) {
	mockRepo := &mockDeleteGalleryRepository{
		findByIDFunc: func(ctx context.Context, id uint) (*entity.GalleryPost, error) {
			return &entity.GalleryPost{ID: 1, AuthorID: 999}, nil
		},
	}
	uc := NewDeleteUseCase(mockRepo, 10*time.Second)
	ctx := context.Background()

	err := uc.Delete(ctx, 1, 1, "user")
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrForbidden)
	t.Logf("Delete forbidden: user tries to delete other's post")
}

func TestDeleteUseCase_Delete_NotFound(t *testing.T) {
	mockRepo := &mockDeleteGalleryRepository{
		findByIDFunc: func(ctx context.Context, id uint) (*entity.GalleryPost, error) {
			return nil, assert.AnError
		},
	}
	uc := NewDeleteUseCase(mockRepo, 10*time.Second)
	ctx := context.Background()

	err := uc.Delete(ctx, 999, 1, "user")
	require.Error(t, err)
	t.Logf("Delete not found: %v", err)
}
