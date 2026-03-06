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
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type mockCheckEmailRepository struct {
	CheckEmailExistsFunc func(ctx context.Context, email string, provider string) (bool, error)
}

func (m *mockCheckEmailRepository) CheckEmailExists(ctx context.Context, email string, provider string) (bool, error) {
	if m.CheckEmailExistsFunc != nil {
		return m.CheckEmailExistsFunc(ctx, email, provider)
	}
	return false, nil
}

func setupTestDBForCheckEmail(t *testing.T) *gorm.DB {
	// authService용 데이터베이스 연결
	dsn := "root:rootpassword@tcp(localhost:3307)/test_db?charset=utf8mb4&parseTime=True&loc=Local"
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Skipf("Integration test: requires test database: %v", err)
	}
	return db
}

func TestCheckEmailAuthUseCase_EmailExists(t *testing.T) {
	db := setupTestDBForCheckEmail(t)

	// 먼저 테스트용 사용자 생성
	signupRepo := repository.NewSignupAuthRepository(db)
	signupUC := NewSignupAuthUseCase(signupRepo, 10*time.Second)

	ctx := context.Background()
	testEmail := "test-check-" + fmt.Sprintf("%d_%d", time.Now().UnixNano(), rand.Intn(100000)) + "@example.com"

	signupReq := &request.ReqSignUp{
		Email:       testEmail,
		Password:    "password123",
		ServiceType: "game",
		Name:        "Test User",
	}

	// 회원가입
	_, err := signupUC.Signup(ctx, signupReq)
	assert.NoError(t, err, "Signup should succeed")

	// 이메일 중복 체크
	checkRepo := repository.NewCheckEmailAuthRepository(db)
	checkUC := NewCheckEmailAuthUseCase(checkRepo, 10*time.Second)

	checkReq := &request.ReqCheckEmail{
		Email:    testEmail,
		Provider: "game", // ServiceType이 provider로 저장됨
	}

	res, err := checkUC.CheckEmail(ctx, checkReq)
	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.Equal(t, testEmail, res.Email)
	assert.True(t, res.Exists, "Email should exist")
	assert.False(t, res.Available, "Email should not be available")

	t.Logf("Email check result: exists=%v, available=%v", res.Exists, res.Available)
}

func TestCheckEmailAuthUseCase_EmailNotExists(t *testing.T) {
	db := setupTestDBForCheckEmail(t)

	checkRepo := repository.NewCheckEmailAuthRepository(db)
	checkUC := NewCheckEmailAuthUseCase(checkRepo, 10*time.Second)

	ctx := context.Background()
	checkReq := &request.ReqCheckEmail{
		Email:    "nonexistent-" + fmt.Sprintf("%d_%d", time.Now().UnixNano(), rand.Intn(100000)) + "@example.com",
		Provider: "email",
	}

	res, err := checkUC.CheckEmail(ctx, checkReq)
	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.False(t, res.Exists, "Email should not exist")
	assert.True(t, res.Available, "Email should be available")

	t.Logf("Email check result: exists=%v, available=%v", res.Exists, res.Available)
}

func TestCheckEmailAuthUseCase_DifferentProviders(t *testing.T) {
	db := setupTestDBForCheckEmail(t)

	// email provider로 사용자 생성
	signupRepo := repository.NewSignupAuthRepository(db)
	signupUC := NewSignupAuthUseCase(signupRepo, 10*time.Second)

	ctx := context.Background()
	email := "test-provider-" + fmt.Sprintf("%d_%d", time.Now().UnixNano(), rand.Intn(100000)) + "@example.com"

	signupReq := &request.ReqSignUp{
		Email:       email,
		Password:    "password123",
		ServiceType: "game",
		Name:        "Test Provider User",
	}

	_, err := signupUC.Signup(ctx, signupReq)
	if err != nil {
		t.Logf("Signup may have failed (expected if duplicate): %v", err)
	}

	checkRepo := repository.NewCheckEmailAuthRepository(db)
	checkUC := NewCheckEmailAuthUseCase(checkRepo, 10*time.Second)

	// game provider로 체크 - 존재해야 함
	checkReq1 := &request.ReqCheckEmail{
		Email:    email,
		Provider: "game",
	}

	res1, err := checkUC.CheckEmail(ctx, checkReq1)
	assert.NoError(t, err)
	assert.True(t, res1.Exists, "Same email with game provider should exist")

	// google provider로 체크 - 존재하지 않아야 함
	checkReq2 := &request.ReqCheckEmail{
		Email:    email,
		Provider: "google",
	}

	res2, err := checkUC.CheckEmail(ctx, checkReq2)
	assert.NoError(t, err)
	assert.NotNil(t, res2)
	assert.False(t, res2.Exists, "Same email with google provider should not exist")
	assert.True(t, res2.Available, "Email with google provider should be available")

	t.Logf("Game provider: exists=%v, Google provider: exists=%v", res1.Exists, res2.Exists)
}

func TestNewCheckEmailAuthUseCase(t *testing.T) {
	repo := &mockCheckEmailRepository{}
	uc := NewCheckEmailAuthUseCase(repo, 5*time.Second).(*CheckEmailAuthUseCase)
	require.NotNil(t, uc)
	assert.Equal(t, repo, uc.Repository)
	assert.Equal(t, 5*time.Second, uc.ContextTimeout)
	t.Logf("NewCheckEmailAuthUseCase sets Repository and ContextTimeout correctly")
}

func TestCheckEmailAuthUseCase_ImplementsInterface(t *testing.T) {
	repo := &mockCheckEmailRepository{}
	uc := NewCheckEmailAuthUseCase(repo, 10*time.Second)
	var _ _interface.ICheckEmailAuthUseCase = uc
	t.Logf("CheckEmailAuthUseCase implements ICheckEmailAuthUseCase")
}

func TestCheckEmailAuthUseCase_CheckEmail_RepoReturnsExists(t *testing.T) {
	repo := &mockCheckEmailRepository{
		CheckEmailExistsFunc: func(ctx context.Context, email string, provider string) (bool, error) {
			return true, nil
		},
	}
	uc := NewCheckEmailAuthUseCase(repo, 10*time.Second)
	ctx := context.Background()
	req := &request.ReqCheckEmail{Email: "exists@example.com", Provider: "game"}

	res, err := uc.CheckEmail(ctx, req)
	require.NoError(t, err)
	require.NotNil(t, res)
	assert.Equal(t, "exists@example.com", res.Email)
	assert.True(t, res.Exists)
	assert.False(t, res.Available)
	t.Logf("CheckEmail with exists=true: res=%+v", res)
}

func TestCheckEmailAuthUseCase_CheckEmail_RepoReturnsNotExists(t *testing.T) {
	repo := &mockCheckEmailRepository{
		CheckEmailExistsFunc: func(ctx context.Context, email string, provider string) (bool, error) {
			return false, nil
		},
	}
	uc := NewCheckEmailAuthUseCase(repo, 10*time.Second)
	ctx := context.Background()
	req := &request.ReqCheckEmail{Email: "new@example.com", Provider: "email"}

	res, err := uc.CheckEmail(ctx, req)
	require.NoError(t, err)
	require.NotNil(t, res)
	assert.Equal(t, "new@example.com", res.Email)
	assert.False(t, res.Exists)
	assert.True(t, res.Available)
	t.Logf("CheckEmail with exists=false: res=%+v", res)
}

func TestCheckEmailAuthUseCase_CheckEmail_RepoError(t *testing.T) {
	repoErr := errors.New("database connection failed")
	repo := &mockCheckEmailRepository{
		CheckEmailExistsFunc: func(ctx context.Context, email string, provider string) (bool, error) {
			return false, repoErr
		},
	}
	uc := NewCheckEmailAuthUseCase(repo, 10*time.Second)
	ctx := context.Background()
	req := &request.ReqCheckEmail{Email: "test@example.com", Provider: "game"}

	res, err := uc.CheckEmail(ctx, req)
	require.Error(t, err)
	assert.Nil(t, res)
	assert.Contains(t, err.Error(), "failed to check email")
	assert.ErrorIs(t, err, repoErr)
	t.Logf("CheckEmail repo error correctly propagated: %v", err)
}

func TestCheckEmailAuthUseCase_CheckEmail_RepoCalledWithCorrectArgs(t *testing.T) {
	var capturedEmail, capturedProvider string
	repo := &mockCheckEmailRepository{
		CheckEmailExistsFunc: func(ctx context.Context, email string, provider string) (bool, error) {
			capturedEmail = email
			capturedProvider = provider
			return false, nil
		},
	}
	uc := NewCheckEmailAuthUseCase(repo, 10*time.Second)
	ctx := context.Background()
	req := &request.ReqCheckEmail{Email: "verify-args@example.com", Provider: "google"}

	res, err := uc.CheckEmail(ctx, req)
	require.NoError(t, err)
	require.NotNil(t, res)
	assert.Equal(t, "verify-args@example.com", capturedEmail, "Repo should receive email from request")
	assert.Equal(t, "google", capturedProvider, "Repo should receive provider from request")
	assert.Equal(t, "verify-args@example.com", res.Email)
	assert.False(t, res.Exists)
	assert.True(t, res.Available)
	t.Logf("CheckEmail passes correct args to repo: email=%s, provider=%s", capturedEmail, capturedProvider)
}
