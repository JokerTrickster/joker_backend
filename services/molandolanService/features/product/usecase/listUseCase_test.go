package usecase

import (
	"context"
	"testing"
	"time"

	"github.com/JokerTrickster/joker_backend/services/molandolanService/features/product/model/entity"
	"github.com/JokerTrickster/joker_backend/services/molandolanService/features/product/model/request"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/mock"
)

type mockListProductRepo struct {
	mock.Mock
}

func (m *mockListProductRepo) List(ctx context.Context, page, limit int, category string, inStock *bool) ([]entity.Product, int64, error) {
	args := m.Called(ctx, page, limit, category, inStock)
	if args.Get(0) == nil {
		return nil, 0, args.Error(2)
	}
	return args.Get(0).([]entity.Product), args.Get(1).(int64), args.Error(2)
}

func (m *mockListProductRepo) FindByID(ctx context.Context, id uint) (*entity.Product, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Product), args.Error(1)
}

func (m *mockListProductRepo) Create(ctx context.Context, product *entity.Product) (*entity.Product, error) {
	args := m.Called(ctx, product)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Product), args.Error(1)
}

func (m *mockListProductRepo) Update(ctx context.Context, id uint, updates map[string]interface{}) (*entity.Product, error) {
	args := m.Called(ctx, id, updates)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Product), args.Error(1)
}

func (m *mockListProductRepo) Delete(ctx context.Context, id uint) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func TestListProductUseCase_List_Success(t *testing.T) {
	t.Log("List: success")
	mockRepo := new(mockListProductRepo)
	uc := NewListUseCase(mockRepo, 5*time.Second)
	ctx := context.Background()
	req := &request.ReqListProduct{Page: 1, Limit: 10, Category: "merch", InStock: nil}

	items := []entity.Product{{ID: 1, Name: "P1", Price: 100, Description: "D", Image: "i", Category: "merch", InStock: true}}
	mockRepo.On("List", mock.Anything, 1, 10, "merch", (*bool)(nil)).Return(items, int64(1), nil)

	res, err := uc.List(ctx, req)
	require.NoError(t, err)
	require.NotNil(t, res)
	assert.Len(t, res.Items, 1)
	assert.Equal(t, "product-001", res.Items[0].ID)
	mockRepo.AssertExpectations(t)
}

func TestListProductUseCase_List_RepoError(t *testing.T) {
	t.Log("List: repo error")
	mockRepo := new(mockListProductRepo)
	uc := NewListUseCase(mockRepo, 5*time.Second)
	ctx := context.Background()
	req := &request.ReqListProduct{Page: 1, Limit: 10}

	mockRepo.On("List", mock.Anything, 1, 10, "", (*bool)(nil)).Return(nil, int64(0), assert.AnError)

	res, err := uc.List(ctx, req)
	assert.Error(t, err)
	assert.Nil(t, res)
	mockRepo.AssertExpectations(t)
}
