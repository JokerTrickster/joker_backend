package usecase

import (
	"context"
	"testing"
	"time"

	"github.com/JokerTrickster/joker_backend/services/morandoranService/features/gallery/model/entity"
	_interface "github.com/JokerTrickster/joker_backend/services/morandoranService/features/gallery/model/interface"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockDetailGalleryRepository struct {
	findByIDFunc       func(ctx context.Context, id uint) (*entity.GalleryPost, error)
	getAuthorNickname  func(ctx context.Context, userID uint) (string, error)
	isLikedFunc        func(ctx context.Context, userID, galleryID uint) (bool, error)
}

func (m *mockDetailGalleryRepository) List(ctx context.Context, page, limit int) ([]entity.GalleryPost, int64, error) {
	return nil, 0, nil
}
func (m *mockDetailGalleryRepository) FindByID(ctx context.Context, id uint) (*entity.GalleryPost, error) {
	if m.findByIDFunc != nil {
		return m.findByIDFunc(ctx, id)
	}
	return nil, nil
}
func (m *mockDetailGalleryRepository) Create(ctx context.Context, post *entity.GalleryPost) (*entity.GalleryPost, error) {
	return nil, nil
}
func (m *mockDetailGalleryRepository) Delete(ctx context.Context, id uint) error {
	return nil
}
func (m *mockDetailGalleryRepository) IsLiked(ctx context.Context, userID, galleryID uint) (bool, error) {
	if m.isLikedFunc != nil {
		return m.isLikedFunc(ctx, userID, galleryID)
	}
	return false, nil
}
func (m *mockDetailGalleryRepository) ToggleLike(ctx context.Context, userID, galleryID uint) (bool, int, error) {
	return false, 0, nil
}
func (m *mockDetailGalleryRepository) ListComments(ctx context.Context, galleryID uint, page, limit int) ([]entity.GalleryComment, int64, error) {
	return nil, 0, nil
}
func (m *mockDetailGalleryRepository) FindCommentByID(ctx context.Context, id uint) (*entity.GalleryComment, error) {
	return nil, nil
}
func (m *mockDetailGalleryRepository) CreateComment(ctx context.Context, comment *entity.GalleryComment) (*entity.GalleryComment, error) {
	return nil, nil
}
func (m *mockDetailGalleryRepository) DeleteComment(ctx context.Context, id uint) error {
	return nil
}
func (m *mockDetailGalleryRepository) GetAuthorNickname(ctx context.Context, userID uint) (string, error) {
	if m.getAuthorNickname != nil {
		return m.getAuthorNickname(ctx, userID)
	}
	return "", nil
}

var _ _interface.IGalleryRepository = (*mockDetailGalleryRepository)(nil)

func TestDetailUseCase_Detail_Success(t *testing.T) {
	t.Log("TestDetailUseCase_Detail_Success: FindByID returns post, GetAuthorNickname returns author, IsLiked returns true -> verify response")
	now := time.Now()
	mockRepo := &mockDetailGalleryRepository{
		findByIDFunc: func(ctx context.Context, id uint) (*entity.GalleryPost, error) {
			return &entity.GalleryPost{
				ID:           5,
				AuthorID:     10,
				MediaType:    "image",
				MediaURL:     "https://example.com/media.jpg",
				ThumbnailURL: "https://example.com/thumb.jpg",
				Caption:      "test caption",
				LikeCount:    7,
				CommentCount: 2,
				CreatedAt:    now,
			}, nil
		},
		getAuthorNickname: func(ctx context.Context, userID uint) (string, error) {
			return "author", nil
		},
		isLikedFunc: func(ctx context.Context, userID, galleryID uint) (bool, error) {
			return true, nil
		},
	}
	uc := NewDetailUseCase(mockRepo, 10*time.Second)
	ctx := context.Background()
	userID := uint(1)

	res, err := uc.Detail(ctx, 5, &userID)
	require.NoError(t, err)
	require.NotNil(t, res)
	assert.Equal(t, "gallery-005", res.ID)
	assert.Equal(t, "user-010", res.Author.ID)
	assert.Equal(t, "author", res.Author.Nickname)
	assert.Equal(t, "image", res.MediaType)
	assert.Equal(t, "https://example.com/media.jpg", res.MediaURL)
	assert.Equal(t, 7, res.LikeCount)
	assert.Equal(t, 2, res.CommentCount)
	assert.True(t, res.IsLiked)
	assert.Equal(t, now, res.CreatedAt)
	t.Logf("Detail success: id=%s author=%s isLiked=%v", res.ID, res.Author.Nickname, res.IsLiked)
}

func TestDetailUseCase_Detail_NoUser(t *testing.T) {
	t.Log("TestDetailUseCase_Detail_NoUser: userID is nil -> IsLiked not called, isLiked=false")
	isLikedCalled := false
	mockRepo := &mockDetailGalleryRepository{
		findByIDFunc: func(ctx context.Context, id uint) (*entity.GalleryPost, error) {
			return &entity.GalleryPost{ID: 1, AuthorID: 1, MediaType: "image", MediaURL: "u", ThumbnailURL: "t"}, nil
		},
		getAuthorNickname: func(ctx context.Context, userID uint) (string, error) {
			return "author", nil
		},
		isLikedFunc: func(ctx context.Context, userID, galleryID uint) (bool, error) {
			isLikedCalled = true
			return true, nil
		},
	}
	uc := NewDetailUseCase(mockRepo, 10*time.Second)
	ctx := context.Background()

	res, err := uc.Detail(ctx, 1, nil)
	require.NoError(t, err)
	require.NotNil(t, res)
	assert.False(t, res.IsLiked)
	assert.False(t, isLikedCalled, "IsLiked should not be called when userID is nil")
	t.Logf("Detail no user: isLiked=%v isLikedCalled=%v", res.IsLiked, isLikedCalled)
}

func TestDetailUseCase_Detail_NotFound(t *testing.T) {
	t.Log("TestDetailUseCase_Detail_NotFound: FindByID returns error -> error")
	mockRepo := &mockDetailGalleryRepository{
		findByIDFunc: func(ctx context.Context, id uint) (*entity.GalleryPost, error) {
			return nil, assert.AnError
		},
	}
	uc := NewDetailUseCase(mockRepo, 10*time.Second)
	ctx := context.Background()
	userID := uint(1)

	res, err := uc.Detail(ctx, 999, &userID)
	require.Error(t, err)
	assert.Nil(t, res)
	assert.Equal(t, assert.AnError, err)
	t.Logf("Detail not found: %v", err)
}

func TestDetailUseCase_Detail_AuthorNicknameError(t *testing.T) {
	t.Log("TestDetailUseCase_Detail_AuthorNicknameError: GetAuthorNickname errors -> still returns (nickname defaults to empty)")
	mockRepo := &mockDetailGalleryRepository{
		findByIDFunc: func(ctx context.Context, id uint) (*entity.GalleryPost, error) {
			return &entity.GalleryPost{ID: 1, AuthorID: 1, MediaType: "image", MediaURL: "u", ThumbnailURL: "t"}, nil
		},
		getAuthorNickname: func(ctx context.Context, userID uint) (string, error) {
			return "", assert.AnError
		},
	}
	uc := NewDetailUseCase(mockRepo, 10*time.Second)
	ctx := context.Background()

	res, err := uc.Detail(ctx, 1, nil)
	require.NoError(t, err)
	require.NotNil(t, res)
	assert.Equal(t, "", res.Author.Nickname, "nickname should default to empty string when GetAuthorNickname errors")
	t.Logf("Detail author nickname error: nickname=%q", res.Author.Nickname)
}
