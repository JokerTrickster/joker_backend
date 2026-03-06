package usecase

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"testing"
	"time"

	_interface "github.com/JokerTrickster/joker_backend/services/authService/features/auth/model/interface"
	"github.com/JokerTrickster/joker_backend/services/authService/features/auth/model/request"
	"github.com/JokerTrickster/joker_backend/services/authService/features/auth/repository"
	"github.com/JokerTrickster/joker_backend/shared/db/mysql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockLogoutRepository struct {
	DeleteTokenByUserIDFunc func(ctx context.Context, userID uint) error
}

func (m *mockLogoutRepository) DeleteTokenByUserID(ctx context.Context, userID uint) error {
	if m.DeleteTokenByUserIDFunc != nil {
		return m.DeleteTokenByUserIDFunc(ctx, userID)
	}
	return nil
}

func TestLogoutAuthUseCase_Success(t *testing.T) {
	db := setupTestDB(t)
	requireTokensTable(t, db)
	initJWTForTest(t)

	// Sign up a user
	signupRepo := repository.NewSignupAuthRepository(db)
	signupUC := NewSignupAuthUseCase(signupRepo, 10*time.Second)

	ctx := context.Background()
	email := "logout-test-" + fmt.Sprintf("%d_%d", time.Now().UnixNano(), rand.Intn(100000)) + "@example.com"

	signupReq := &request.ReqSignUp{
		Email:       email,
		Password:    "password123",
		ServiceType: "game",
		Name:        "Logout Test User",
	}

	signupRes, err := signupUC.Signup(ctx, signupReq)
	require.NoError(t, err, "Signup should succeed")

	// Find user ID
	var user mysql.Users
	err = db.Where("email = ? AND provider = ?", email, "game").First(&user).Error
	require.NoError(t, err, "User should be found")
	userID := uint(user.ID)

	t.Logf("User created: id=%d, email=%s", userID, email)

	// Store a token in DB (simulating what refresh does)
	refreshRepo := repository.NewRefreshTokenAuthRepository(db)
	tokenDTO := createTokenDTO(userID, signupRes.AccessToken, signupRes.RefreshToken)
	err = refreshRepo.CreateToken(ctx, tokenDTO)
	require.NoError(t, err, "Token storage should succeed")

	// Verify token exists
	var tokenCount int64
	err = db.Model(&mysql.Tokens{}).Where("user_id = ?", userID).Count(&tokenCount).Error
	require.NoError(t, err)
	assert.Equal(t, int64(1), tokenCount, "Token should exist before logout")

	t.Logf("Token stored for user %d, count=%d", userID, tokenCount)

	// Logout
	logoutRepo := repository.NewLogoutAuthRepository(db)
	logoutUC := NewLogoutAuthUseCase(logoutRepo, 10*time.Second)

	err = logoutUC.Logout(ctx, userID)
	require.NoError(t, err, "Logout should succeed")

	// Verify token was deleted
	err = db.Model(&mysql.Tokens{}).Where("user_id = ?", userID).Count(&tokenCount).Error
	require.NoError(t, err)
	assert.Equal(t, int64(0), tokenCount, "Token should be deleted after logout")

	t.Logf("Logout succeeded: tokens for user %d = %d", userID, tokenCount)
}

func TestLogoutAuthUseCase_NoTokens(t *testing.T) {
	db := setupTestDB(t)
	requireTokensTable(t, db)
	initJWTForTest(t)

	logoutRepo := repository.NewLogoutAuthRepository(db)
	logoutUC := NewLogoutAuthUseCase(logoutRepo, 10*time.Second)

	ctx := context.Background()

	// Logout with a user that has no tokens - should return error
	err := logoutUC.Logout(ctx, 99999)
	assert.Error(t, err, "Logout with no tokens should return error")

	t.Logf("No-tokens logout correctly handled: %v", err)
}

func TestNewLogoutAuthUseCase(t *testing.T) {
	repo := &mockLogoutRepository{}
	uc := NewLogoutAuthUseCase(repo, 5*time.Second).(*LogoutAuthUseCase)
	require.NotNil(t, uc)
	assert.Equal(t, repo, uc.Repository)
	assert.Equal(t, 5*time.Second, uc.ContextTimeout)
	t.Logf("NewLogoutAuthUseCase sets Repository and ContextTimeout correctly")
}

func TestLogoutAuthUseCase_ImplementsInterface(t *testing.T) {
	repo := &mockLogoutRepository{}
	uc := NewLogoutAuthUseCase(repo, 10*time.Second)
	var _ _interface.ILogoutAuthUseCase = uc
	t.Logf("LogoutAuthUseCase implements ILogoutAuthUseCase")
}

func TestLogoutAuthUseCase_Logout_RepoError(t *testing.T) {
	repoErr := errors.New("database error")
	repo := &mockLogoutRepository{
		DeleteTokenByUserIDFunc: func(ctx context.Context, userID uint) error {
			return repoErr
		},
	}
	uc := NewLogoutAuthUseCase(repo, 10*time.Second)
	ctx := context.Background()

	err := uc.Logout(ctx, 123)
	require.Error(t, err)
	assert.ErrorIs(t, err, repoErr)
	t.Logf("Logout with repo error correctly propagated: %v", err)
}

func TestLogoutAuthUseCase_Logout_SuccessWithMock(t *testing.T) {
	repo := &mockLogoutRepository{
		DeleteTokenByUserIDFunc: func(ctx context.Context, userID uint) error {
			return nil
		},
	}
	uc := NewLogoutAuthUseCase(repo, 10*time.Second)
	ctx := context.Background()

	err := uc.Logout(ctx, 42)
	require.NoError(t, err)
	t.Logf("Logout with mock success returns nil")
}
