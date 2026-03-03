package usecase

import (
	"context"
	"testing"
	"time"

	"github.com/JokerTrickster/joker_backend/services/lottoDefenseService/features/towerDefense/model/entity"
	"github.com/JokerTrickster/joker_backend/services/lottoDefenseService/features/towerDefense/model/request"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type mockTDUserRepository struct {
	mock.Mock
}

func (m *mockTDUserRepository) Create(ctx context.Context, user *entity.TDUser) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *mockTDUserRepository) GetByID(ctx context.Context, id uint) (*entity.TDUser, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.TDUser), args.Error(1)
}

func (m *mockTDUserRepository) GetByEmail(ctx context.Context, email string) (*entity.TDUser, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.TDUser), args.Error(1)
}

func (m *mockTDUserRepository) GetByUsername(ctx context.Context, username string) (*entity.TDUser, error) {
	args := m.Called(ctx, username)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.TDUser), args.Error(1)
}

func (m *mockTDUserRepository) Update(ctx context.Context, user *entity.TDUser) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *mockTDUserRepository) UpdateLastLogin(ctx context.Context, id uint) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *mockTDUserRepository) GetStats(ctx context.Context, userID uint) (*entity.TDUserStats, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.TDUserStats), args.Error(1)
}

func (m *mockTDUserRepository) CreateStats(ctx context.Context, stats *entity.TDUserStats) error {
	args := m.Called(ctx, stats)
	return args.Error(0)
}

func (m *mockTDUserRepository) UpdateStats(ctx context.Context, stats *entity.TDUserStats) error {
	args := m.Called(ctx, stats)
	return args.Error(0)
}

func TestTDAuthUseCase_Register_Success(t *testing.T) {
	t.Log("Register: success - creates user, hashes password, generates JWT")
	mockRepo := new(mockTDUserRepository)
	uc := NewTDAuthUseCase(mockRepo, "test-secret", 5*time.Second)
	ctx := context.Background()
	req := &request.RegisterRequest{Username: "newuser", Email: "new@test.com", Password: "password123"}

	mockRepo.On("GetByEmail", ctx, req.Email).Return(nil, gorm.ErrRecordNotFound)
	mockRepo.On("GetByUsername", ctx, req.Username).Return(nil, gorm.ErrRecordNotFound)
	mockRepo.On("Create", ctx, mock.AnythingOfType("*entity.TDUser")).Run(func(args mock.Arguments) {
		u := args.Get(1).(*entity.TDUser)
		u.ID = 1
		assert.NotEmpty(t, u.PasswordHash)
		assert.NotEqual(t, "password123", u.PasswordHash)
	}).Return(nil)
	mockRepo.On("CreateStats", ctx, mock.AnythingOfType("*entity.TDUserStats")).Return(nil)

	resp, err := uc.Register(ctx, req)
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.NotEmpty(t, resp.Token)
	assert.Equal(t, "newuser", resp.User.Username)
	assert.Equal(t, "new@test.com", resp.User.Email)
	mockRepo.AssertExpectations(t)
}

func TestTDAuthUseCase_Register_EmailExists(t *testing.T) {
	t.Log("Register: email exists")
	mockRepo := new(mockTDUserRepository)
	uc := NewTDAuthUseCase(mockRepo, "test-secret", 5*time.Second)
	ctx := context.Background()
	req := &request.RegisterRequest{Username: "user", Email: "exists@test.com", Password: "password123"}

	mockRepo.On("GetByEmail", ctx, req.Email).Return(&entity.TDUser{ID: 1}, nil)

	resp, err := uc.Register(ctx, req)
	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "email already exists")
	mockRepo.AssertExpectations(t)
	mockRepo.AssertNotCalled(t, "Create")
}

func TestTDAuthUseCase_Register_UsernameExists(t *testing.T) {
	t.Log("Register: username exists")
	mockRepo := new(mockTDUserRepository)
	uc := NewTDAuthUseCase(mockRepo, "test-secret", 5*time.Second)
	ctx := context.Background()
	req := &request.RegisterRequest{Username: "existing", Email: "new@test.com", Password: "password123"}

	mockRepo.On("GetByEmail", ctx, req.Email).Return(nil, gorm.ErrRecordNotFound)
	mockRepo.On("GetByUsername", ctx, req.Username).Return(&entity.TDUser{ID: 1}, nil)

	resp, err := uc.Register(ctx, req)
	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "username already exists")
	mockRepo.AssertExpectations(t)
	mockRepo.AssertNotCalled(t, "Create")
}

func TestTDAuthUseCase_Login_Success(t *testing.T) {
	t.Log("Login: success")
	mockRepo := new(mockTDUserRepository)
	uc := NewTDAuthUseCase(mockRepo, "test-secret", 5*time.Second)
	ctx := context.Background()
	req := &request.LoginRequest{Email: "user@test.com", Password: "password123"}

	hashed, err := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	require.NoError(t, err)
	user := &entity.TDUser{ID: 1, Username: "user", Email: "user@test.com", PasswordHash: string(hashed)}
	mockRepo.On("GetByEmail", ctx, req.Email).Return(user, nil)
	mockRepo.On("UpdateLastLogin", ctx, user.ID).Return(nil)

	resp, err := uc.Login(ctx, req)
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.NotEmpty(t, resp.Token)
	mockRepo.AssertExpectations(t)
}

func TestTDAuthUseCase_Login_InvalidCredentials(t *testing.T) {
	t.Log("Login: invalid credentials - user not found")
	mockRepo := new(mockTDUserRepository)
	uc := NewTDAuthUseCase(mockRepo, "test-secret", 5*time.Second)
	ctx := context.Background()
	req := &request.LoginRequest{Email: "user@test.com", Password: "wrongpass"}

	mockRepo.On("GetByEmail", ctx, req.Email).Return(nil, gorm.ErrRecordNotFound)

	resp, err := uc.Login(ctx, req)
	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "invalid credentials")
	mockRepo.AssertExpectations(t)
}

func TestTDAuthUseCase_Login_InvalidPassword(t *testing.T) {
	t.Log("Login: invalid credentials - wrong password")
	mockRepo := new(mockTDUserRepository)
	uc := NewTDAuthUseCase(mockRepo, "test-secret", 5*time.Second)
	ctx := context.Background()
	req := &request.LoginRequest{Email: "user@test.com", Password: "wrongpass"}

	hashed, _ := bcrypt.GenerateFromPassword([]byte("correctpass"), bcrypt.DefaultCost)
	user := &entity.TDUser{ID: 1, Username: "user", Email: "user@test.com", PasswordHash: string(hashed)}
	mockRepo.On("GetByEmail", ctx, req.Email).Return(user, nil)

	resp, err := uc.Login(ctx, req)
	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "invalid credentials")
	mockRepo.AssertExpectations(t)
	mockRepo.AssertNotCalled(t, "UpdateLastLogin")
}

func TestTDAuthUseCase_GetUserInfo_Success(t *testing.T) {
	t.Log("GetUserInfo: success")
	mockRepo := new(mockTDUserRepository)
	uc := NewTDAuthUseCase(mockRepo, "test-secret", 5*time.Second)
	ctx := context.Background()
	userID := uint(1)

	mockRepo.On("GetByID", ctx, userID).Return(&entity.TDUser{ID: 1, Username: "user", Email: "user@test.com"}, nil)
	mockRepo.On("GetStats", ctx, userID).Return(&entity.TDUserStats{UserID: 1}, nil)

	resp, err := uc.GetUserInfo(ctx, userID)
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, "user", resp.User.Username)
	mockRepo.AssertExpectations(t)
}
