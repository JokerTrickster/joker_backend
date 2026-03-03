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

type mockListGalleryRepository struct {
	listFunc func(ctx context.Context, page, limit int) ([]entity.GalleryPost, int64, error)
}

func (m *mockListGalleryRepository) List(ctx context.Context, page, limit int) ([]entity.GalleryPost, int64, error) {
	if m.listFunc != nil {
		return m.listFunc(ctx, page, limit)
	}
	return nil, 0, nil
}
func (m *mockListGalleryRepository) FindByID(ctx context.Context, id uint) (*entity.GalleryPost, error) {
	return nil, nil
}
func (m *mockListGalleryRepository) Create(ctx context.Context, post *entity.GalleryPost) (*entity.GalleryPost, error) {
	return nil, nil
}
func (m *mockListGalleryRepository) Delete(ctx context.Context, id uint) error {
	return nil
}
func (m *mockListGalleryRepository) IsLiked(ctx context.Context, userID, galleryID uint) (bool, error) {
	return false, nil
}
func (m *mockListGalleryRepository) ToggleLike(ctx context.Context, userID, galleryID uint) (bool, int, error) {
	return false, 0, nil
}
func (m *mockListGalleryRepository) ListComments(ctx context.Context, galleryID uint, page, limit int) ([]entity.GalleryComment, int64, error) {
	return nil, 0, nil
}
func (m *mockListGalleryRepository) FindCommentByID(ctx context.Context, id uint) (*entity.GalleryComment, error) {
	return nil, nil
}
func (m *mockListGalleryRepository) CreateComment(ctx context.Context, comment *entity.GalleryComment) (*entity.GalleryComment, error) {
	return nil, nil
}
func (m *mockListGalleryRepository) DeleteComment(ctx context.Context, id uint) error {
	return nil
}
func (m *mockListGalleryRepository) GetAuthorNickname(ctx context.Context, userID uint) (string, error) {
	return "", nil
}

var _ _interface.IGalleryRepository = (*mockListGalleryRepository)(nil)

func TestListUseCase_List_Success(t *testing.T) {
	t.Log("TestListUseCase_List_Success: returns items and pagination")
	now := time.Now()
	mockRepo := &mockListGalleryRepository{
		listFunc: func(ctx context.Context, page, limit int) ([]entity.GalleryPost, int64, error) {
			return []entity.GalleryPost{
				{ID: 1, MediaType: "image", ThumbnailURL: "t1", LikeCount: 5, CommentCount: 2, CreatedAt: now},
				{ID: 2, MediaType: "video", ThumbnailURL: "t2", LikeCount: 10, CommentCount: 1, CreatedAt: now},
			}, 2, nil
		},
	}
	uc := NewListUseCase(mockRepo, 10*time.Second)
	ctx := context.Background()
	req := &request.ReqListGallery{Page: 1, Limit: 10}

	res, err := uc.List(ctx, req)
	require.NoError(t, err)
	require.NotNil(t, res)
	require.Len(t, res.Items, 2)
	assert.Equal(t, "gallery-001", res.Items[0].ID)
	assert.Equal(t, "image", res.Items[0].MediaType)
	assert.Equal(t, 5, res.Items[0].LikeCount)
	assert.Equal(t, "gallery-002", res.Items[1].ID)
	assert.Equal(t, "video", res.Items[1].MediaType)
	assert.Equal(t, 1, res.Pagination.Page)
	assert.Equal(t, 10, res.Pagination.Limit)
	assert.Equal(t, int64(2), res.Pagination.Total)
	assert.Equal(t, 1, res.Pagination.TotalPages)
	t.Logf("List success: items=%d total=%d totalPages=%d", len(res.Items), res.Pagination.Total, res.Pagination.TotalPages)
}

func TestListUseCase_List_DefaultPageAndLimit(t *testing.T) {
	t.Log("TestListUseCase_List_DefaultPageAndLimit: page=0, limit=0 -> defaults to page=1, limit=30")
	capturedPage, capturedLimit := 0, 0
	mockRepo := &mockListGalleryRepository{
		listFunc: func(ctx context.Context, page, limit int) ([]entity.GalleryPost, int64, error) {
			capturedPage, capturedLimit = page, limit
			return nil, 0, nil
		},
	}
	uc := NewListUseCase(mockRepo, 10*time.Second)
	ctx := context.Background()
	req := &request.ReqListGallery{Page: 0, Limit: 0}

	res, err := uc.List(ctx, req)
	require.NoError(t, err)
	require.NotNil(t, res)
	assert.Equal(t, 1, capturedPage)
	assert.Equal(t, 30, capturedLimit)
	assert.Equal(t, 1, res.Pagination.Page)
	assert.Equal(t, 30, res.Pagination.Limit)
	t.Logf("List default: page=%d limit=%d", capturedPage, capturedLimit)
}

func TestListUseCase_List_EmptyResult(t *testing.T) {
	t.Log("TestListUseCase_List_EmptyResult: no items -> empty slice, pagination zeros")
	mockRepo := &mockListGalleryRepository{
		listFunc: func(ctx context.Context, page, limit int) ([]entity.GalleryPost, int64, error) {
			return []entity.GalleryPost{}, 0, nil
		},
	}
	uc := NewListUseCase(mockRepo, 10*time.Second)
	ctx := context.Background()
	req := &request.ReqListGallery{Page: 1, Limit: 10}

	res, err := uc.List(ctx, req)
	require.NoError(t, err)
	require.NotNil(t, res)
	assert.Empty(t, res.Items)
	assert.Equal(t, 1, res.Pagination.Page)
	assert.Equal(t, 10, res.Pagination.Limit)
	assert.Equal(t, int64(0), res.Pagination.Total)
	assert.Equal(t, 0, res.Pagination.TotalPages)
	t.Logf("List empty: items=%d totalPages=%d", len(res.Items), res.Pagination.TotalPages)
}

func TestListUseCase_List_RepoError(t *testing.T) {
	t.Log("TestListUseCase_List_RepoError: List returns error -> error")
	mockRepo := &mockListGalleryRepository{
		listFunc: func(ctx context.Context, page, limit int) ([]entity.GalleryPost, int64, error) {
			return nil, 0, assert.AnError
		},
	}
	uc := NewListUseCase(mockRepo, 10*time.Second)
	ctx := context.Background()
	req := &request.ReqListGallery{Page: 1, Limit: 10}

	res, err := uc.List(ctx, req)
	require.Error(t, err)
	assert.Nil(t, res)
	assert.Equal(t, assert.AnError, err)
	t.Logf("List repo error: %v", err)
}

func TestListUseCase_List_Pagination(t *testing.T) {
	t.Log("TestListUseCase_List_Pagination: total=25, limit=10 -> totalPages=3")
	now := time.Now()
	mockRepo := &mockListGalleryRepository{
		listFunc: func(ctx context.Context, page, limit int) ([]entity.GalleryPost, int64, error) {
			items := make([]entity.GalleryPost, 0, limit)
			start := (page-1)*limit + 1
			for i := 0; i < limit && int64(start+i) <= 25; i++ {
				items = append(items, entity.GalleryPost{ID: uint(start + i), CreatedAt: now})
			}
			return items, 25, nil
		},
	}
	uc := NewListUseCase(mockRepo, 10*time.Second)
	ctx := context.Background()
	req := &request.ReqListGallery{Page: 2, Limit: 10}

	res, err := uc.List(ctx, req)
	require.NoError(t, err)
	require.NotNil(t, res)
	assert.Len(t, res.Items, 10)
	assert.Equal(t, 2, res.Pagination.Page)
	assert.Equal(t, 10, res.Pagination.Limit)
	assert.Equal(t, int64(25), res.Pagination.Total)
	assert.Equal(t, 3, res.Pagination.TotalPages)
	t.Logf("List pagination: page=%d total=%d totalPages=%d", res.Pagination.Page, res.Pagination.Total, res.Pagination.TotalPages)
}
