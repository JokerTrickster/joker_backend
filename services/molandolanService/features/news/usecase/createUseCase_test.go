package usecase

import (
	"context"
	"testing"
	"time"

	"github.com/JokerTrickster/joker_backend/services/molandolanService/features/news/model/entity"
	"github.com/JokerTrickster/joker_backend/services/molandolanService/features/news/model/request"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/mock"
)

type mockCreateNewsRepo struct {
	mock.Mock
}

func (m *mockCreateNewsRepo) List(ctx context.Context, page, limit int, category string) ([]entity.News, int64, error) {
	args := m.Called(ctx, page, limit, category)
	if args.Get(0) == nil {
		return nil, 0, args.Error(2)
	}
	return args.Get(0).([]entity.News), args.Get(1).(int64), args.Error(2)
}

func (m *mockCreateNewsRepo) FindByID(ctx context.Context, id uint) (*entity.News, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.News), args.Error(1)
}

func (m *mockCreateNewsRepo) Create(ctx context.Context, news *entity.News) (*entity.News, error) {
	args := m.Called(ctx, news)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.News), args.Error(1)
}

func (m *mockCreateNewsRepo) Update(ctx context.Context, id uint, updates map[string]interface{}) (*entity.News, error) {
	args := m.Called(ctx, id, updates)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.News), args.Error(1)
}

func (m *mockCreateNewsRepo) Delete(ctx context.Context, id uint) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func TestCreateUseCase_Create_Success(t *testing.T) {
	t.Log("Create: success - creates news and returns formatted response")
	mockRepo := new(mockCreateNewsRepo)
	uc := NewCreateUseCase(mockRepo, 5*time.Second)
	ctx := context.Background()
	req := &request.ReqCreateNews{
		Title: "Test", Summary: "Sum", Content: "Body", Thumbnail: "", Category: "cat", Date: "2024-01-15",
	}

	mockRepo.On("Create", mock.Anything, mock.AnythingOfType("*entity.News")).Run(func(args mock.Arguments) {
		n := args.Get(1).(*entity.News)
		n.ID = 5
	}).Return(&entity.News{
		ID: 5, Title: req.Title, Summary: req.Summary, Content: req.Content,
		Thumbnail: req.Thumbnail, Category: req.Category, Date: req.Date,
	}, nil)

	res, err := uc.Create(ctx, req)
	require.NoError(t, err)
	require.NotNil(t, res)
	assert.Equal(t, "news-005", res.ID)
	assert.Equal(t, req.Title, res.Title)
	assert.Equal(t, req.Content, res.Content)
	t.Logf("Create result: %+v", res)
	mockRepo.AssertExpectations(t)
}

func TestCreateUseCase_Create_RepoError(t *testing.T) {
	t.Log("Create: repo error propagates")
	mockRepo := new(mockCreateNewsRepo)
	uc := NewCreateUseCase(mockRepo, 5*time.Second)
	ctx := context.Background()
	req := &request.ReqCreateNews{
		Title: "T", Summary: "", Content: "C", Thumbnail: "", Category: "c", Date: "2024-01-01",
	}

	mockRepo.On("Create", mock.Anything, mock.AnythingOfType("*entity.News")).Return(nil, assert.AnError)

	res, err := uc.Create(ctx, req)
	assert.Error(t, err)
	assert.Nil(t, res)
	mockRepo.AssertExpectations(t)
}
