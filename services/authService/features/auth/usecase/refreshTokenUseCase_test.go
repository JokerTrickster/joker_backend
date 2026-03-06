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
	"github.com/JokerTrickster/joker_backend/shared/jwt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockRefreshTokenRepository struct {
	CreateTokenFunc                    func(ctx context.Context, tokenDTO *mysql.Tokens) error
	FindOneByUserIDAndDeleteTokenFunc  func(ctx context.Context, userID uint) error
}

func (m *mockRefreshTokenRepository) CreateToken(ctx context.Context, tokenDTO *mysql.Tokens) error {
	if m.CreateTokenFunc != nil {
		return m.CreateTokenFunc(ctx, tokenDTO)
	}
	return nil
}

func (m *mockRefreshTokenRepository) FindOneByUserIDAndDeleteToken(ctx context.Context, userID uint) error {
	if m.FindOneByUserIDAndDeleteTokenFunc != nil {
		return m.FindOneByUserIDAndDeleteTokenFunc(ctx, userID)
	}
	return nil
}

func TestRefreshTokenUseCase_Success(t *testing.T) {
	db := setupTestDB(t)
	requireTokensTable(t, db)
	initJWTForTest(t)

	// Sign up a user to get initial tokens
	signupRepo := repository.NewSignupAuthRepository(db)
	signupUC := NewSignupAuthUseCase(signupRepo, 10*time.Second)

	ctx := context.Background()
	email := "refresh-test-" + fmt.Sprintf("%d_%d", time.Now().UnixNano(), rand.Intn(100000)) + "@example.com"

	signupReq := &request.ReqSignUp{
		Email:       email,
		Password:    "password123",
		ServiceType: "game",
		Name:        "Refresh Test User",
	}

	signupRes, err := signupUC.Signup(ctx, signupReq)
	require.NoError(t, err, "Signup should succeed")

	t.Logf("Signup tokens: access=%s..., refresh=%s...",
		signupRes.AccessToken[:20], signupRes.RefreshToken[:20])

	// Generate a valid refresh token for the user
	_, _, refreshToken, _, err := jwt.GenerateToken(email, 1)
	require.NoError(t, err, "Token generation should succeed")

	// Use refresh token to get new tokens
	refreshRepo := repository.NewRefreshTokenAuthRepository(db)
	refreshUC := NewRefreshTokenUseCase(refreshRepo, 10*time.Second)

	refreshReq := &request.ReqRefreshToken{
		RefreshToken: refreshToken,
	}

	res, err := refreshUC.RefreshToken(ctx, refreshReq)
	require.NoError(t, err, "Refresh should succeed")
	assert.NotEmpty(t, res.AccessToken, "New AccessToken should be set")
	assert.NotEmpty(t, res.RefreshToken, "New RefreshToken should be set")
	assert.NotEqual(t, refreshToken, res.RefreshToken, "New refresh token should differ from old one")

	t.Logf("Refresh succeeded: newAccess=%s..., newRefresh=%s...",
		res.AccessToken[:20], res.RefreshToken[:20])
}

func TestRefreshTokenUseCase_InvalidToken(t *testing.T) {
	db := setupTestDB(t)
	requireTokensTable(t, db)
	initJWTForTest(t)

	refreshRepo := repository.NewRefreshTokenAuthRepository(db)
	refreshUC := NewRefreshTokenUseCase(refreshRepo, 10*time.Second)

	ctx := context.Background()
	refreshReq := &request.ReqRefreshToken{
		RefreshToken: "invalid.token.string",
	}

	_, err := refreshUC.RefreshToken(ctx, refreshReq)
	assert.Error(t, err, "Refresh with invalid token should fail")

	t.Logf("Invalid token correctly rejected: %v", err)
}

func TestNewRefreshTokenUseCase(t *testing.T) {
	repo := &mockRefreshTokenRepository{}
	uc := NewRefreshTokenUseCase(repo, 5*time.Second).(*RefreshTokenUseCase)
	require.NotNil(t, uc)
	assert.Equal(t, repo, uc.Repository)
	assert.Equal(t, 5*time.Second, uc.ContextTimeout)
	t.Logf("NewRefreshTokenUseCase sets Repository and ContextTimeout correctly")
}

func TestRefreshTokenUseCase_ImplementsInterface(t *testing.T) {
	repo := &mockRefreshTokenRepository{}
	uc := NewRefreshTokenUseCase(repo, 10*time.Second)
	var _ _interface.IRefreshTokenUseCase = uc
	t.Logf("RefreshTokenUseCase implements IRefreshTokenUseCase")
}

func TestRefreshTokenUseCase_RefreshToken_CreateTokenError(t *testing.T) {
	initJWTForTest(t)
	repoErr := errors.New("failed to insert token")
	repo := &mockRefreshTokenRepository{
		CreateTokenFunc: func(ctx context.Context, tokenDTO *mysql.Tokens) error {
			return repoErr
		},
		FindOneByUserIDAndDeleteTokenFunc: func(ctx context.Context, userID uint) error {
			return nil
		},
	}
	uc := NewRefreshTokenUseCase(repo, 10*time.Second)
	ctx := context.Background()

	_, _, refreshToken, _, err := jwt.GenerateToken("test@example.com", 1)
	require.NoError(t, err)

	req := &request.ReqRefreshToken{RefreshToken: refreshToken}
	res, err := uc.RefreshToken(ctx, req)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to store refresh token")
	assert.ErrorIs(t, err, repoErr)
	assert.Empty(t, res.AccessToken)
	assert.Empty(t, res.RefreshToken)
	t.Logf("RefreshToken with CreateToken error correctly propagated: %v", err)
}

func TestRefreshTokenUseCase_RefreshToken_SuccessWithMock(t *testing.T) {
	initJWTForTest(t)

	var createCalled, deleteCalled bool
	var capturedTokenDTO *mysql.Tokens
	repo := &mockRefreshTokenRepository{
		CreateTokenFunc: func(ctx context.Context, tokenDTO *mysql.Tokens) error {
			createCalled = true
			capturedTokenDTO = tokenDTO
			return nil
		},
		FindOneByUserIDAndDeleteTokenFunc: func(ctx context.Context, userID uint) error {
			deleteCalled = true
			assert.Equal(t, uint(7), userID)
			return nil
		},
	}
	uc := NewRefreshTokenUseCase(repo, 10*time.Second)
	ctx := context.Background()

	_, _, refreshToken, _, err := jwt.GenerateToken("refresh-unit@example.com", 7)
	require.NoError(t, err)

	req := &request.ReqRefreshToken{RefreshToken: refreshToken}
	res, err := uc.RefreshToken(ctx, req)
	require.NoError(t, err)
	require.True(t, createCalled)
	require.True(t, deleteCalled)
	assert.NotNil(t, capturedTokenDTO)
	assert.Equal(t, uint(7), capturedTokenDTO.UserID)
	assert.NotEmpty(t, res.AccessToken)
	assert.NotEmpty(t, res.RefreshToken)
	// Note: new refresh token may equal old if generated in same second; both are valid JWTs
	t.Logf("RefreshToken success with mock: new tokens generated")
}

func TestRefreshTokenUseCase_RefreshToken_InvalidTokenWithMock(t *testing.T) {
	initJWTForTest(t)

	repo := &mockRefreshTokenRepository{}
	uc := NewRefreshTokenUseCase(repo, 10*time.Second)
	ctx := context.Background()
	req := &request.ReqRefreshToken{RefreshToken: "invalid.not.a.valid.jwt"}

	res, err := uc.RefreshToken(ctx, req)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid or expired refresh token")
	assert.Empty(t, res.AccessToken)
	assert.Empty(t, res.RefreshToken)
	t.Logf("Invalid token correctly rejected with mock repo: %v", err)
}

func TestRefreshTokenUseCase_RefreshToken_DeleteTokenErrorStillSucceeds(t *testing.T) {
	initJWTForTest(t)
	createCalled := false
	repo := &mockRefreshTokenRepository{
		CreateTokenFunc: func(ctx context.Context, tokenDTO *mysql.Tokens) error {
			createCalled = true
			return nil
		},
		FindOneByUserIDAndDeleteTokenFunc: func(ctx context.Context, userID uint) error {
			return errors.New("delete failed")
		},
	}
	uc := NewRefreshTokenUseCase(repo, 10*time.Second)
	ctx := context.Background()

	_, _, refreshToken, _, err := jwt.GenerateToken("test@example.com", 1)
	require.NoError(t, err)

	req := &request.ReqRefreshToken{RefreshToken: refreshToken}
	res, err := uc.RefreshToken(ctx, req)
	require.NoError(t, err)
	assert.True(t, createCalled)
	assert.NotEmpty(t, res.AccessToken)
	assert.NotEmpty(t, res.RefreshToken)
	t.Logf("RefreshToken succeeds even when FindOneByUserIDAndDeleteToken fails: new tokens returned")
}
