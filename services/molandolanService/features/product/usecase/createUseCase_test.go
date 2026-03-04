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

type mockCreateProductRepo struct {
	mock.Mock
}

func (m *mockCreateProductRepo) List(ctx context.Context, page, limit int, category string, inStock *bool) ([]entity.Product, int64, error) {
	args := m.Called(ctx, page, limit, category, inStock)
	if args.Get(0) == nil {
		return nil, 0, args.Error(2)
	}
	return args.Get(0).([]entity.Product), args.Get(1).(int64), args.Error(2)
}

func (m *mockCreateProductRepo) FindByID(ctx context.Context, id uint) (*entity.Product, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Product), args.Error(1)
}

func (m *mockCreateProductRepo) Create(ctx context.Context, product *entity.Product) (*entity.Product, error) {
	args := m.Called(ctx, product)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Product), args.Error(1)
}

func (m *mockCreateProductRepo) Update(ctx context.Context, id uint, updates map[string]interface{}) (*entity.Product, error) {
	args := m.Called(ctx, id, updates)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Product), args.Error(1)
}

func (m *mockCreateProductRepo) Delete(ctx context.Context, id uint) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func TestCreateProductUseCase_Create_Success(t *testing.T) {
	t.Log("Create: success")
	mockRepo := new(mockCreateProductRepo)
	uc := NewCreateUseCase(mockRepo, 5*time.Second)
	ctx := context.Background()
	req := &request.ReqCreateProduct{
		Name: "P", Price: 100, Description: "D", Image: "i", Category: "c",
	}

	mockRepo.On("Create", mock.Anything, mock.AnythingOfType("*entity.Product")).Return(&entity.Product{
		ID: 7, Name: req.Name, Price: req.Price, Description: req.Description, Image: req.Image, Category: req.Category, InStock: true,
	}, nil)

	res, err := uc.Create(ctx, req)
	require.NoError(t, err)
	require.NotNil(t, res)
	assert.Equal(t, "product-007", res.ID)
	mockRepo.AssertExpectations(t)
}

func TestCreateProductUseCase_Create_RepoError(t *testing.T) {
	t.Log("Create: repo error")
	mockRepo := new(mockCreateProductRepo)
	uc := NewCreateUseCase(mockRepo, 5*time.Second)
	ctx := context.Background()
	req := &request.ReqCreateProduct{Name: "P", Price: 1, Description: "D", Image: "i", Category: "c"}

	mockRepo.On("Create", mock.Anything, mock.AnythingOfType("*entity.Product")).Return(nil, assert.AnError)

	res, err := uc.Create(ctx, req)
	assert.Error(t, err)
	assert.Nil(t, res)
	mockRepo.AssertExpectations(t)
}
