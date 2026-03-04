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

type mockDeleteProductRepo struct {
	mock.Mock
}

func (m *mockDeleteProductRepo) List(ctx context.Context, page, limit int, category string, inStock *bool) ([]entity.Product, int64, error) {
	args := m.Called(ctx, page, limit, category, inStock)
	if args.Get(0) == nil {
		return nil, 0, args.Error(2)
	}
	return args.Get(0).([]entity.Product), args.Get(1).(int64), args.Error(2)
}

func (m *mockDeleteProductRepo) FindByID(ctx context.Context, id uint) (*entity.Product, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Product), args.Error(1)
}

func (m *mockDeleteProductRepo) Create(ctx context.Context, product *entity.Product) (*entity.Product, error) {
	args := m.Called(ctx, product)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Product), args.Error(1)
}

func (m *mockDeleteProductRepo) Update(ctx context.Context, id uint, updates map[string]interface{}) (*entity.Product, error) {
	args := m.Called(ctx, id, updates)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Product), args.Error(1)
}

func (m *mockDeleteProductRepo) Delete(ctx context.Context, id uint) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func TestDeleteProductUseCase_Delete_Success(t *testing.T) {
	t.Log("Delete: success")
	mockRepo := new(mockDeleteProductRepo)
	uc := NewDeleteUseCase(mockRepo, 5*time.Second)
	ctx := context.Background()

	mockRepo.On("Delete", mock.Anything, uint(1)).Return(nil)

	err := uc.Delete(ctx, 1)
	require.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestDeleteProductUseCase_Delete_RepoError(t *testing.T) {
	t.Log("Delete: repo error")
	mockRepo := new(mockDeleteProductRepo)
	uc := NewDeleteUseCase(mockRepo, 5*time.Second)
	ctx := context.Background()

	mockRepo.On("Delete", mock.Anything, uint(999)).Return(assert.AnError)

	err := uc.Delete(ctx, 999)
	assert.Error(t, err)
	mockRepo.AssertExpectations(t)
}
