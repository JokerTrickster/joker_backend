package usecase

import (
	"context"
	"testing"
	"time"

	"github.com/JokerTrickster/joker_backend/services/morandoranService/features/gallery/model/entity"
	_interface "github.com/JokerTrickster/joker_backend/services/morandoranService/features/gallery/model/interface"
	"github.com/JokerTrickster/joker_backend/services/morandoranService/features/gallery/model/request"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockCommentCreateGalleryRepository struct {
	createCommentFunc func(ctx context.Context, comment *entity.GalleryComment) (*entity.GalleryComment, error)
	getAuthorNickname func(ctx context.Context, userID uint) (string, error)
}

func (m *mockCommentCreateGalleryRepository) List(ctx context.Context, page, limit int) ([]entity.GalleryPost, int64, error) {
	return nil, 0, nil
}
func (m *mockCommentCreateGalleryRepository) FindByID(ctx context.Context, id uint) (*entity.GalleryPost, error) {
	return nil, nil
}
func (m *mockCommentCreateGalleryRepository) Create(ctx context.Context, post *entity.GalleryPost) (*entity.GalleryPost, error) {
	return nil, nil
}
func (m *mockCommentCreateGalleryRepository) Delete(ctx context.Context, id uint) error {
	return nil
}
func (m *mockCommentCreateGalleryRepository) IsLiked(ctx context.Context, userID, galleryID uint) (bool, error) {
	return false, nil
}
func (m *mockCommentCreateGalleryRepository) ToggleLike(ctx context.Context, userID, galleryID uint) (bool, int, error) {
	return false, 0, nil
}
func (m *mockCommentCreateGalleryRepository) ListComments(ctx context.Context, galleryID uint, page, limit int) ([]entity.GalleryComment, int64, error) {
	return nil, 0, nil
}
func (m *mockCommentCreateGalleryRepository) FindCommentByID(ctx context.Context, id uint) (*entity.GalleryComment, error) {
	return nil, nil
}
func (m *mockCommentCreateGalleryRepository) CreateComment(ctx context.Context, comment *entity.GalleryComment) (*entity.GalleryComment, error) {
	if m.createCommentFunc != nil {
		return m.createCommentFunc(ctx, comment)
	}
	comment.ID = 1
	return comment, nil
}
func (m *mockCommentCreateGalleryRepository) DeleteComment(ctx context.Context, id uint) error {
	return nil
}
func (m *mockCommentCreateGalleryRepository) GetAuthorNickname(ctx context.Context, userID uint) (string, error) {
	if m.getAuthorNickname != nil {
		return m.getAuthorNickname(ctx, userID)
	}
	return "user", nil
}

var _ _interface.IGalleryRepository = (*mockCommentCreateGalleryRepository)(nil)

func TestCommentCreateUseCase_Create_Success(t *testing.T) {
	mockRepo := &mockCommentCreateGalleryRepository{
		createCommentFunc: func(ctx context.Context, comment *entity.GalleryComment) (*entity.GalleryComment, error) {
			comment.ID = 1
			return comment, nil
		},
		getAuthorNickname: func(ctx context.Context, userID uint) (string, error) {
			return "commenter", nil
		},
	}
	uc := NewCommentCreateUseCase(mockRepo, 10*time.Second)
	ctx := context.Background()
	req := &request.ReqCreateComment{Content: "nice post!"}

	res, err := uc.Create(ctx, 1, 1, req)
	require.NoError(t, err)
	require.NotNil(t, res)
	assert.Equal(t, "comment-001", res.ID)
	assert.Equal(t, "commenter", res.Author.Nickname)
	assert.Equal(t, "nice post!", res.Content)
	t.Logf("CommentCreate success: id=%s author=%s", res.ID, res.Author.Nickname)
}

func TestCommentCreateUseCase_Create_RepoError(t *testing.T) {
	mockRepo := &mockCommentCreateGalleryRepository{
		createCommentFunc: func(ctx context.Context, comment *entity.GalleryComment) (*entity.GalleryComment, error) {
			return nil, assert.AnError
		},
	}
	uc := NewCommentCreateUseCase(mockRepo, 10*time.Second)
	ctx := context.Background()
	req := &request.ReqCreateComment{Content: "hi"}

	res, err := uc.Create(ctx, 1, 1, req)
	require.Error(t, err)
	assert.Nil(t, res)
	t.Logf("CommentCreate repo error: %v", err)
}
