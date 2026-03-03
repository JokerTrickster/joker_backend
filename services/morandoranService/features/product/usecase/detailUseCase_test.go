package usecase

import (
	"context"
	"testing"
	"time"

	"github.com/JokerTrickster/joker_backend/services/morandoranService/features/product/model/entity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/mock"
)

type mockDetailProductRepo struct {
	mock.Mock
}

func (m *mockDetailProductRepo) List(ctx context.Context, page, limit int, category string, inStock *bool) ([]entity.Product, int64, error) {
	args := m.Called(ctx, page, limit, category, inStock)
	if args.Get(0) == nil {
		return nil, 0, args.Error(2)
	}
	return args.Get(0).([]entity.Product), args.Get(1).(int64), args.Error(2)
}

func (m *mockDetailProductRepo) FindByID(ctx context.Context, id uint) (*entity.Product, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Product), args.Error(1)
}

func (m *mockDetailProductRepo) Create(ctx context.Context, product *entity.Product) (*entity.Product, error) {
	args := m.Called(ctx, product)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Product), args.Error(1)
}

func (m *mockDetailProductRepo) Update(ctx context.Context, id uint, updates map[string]interface{}) (*entity.Product, error) {
	args := m.Called(ctx, id, updates)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Product), args.Error(1)
}

func (m *mockDetailProductRepo) Delete(ctx context.Context, id uint) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func TestDetailProductUseCase_Detail_Success(t *testing.T) {
	t.Log("Detail: success")
	mockRepo := new(mockDetailProductRepo)
	uc := NewDetailUseCase(mockRepo, 5*time.Second)
	ctx := context.Background()

	mockRepo.On("FindByID", mock.Anything, uint(1)).Return(&entity.Product{
		ID: 1, Name: "P", Price: 100, Description: "D", Image: "i", Category: "c", InStock: true,
	}, nil)

	res, err := uc.Detail(ctx, 1)
	require.NoError(t, err)
	require.NotNil(t, res)
	assert.Equal(t, "product-001", res.ID)
	mockRepo.AssertExpectations(t)
}

func TestDetailProductUseCase_Detail_NotFound(t *testing.T) {
	t.Log("Detail: not found")
	mockRepo := new(mockDetailProductRepo)
	uc := NewDetailUseCase(mockRepo, 5*time.Second)
	ctx := context.Background()

	mockRepo.On("FindByID", mock.Anything, uint(999)).Return(nil, assert.AnError)

	res, err := uc.Detail(ctx, 999)
	assert.Error(t, err)
	assert.Nil(t, res)
	mockRepo.AssertExpectations(t)
}
