package usecase

import (
	"context"
	"testing"
	"time"

	"github.com/JokerTrickster/joker_backend/services/authService/features/auth/model/request"
	"github.com/JokerTrickster/joker_backend/services/authService/features/auth/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSigninAuthUseCase_Success(t *testing.T) {
	db := setupTestDB(t)
	initJWTForTest(t)

	// First, sign up a user (password will be bcrypt hashed)
	signupRepo := repository.NewSignupAuthRepository(db)
	signupUC := NewSignupAuthUseCase(signupRepo, 10*time.Second)

	ctx := context.Background()
	email := "signin-test-" + time.Now().Format("20060102150405") + "@example.com"
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
	email := "signin-wrongpw-" + time.Now().Format("20060102150405") + "@example.com"

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
		Email:       "nonexistent-" + time.Now().Format("20060102150405") + "@example.com",
		Password:    "anypassword",
		ServiceType: "game",
	}

	_, err := signinUC.Signin(ctx, signinReq)
	assert.Error(t, err, "Signin with nonexistent user should fail")
	assert.Contains(t, err.Error(), "user not found", "Error should mention user not found")

	t.Logf("Nonexistent user correctly rejected: %v", err)
}
