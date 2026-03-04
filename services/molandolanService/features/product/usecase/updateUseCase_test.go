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

type mockUpdateProductRepo struct {
	mock.Mock
}

func (m *mockUpdateProductRepo) List(ctx context.Context, page, limit int, category string, inStock *bool) ([]entity.Product, int64, error) {
	args := m.Called(ctx, page, limit, category, inStock)
	if args.Get(0) == nil {
		return nil, 0, args.Error(2)
	}
	return args.Get(0).([]entity.Product), args.Get(1).(int64), args.Error(2)
}

func (m *mockUpdateProductRepo) FindByID(ctx context.Context, id uint) (*entity.Product, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Product), args.Error(1)
}

func (m *mockUpdateProductRepo) Create(ctx context.Context, product *entity.Product) (*entity.Product, error) {
	args := m.Called(ctx, product)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Product), args.Error(1)
}

func (m *mockUpdateProductRepo) Update(ctx context.Context, id uint, updates map[string]interface{}) (*entity.Product, error) {
	args := m.Called(ctx, id, updates)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Product), args.Error(1)
}

func (m *mockUpdateProductRepo) Delete(ctx context.Context, id uint) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func TestUpdateProductUseCase_Update_Success(t *testing.T) {
	t.Log("Update: success")
	mockRepo := new(mockUpdateProductRepo)
	uc := NewUpdateUseCase(mockRepo, 5*time.Second)
	ctx := context.Background()
	name := "New Name"
	req := &request.ReqUpdateProduct{Name: &name}

	mockRepo.On("Update", mock.Anything, uint(1), mock.Anything).Return(&entity.Product{
		ID: 1, Name: "New Name", Price: 100, Description: "D", Image: "i", Category: "c", InStock: true,
	}, nil)

	res, err := uc.Update(ctx, 1, req)
	require.NoError(t, err)
	require.NotNil(t, res)
	assert.Equal(t, "product-001", res.ID)
	assert.Equal(t, "New Name", res.Name)
	mockRepo.AssertExpectations(t)
}

func TestUpdateProductUseCase_Update_RepoError(t *testing.T) {
	t.Log("Update: repo error")
	mockRepo := new(mockUpdateProductRepo)
	uc := NewUpdateUseCase(mockRepo, 5*time.Second)
	ctx := context.Background()
	name := "X"
	req := &request.ReqUpdateProduct{Name: &name}

	mockRepo.On("Update", mock.Anything, uint(999), mock.Anything).Return(nil, assert.AnError)

	res, err := uc.Update(ctx, 999, req)
	assert.Error(t, err)
	assert.Nil(t, res)
	mockRepo.AssertExpectations(t)
}
