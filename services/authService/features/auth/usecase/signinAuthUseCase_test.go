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
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockSigninRepository struct {
	FindUserByEmailFunc func(c context.Context, email string, password string, serviceType string) (uint, string, error)
}

func (m *mockSigninRepository) FindUserByEmail(c context.Context, email string, password string, serviceType string) (uint, string, error) {
	if m.FindUserByEmailFunc != nil {
		return m.FindUserByEmailFunc(c, email, password, serviceType)
	}
	return 0, "", nil
}

func TestSigninAuthUseCase_Success(t *testing.T) {
	db := setupTestDB(t)
	initJWTForTest(t)

	// First, sign up a user (password will be bcrypt hashed)
	signupRepo := repository.NewSignupAuthRepository(db)
	signupUC := NewSignupAuthUseCase(signupRepo, 10*time.Second)

	ctx := context.Background()
	email := fmt.Sprintf("signin-test-%d-%d@example.com", time.Now().UnixNano(), rand.Intn(100000))
	password := "correctpassword123"

	signupReq := &request.ReqSignUp{
		Email:       email,
		Password:    password,
		ServiceType: "game",
		Name:        "Signin Test User",
	}

	_, err := signupUC.Signup(ctx, signupReq)
	require.NoError(t, err, "Signup should succeed for signin test setup")

	// Now sign in with correct password
	signinRepo := repository.NewSigninAuthRepository(db)
	signinUC := NewSigninAuthUseCase(signinRepo, 10*time.Second)

	signinReq := &request.ReqSignIn{
		Email:       email,
		Password:    password,
		ServiceType: "game",
	}

	res, err := signinUC.Signin(ctx, signinReq)
	require.NoError(t, err, "Signin with correct password should succeed")
	assert.NotEmpty(t, res.AccessToken, "AccessToken should be set")
	assert.NotEmpty(t, res.RefreshToken, "RefreshToken should be set")

	t.Logf("Signin succeeded: email=%s, accessToken=%s...", email, res.AccessToken[:20])
}

func TestSigninAuthUseCase_WrongPassword(t *testing.T) {
	db := setupTestDB(t)
	initJWTForTest(t)

	signupRepo := repository.NewSignupAuthRepository(db)
	signupUC := NewSignupAuthUseCase(signupRepo, 10*time.Second)

	ctx := context.Background()
	email := fmt.Sprintf("signin-wrongpw-%d-%d@example.com", time.Now().UnixNano(), rand.Intn(100000))

	signupReq := &request.ReqSignUp{
		Email:       email,
		Password:    "correctpassword",
		ServiceType: "game",
		Name:        "Wrong PW User",
	}

	_, err := signupUC.Signup(ctx, signupReq)
	require.NoError(t, err, "Signup should succeed")

	signinRepo := repository.NewSigninAuthRepository(db)
	signinUC := NewSigninAuthUseCase(signinRepo, 10*time.Second)

	signinReq := &request.ReqSignIn{
		Email:       email,
		Password:    "wrongpassword",
		ServiceType: "game",
	}

	_, err = signinUC.Signin(ctx, signinReq)
	assert.Error(t, err, "Signin with wrong password should fail")
	assert.Contains(t, err.Error(), "password not match", "Error should mention password mismatch")

	t.Logf("Wrong password correctly rejected: %v", err)
}

func TestSigninAuthUseCase_UserNotFound(t *testing.T) {
	db := setupTestDB(t)
	initJWTForTest(t)

	signinRepo := repository.NewSigninAuthRepository(db)
	signinUC := NewSigninAuthUseCase(signinRepo, 10*time.Second)

	ctx := context.Background()
	signinReq := &request.ReqSignIn{
		Email:       fmt.Sprintf("nonexistent-%d-%d@example.com", time.Now().UnixNano(), rand.Intn(100000)),
		Password:    "anypassword",
		ServiceType: "game",
	}

	_, err := signinUC.Signin(ctx, signinReq)
	assert.Error(t, err, "Signin with nonexistent user should fail")
	assert.Contains(t, err.Error(), "user not found", "Error should mention user not found")

	t.Logf("Nonexistent user correctly rejected: %v", err)
}

func TestNewSigninAuthUseCase(t *testing.T) {
	repo := &mockSigninRepository{}
	uc := NewSigninAuthUseCase(repo, 5*time.Second).(*SigninAuthUseCase)
	require.NotNil(t, uc)
	assert.Equal(t, repo, uc.Repository)
	assert.Equal(t, 5*time.Second, uc.ContextTimeout)
	t.Logf("NewSigninAuthUseCase sets Repository and ContextTimeout correctly")
}

func TestSigninAuthUseCase_ImplementsInterface(t *testing.T) {
	repo := &mockSigninRepository{}
	uc := NewSigninAuthUseCase(repo, 10*time.Second)
	var _ _interface.ISigninAuthUseCase = uc
	t.Logf("SigninAuthUseCase implements ISigninAuthUseCase")
}

func TestSigninAuthUseCase_Signin_PasswordMismatchFromRepo(t *testing.T) {
	repoErr := errors.New("password not match")
	repo := &mockSigninRepository{
		FindUserByEmailFunc: func(c context.Context, email string, password string, serviceType string) (uint, string, error) {
			return 0, "", repoErr
		},
	}
	uc := NewSigninAuthUseCase(repo, 10*time.Second)
	ctx := context.Background()
	req := &request.ReqSignIn{Email: "test@example.com", Password: "wrongpassword", ServiceType: "game"}

	res, err := uc.Signin(ctx, req)
	require.Error(t, err)
	assert.ErrorIs(t, err, repoErr)
	assert.Empty(t, res.AccessToken)
	assert.Empty(t, res.RefreshToken)
	t.Logf("Password mismatch correctly propagated: %v", err)
}

func TestSigninAuthUseCase_Signin_UserNotFoundFromRepo(t *testing.T) {
	repoErr := errors.New("user not found")
	repo := &mockSigninRepository{
		FindUserByEmailFunc: func(c context.Context, email string, password string, serviceType string) (uint, string, error) {
			return 0, "", repoErr
		},
	}
	uc := NewSigninAuthUseCase(repo, 10*time.Second)
	ctx := context.Background()
	req := &request.ReqSignIn{Email: "nonexistent@example.com", Password: "anypass", ServiceType: "game"}

	res, err := uc.Signin(ctx, req)
	require.Error(t, err)
	assert.ErrorIs(t, err, repoErr)
	assert.Empty(t, res.AccessToken)
	t.Logf("User not found correctly propagated: %v", err)
}

func TestSigninAuthUseCase_Signin_SuccessWithMock(t *testing.T) {
	initJWTForTest(t)

	var findUserCalled bool
	repo := &mockSigninRepository{
		FindUserByEmailFunc: func(c context.Context, email string, password string, serviceType string) (uint, string, error) {
			findUserCalled = true
			assert.Equal(t, "signin-unit@example.com", email)
			assert.Equal(t, "mypassword", password)
			assert.Equal(t, "game", serviceType)
			return 100, "signin-unit@example.com", nil
		},
	}
	uc := NewSigninAuthUseCase(repo, 10*time.Second)
	ctx := context.Background()
	req := &request.ReqSignIn{
		Email:       "signin-unit@example.com",
		Password:    "mypassword",
		ServiceType: "game",
	}

	res, err := uc.Signin(ctx, req)
	require.NoError(t, err)
	require.True(t, findUserCalled)
	assert.NotEmpty(t, res.AccessToken)
	assert.NotEmpty(t, res.RefreshToken)
	t.Logf("Signin success with mock: accessToken=%s...", res.AccessToken[:min(20, len(res.AccessToken))])
}

func TestSigninAuthUseCase_Signin_RepoError(t *testing.T) {
	repoErr := errors.New("user not found")
	repo := &mockSigninRepository{
		FindUserByEmailFunc: func(c context.Context, email string, password string, serviceType string) (uint, string, error) {
			return 0, "", repoErr
		},
	}
	uc := NewSigninAuthUseCase(repo, 10*time.Second)
	ctx := context.Background()
	req := &request.ReqSignIn{Email: "test@example.com", Password: "pass", ServiceType: "game"}

	res, err := uc.Signin(ctx, req)
	require.Error(t, err)
	assert.ErrorIs(t, err, repoErr)
	assert.Empty(t, res.AccessToken)
	assert.Empty(t, res.RefreshToken)
	t.Logf("Signin with repo error correctly propagated: %v", err)
}
