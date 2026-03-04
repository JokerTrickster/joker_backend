package usecase

import (
	"context"
	"testing"
	"time"

	"github.com/JokerTrickster/joker_backend/services/molandolanService/features/news/model/entity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/mock"
)

type mockDetailNewsRepo struct {
	mock.Mock
}

func (m *mockDetailNewsRepo) List(ctx context.Context, page, limit int, category string) ([]entity.News, int64, error) {
	args := m.Called(ctx, page, limit, category)
	if args.Get(0) == nil {
		return nil, 0, args.Error(2)
	}
	return args.Get(0).([]entity.News), args.Get(1).(int64), args.Error(2)
}

func (m *mockDetailNewsRepo) FindByID(ctx context.Context, id uint) (*entity.News, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.News), args.Error(1)
}

func (m *mockDetailNewsRepo) Create(ctx context.Context, news *entity.News) (*entity.News, error) {
	args := m.Called(ctx, news)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.News), args.Error(1)
}

func (m *mockDetailNewsRepo) Update(ctx context.Context, id uint, updates map[string]interface{}) (*entity.News, error) {
	args := m.Called(ctx, id, updates)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.News), args.Error(1)
}

func (m *mockDetailNewsRepo) Delete(ctx context.Context, id uint) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func TestDetailUseCase_Detail_Success(t *testing.T) {
	t.Log("Detail: success - returns formatted news")
	mockRepo := new(mockDetailNewsRepo)
	uc := NewDetailUseCase(mockRepo, 5*time.Second)
	ctx := context.Background()

	news := &entity.News{
		ID: 42, Title: "Title", Summary: "Sum", Content: "Body", Thumbnail: "x", Category: "cat", Date: "2024-01-15",
	}
	mockRepo.On("FindByID", mock.Anything, uint(42)).Return(news, nil)

	res, err := uc.Detail(ctx, 42)
	require.NoError(t, err)
	require.NotNil(t, res)
	assert.Equal(t, "news-042", res.ID)
	assert.Equal(t, "Title", res.Title)
	assert.Equal(t, "Body", res.Content)
	t.Logf("Detail result: %+v", res)
	mockRepo.AssertExpectations(t)
}

func TestDetailUseCase_Detail_NotFound(t *testing.T) {
	t.Log("Detail: not found returns error")
	mockRepo := new(mockDetailNewsRepo)
	uc := NewDetailUseCase(mockRepo, 5*time.Second)
	ctx := context.Background()

	mockRepo.On("FindByID", mock.Anything, uint(999)).Return(nil, assert.AnError)

	res, err := uc.Detail(ctx, 999)
	assert.Error(t, err)
	assert.Nil(t, res)
	mockRepo.AssertExpectations(t)
}
