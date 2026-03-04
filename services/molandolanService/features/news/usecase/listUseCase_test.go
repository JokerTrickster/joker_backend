package usecase

import (
	"context"
	"testing"
	"time"

	"github.com/JokerTrickster/joker_backend/services/morandoranService/features/news/model/entity"
	"github.com/JokerTrickster/joker_backend/services/morandoranService/features/news/model/request"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/mock"
)

type mockListNewsRepo struct {
	mock.Mock
}

func (m *mockListNewsRepo) List(ctx context.Context, page, limit int, category string) ([]entity.News, int64, error) {
	args := m.Called(ctx, page, limit, category)
	if args.Get(0) == nil {
		return nil, 0, args.Error(2)
	}
	return args.Get(0).([]entity.News), args.Get(1).(int64), args.Error(2)
}

func (m *mockListNewsRepo) FindByID(ctx context.Context, id uint) (*entity.News, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.News), args.Error(1)
}

func (m *mockListNewsRepo) Create(ctx context.Context, news *entity.News) (*entity.News, error) {
	args := m.Called(ctx, news)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.News), args.Error(1)
}

func (m *mockListNewsRepo) Update(ctx context.Context, id uint, updates map[string]interface{}) (*entity.News, error) {
	args := m.Called(ctx, id, updates)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.News), args.Error(1)
}

func (m *mockListNewsRepo) Delete(ctx context.Context, id uint) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func TestListUseCase_List_Success(t *testing.T) {
	t.Log("List: success - returns paginated list")
	mockRepo := new(mockListNewsRepo)
	uc := NewListUseCase(mockRepo, 5*time.Second)
	ctx := context.Background()
	req := &request.ReqListNews{Page: 1, Limit: 10, Category: ""}

	items := []entity.News{
		{ID: 1, Title: "N1", Summary: "S1", Category: "cat", Date: "2024-01-15"},
		{ID: 2, Title: "N2", Summary: "S2", Category: "cat", Date: "2024-01-14"},
	}
	mockRepo.On("List", mock.Anything, 1, 10, "").Return(items, int64(2), nil)

	res, err := uc.List(ctx, req)
	require.NoError(t, err)
	require.NotNil(t, res)
	assert.Len(t, res.Items, 2)
	assert.Equal(t, "news-001", res.Items[0].ID)
	assert.Equal(t, "N1", res.Items[0].Title)
	assert.Equal(t, 1, res.Pagination.Page)
	assert.Equal(t, 10, res.Pagination.Limit)
	assert.Equal(t, int64(2), res.Pagination.Total)
	assert.Equal(t, 1, res.Pagination.TotalPages)
	t.Logf("List result: %+v", res)
	mockRepo.AssertExpectations(t)
}

func TestListUseCase_List_DefaultPagination(t *testing.T) {
	t.Log("List: page<1 and limit<1 use defaults")
	mockRepo := new(mockListNewsRepo)
	uc := NewListUseCase(mockRepo, 5*time.Second)
	ctx := context.Background()
	req := &request.ReqListNews{Page: 0, Limit: 0, Category: ""}

	mockRepo.On("List", mock.Anything, 1, 20, "").Return([]entity.News{}, int64(0), nil)

	res, err := uc.List(ctx, req)
	require.NoError(t, err)
	require.NotNil(t, res)
	assert.Equal(t, 1, res.Pagination.Page)
	assert.Equal(t, 20, res.Pagination.Limit)
	mockRepo.AssertExpectations(t)
}

func TestListUseCase_List_RepoError(t *testing.T) {
	t.Log("List: repo error propagates")
	mockRepo := new(mockListNewsRepo)
	uc := NewListUseCase(mockRepo, 5*time.Second)
	ctx := context.Background()
	req := &request.ReqListNews{Page: 1, Limit: 10, Category: "x"}

	mockRepo.On("List", mock.Anything, 1, 10, "x").Return(nil, int64(0), assert.AnError)

	res, err := uc.List(ctx, req)
	assert.Error(t, err)
	assert.Nil(t, res)
	mockRepo.AssertExpectations(t)
}
