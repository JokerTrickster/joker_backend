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

type mockCommentListGalleryRepository struct {
	listCommentsFunc func(ctx context.Context, galleryID uint, page, limit int) ([]entity.GalleryComment, int64, error)
	getAuthorNickname func(ctx context.Context, userID uint) (string, error)
}

func (m *mockCommentListGalleryRepository) List(ctx context.Context, page, limit int) ([]entity.GalleryPost, int64, error) {
	return nil, 0, nil
}
func (m *mockCommentListGalleryRepository) FindByID(ctx context.Context, id uint) (*entity.GalleryPost, error) {
	return nil, nil
}
func (m *mockCommentListGalleryRepository) Create(ctx context.Context, post *entity.GalleryPost) (*entity.GalleryPost, error) {
	return nil, nil
}
func (m *mockCommentListGalleryRepository) Delete(ctx context.Context, id uint) error {
	return nil
}
func (m *mockCommentListGalleryRepository) IsLiked(ctx context.Context, userID, galleryID uint) (bool, error) {
	return false, nil
}
func (m *mockCommentListGalleryRepository) ToggleLike(ctx context.Context, userID, galleryID uint) (bool, int, error) {
	return false, 0, nil
}
func (m *mockCommentListGalleryRepository) ListComments(ctx context.Context, galleryID uint, page, limit int) ([]entity.GalleryComment, int64, error) {
	if m.listCommentsFunc != nil {
		return m.listCommentsFunc(ctx, galleryID, page, limit)
	}
	return nil, 0, nil
}
func (m *mockCommentListGalleryRepository) FindCommentByID(ctx context.Context, id uint) (*entity.GalleryComment, error) {
	return nil, nil
}
func (m *mockCommentListGalleryRepository) CreateComment(ctx context.Context, comment *entity.GalleryComment) (*entity.GalleryComment, error) {
	return nil, nil
}
func (m *mockCommentListGalleryRepository) DeleteComment(ctx context.Context, id uint) error {
	return nil
}
func (m *mockCommentListGalleryRepository) GetAuthorNickname(ctx context.Context, userID uint) (string, error) {
	if m.getAuthorNickname != nil {
		return m.getAuthorNickname(ctx, userID)
	}
	return "user", nil
}
func (m *mockCommentListGalleryRepository) GetAuthorInfo(ctx context.Context, userID uint) (string, *string, error) {
	if m.getAuthorNickname != nil {
		nick, _ := m.getAuthorNickname(ctx, userID)
		return nick, nil, nil
	}
	return "user", nil, nil
}
func (m *mockCommentListGalleryRepository) IsLikedBatch(ctx context.Context, userID uint, galleryIDs []uint) (map[uint]bool, error) {
	return map[uint]bool{}, nil
}
func (m *mockCommentListGalleryRepository) GetUserRole(ctx context.Context, userID uint) (string, error) {
	return "user", nil
}

var _ _interface.IGalleryRepository = (*mockCommentListGalleryRepository)(nil)

func TestCommentListUseCase_List_Success(t *testing.T) {
	mockRepo := &mockCommentListGalleryRepository{
		listCommentsFunc: func(ctx context.Context, galleryID uint, page, limit int) ([]entity.GalleryComment, int64, error) {
			return []entity.GalleryComment{
				{ID: 1, GalleryID: 1, AuthorID: 1, Content: "first"},
				{ID: 2, GalleryID: 1, AuthorID: 2, Content: "second"},
			}, 2, nil
		},
		getAuthorNickname: func(ctx context.Context, userID uint) (string, error) {
			if userID == 1 {
				return "user1", nil
			}
			return "user2", nil
		},
	}
	uc := NewCommentListUseCase(mockRepo, 10*time.Second)
	ctx := context.Background()
	req := &request.ReqListComment{Page: 1, Limit: 20}

	res, err := uc.List(ctx, 1, req)
	require.NoError(t, err)
	require.NotNil(t, res)
	assert.Len(t, res.Items, 2)
	assert.Equal(t, "comment-001", res.Items[0].ID)
	assert.Equal(t, "first", res.Items[0].Content)
	assert.Equal(t, 1, res.Pagination.TotalPages)
	assert.Equal(t, int64(2), res.Pagination.Total)
	t.Logf("CommentList success: items=%d total=%d", len(res.Items), res.Pagination.Total)
}

func TestCommentListUseCase_List_DefaultPagination(t *testing.T) {
	var capturedLimit int
	mockRepo := &mockCommentListGalleryRepository{
		listCommentsFunc: func(ctx context.Context, galleryID uint, page, limit int) ([]entity.GalleryComment, int64, error) {
			capturedLimit = limit
			return nil, 0, nil
		},
	}
	uc := NewCommentListUseCase(mockRepo, 10*time.Second)
	ctx := context.Background()
	req := &request.ReqListComment{Page: 0, Limit: 0}

	_, err := uc.List(ctx, 1, req)
	require.NoError(t, err)
	assert.Equal(t, 20, capturedLimit, "limit should default to 20")
	t.Logf("CommentList default limit: %d", capturedLimit)
}

func TestCommentListUseCase_List_RepoError(t *testing.T) {
	mockRepo := &mockCommentListGalleryRepository{
		listCommentsFunc: func(ctx context.Context, galleryID uint, page, limit int) ([]entity.GalleryComment, int64, error) {
			return nil, 0, assert.AnError
		},
	}
	uc := NewCommentListUseCase(mockRepo, 10*time.Second)
	ctx := context.Background()
	req := &request.ReqListComment{Page: 1, Limit: 10}

	res, err := uc.List(ctx, 1, req)
	require.Error(t, err)
	assert.Nil(t, res)
	t.Logf("CommentList repo error: %v", err)
}
