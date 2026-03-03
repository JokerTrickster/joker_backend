package usecase

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/JokerTrickster/joker_backend/services/morandoranService/features/auth/model/entity"
	"github.com/JokerTrickster/joker_backend/services/morandoranService/features/auth/model/request"
	_interface "github.com/JokerTrickster/joker_backend/services/morandoranService/features/auth/model/interface"
	"github.com/JokerTrickster/joker_backend/shared/jwt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

type mockLoginAuthRepository struct {
	findByEmailFunc func(ctx context.Context, email string) (*entity.MorandoranUser, error)
}

func (m *mockLoginAuthRepository) FindUserByEmail(ctx context.Context, email string) (*entity.MorandoranUser, error) {
	if m.findByEmailFunc != nil {
		return m.findByEmailFunc(ctx, email)
	}
	return nil, nil
}

func (m *mockLoginAuthRepository) FindUserByID(ctx context.Context, userID uint) (*entity.MorandoranUser, error) {
	return nil, nil
}

var _ _interface.IAuthRepository = (*mockLoginAuthRepository)(nil)

func initJWTForLoginUseCaseTest(t *testing.T) {
	t.Helper()
	os.Setenv("IS_LOCAL", "true")
	err := jwt.InitJwt()
	require.NoError(t, err, "JWT init should succeed")
}

func TestLoginUseCase_Login_Success(t *testing.T) {
	initJWTForLoginUseCaseTest(t)
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	require.NoError(t, err)

	mockRepo := &mockLoginAuthRepository{
		findByEmailFunc: func(ctx context.Context, email string) (*entity.MorandoranUser, error) {
			return &entity.MorandoranUser{
				ID:       1,
				Nickname: "testuser",
				Email:    "user@example.com",
				Password: string(hashedPassword),
				Role:     "user",
			}, nil
		},
	}
	uc := NewLoginUseCase(mockRepo, 10*time.Second)
	ctx := context.Background()
	req := &request.ReqLogin{Email: "user@example.com", Password: "password123"}

	res, err := uc.Login(ctx, req)
	require.NoError(t, err)
	require.NotNil(t, res)
	assert.NotEmpty(t, res.Token, "Token should be generated")
	assert.Equal(t, "user-001", res.User.ID)
	assert.Equal(t, "testuser", res.User.Nickname)
	assert.Equal(t, "user@example.com", res.User.Email)
	assert.Equal(t, "user", res.User.Role)
	t.Logf("Login success: token=%s... user=%s", res.Token[:20], res.User.Nickname)
}

func TestLoginUseCase_Login_UserNotFound(t *testing.T) {
	initJWTForLoginUseCaseTest(t)
	mockRepo := &mockLoginAuthRepository{
		findByEmailFunc: func(ctx context.Context, email string) (*entity.MorandoranUser, error) {
			return nil, assert.AnError
		},
	}
	uc := NewLoginUseCase(mockRepo, 10*time.Second)
	ctx := context.Background()
	req := &request.ReqLogin{Email: "nonexistent@example.com", Password: "password123"}

	res, err := uc.Login(ctx, req)
	require.Error(t, err)
	assert.Nil(t, res)
	assert.Contains(t, err.Error(), "invalid email or password")
	t.Logf("User not found error: %v", err)
}

func TestLoginUseCase_Login_WrongPassword(t *testing.T) {
	initJWTForLoginUseCaseTest(t)
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("correctpassword"), bcrypt.DefaultCost)
	require.NoError(t, err)

	mockRepo := &mockLoginAuthRepository{
		findByEmailFunc: func(ctx context.Context, email string) (*entity.MorandoranUser, error) {
			return &entity.MorandoranUser{
				ID:       1,
				Nickname: "testuser",
				Email:    "user@example.com",
				Password: string(hashedPassword),
				Role:     "user",
			}, nil
		},
	}
	uc := NewLoginUseCase(mockRepo, 10*time.Second)
	ctx := context.Background()
	req := &request.ReqLogin{Email: "user@example.com", Password: "wrongpassword"}

	res, err := uc.Login(ctx, req)
	require.Error(t, err)
	assert.Nil(t, res)
	assert.Contains(t, err.Error(), "invalid email or password")
	t.Logf("Wrong password error: %v", err)
}
