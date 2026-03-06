package usecase

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"os"
	"testing"
	"time"

	_interface "github.com/JokerTrickster/joker_backend/services/authService/features/auth/model/interface"
	"github.com/JokerTrickster/joker_backend/services/authService/features/auth/model/request"
	"github.com/JokerTrickster/joker_backend/services/authService/features/auth/repository"
	"github.com/JokerTrickster/joker_backend/shared/db/mysql"
	"github.com/JokerTrickster/joker_backend/shared/jwt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
	mysqlDriver "gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type mockSignupRepository struct {
	CreateUserFunc func(ctx context.Context, name string, email string, password string, provider string) (uint, error)
}

func (m *mockSignupRepository) CreateUser(ctx context.Context, name string, email string, password string, provider string) (uint, error) {
	if m.CreateUserFunc != nil {
		return m.CreateUserFunc(ctx, name, email, password, provider)
	}
	return 0, nil
}

func setupTestDB(t *testing.T) *gorm.DB {
	dsn := os.Getenv("TEST_DB_DSN")
	if dsn == "" {
		dsn = "root:rootpassword@tcp(localhost:3307)/test_db?charset=utf8mb4&parseTime=True&loc=Local"
	}
	db, err := gorm.Open(mysqlDriver.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Skipf("Integration test: requires test database: %v", err)
	}
	return db
}

func requireTokensTable(t *testing.T, db *gorm.DB) {
	t.Helper()
	if !db.Migrator().HasTable(&mysql.Tokens{}) {
		t.Skipf("Integration test: requires 'tokens' table in test database")
	}
}

func initJWTForTest(t *testing.T) {
	t.Helper()
	os.Setenv("IS_LOCAL", "true")
	err := jwt.InitJwt()
	require.NoError(t, err, "JWT init should succeed")
}

func TestSignupAuthUseCase_Success(t *testing.T) {
	db := setupTestDB(t)
	initJWTForTest(t)

	repo := repository.NewSignupAuthRepository(db)
	uc := NewSignupAuthUseCase(repo, 10*time.Second)

	ctx := context.Background()
	email := "signup-test-" + fmt.Sprintf("%d_%d", time.Now().UnixNano(), rand.Intn(100000)) + "@example.com"

	req := &request.ReqSignUp{
		Email:       email,
		Password:    "securepassword123",
		ServiceType: "game",
		Name:        "Signup Test User",
	}

	res, err := uc.Signup(ctx, req)
	require.NoError(t, err, "Signup should succeed")
	assert.NotEmpty(t, res.AccessToken, "AccessToken should be set")
	assert.NotEmpty(t, res.RefreshToken, "RefreshToken should be set")

	t.Logf("Signup succeeded: email=%s, accessToken=%s..., refreshToken=%s...",
		email, res.AccessToken[:20], res.RefreshToken[:20])

	// Verify password is stored as bcrypt hash, not plaintext
	var user mysql.Users
	err = db.Where("email = ? AND provider = ?", email, "game").First(&user).Error
	require.NoError(t, err, "User should be found in DB")

	assert.NotEqual(t, "securepassword123", user.Password, "Password should NOT be stored in plaintext")
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte("securepassword123"))
	assert.NoError(t, err, "Stored password should be a valid bcrypt hash of the original password")

	t.Logf("Verified bcrypt hash: stored=%s...", user.Password[:20])
}

func TestSignupAuthUseCase_DuplicateEmail(t *testing.T) {
	db := setupTestDB(t)
	initJWTForTest(t)

	repo := repository.NewSignupAuthRepository(db)
	uc := NewSignupAuthUseCase(repo, 10*time.Second)

	ctx := context.Background()
	email := "dup-test-" + fmt.Sprintf("%d_%d", time.Now().UnixNano(), rand.Intn(100000)) + "@example.com"

	req := &request.ReqSignUp{
		Email:       email,
		Password:    "password123",
		ServiceType: "game",
		Name:        "First User",
	}

	_, err := uc.Signup(ctx, req)
	require.NoError(t, err, "First signup should succeed")

	// Second signup with same email+provider should fail
	req2 := &request.ReqSignUp{
		Email:       email,
		Password:    "differentpassword",
		ServiceType: "game",
		Name:        "Second User",
	}

	_, err = uc.Signup(ctx, req2)
	assert.Error(t, err, "Duplicate signup should fail")
	assert.Contains(t, err.Error(), "email already exists", "Error should mention email duplication")

	t.Logf("Duplicate email correctly rejected: %v", err)
}

func TestNewSignupAuthUseCase(t *testing.T) {
	repo := &mockSignupRepository{}
	uc := NewSignupAuthUseCase(repo, 5*time.Second).(*SignupAuthUseCase)
	require.NotNil(t, uc)
	assert.Equal(t, repo, uc.Repository)
	assert.Equal(t, 5*time.Second, uc.ContextTimeout)
	t.Logf("NewSignupAuthUseCase sets Repository and ContextTimeout correctly")
}

func TestSignupAuthUseCase_ImplementsInterface(t *testing.T) {
	repo := &mockSignupRepository{}
	uc := NewSignupAuthUseCase(repo, 10*time.Second)
	var _ _interface.ISignupAuthUseCase = uc
	t.Logf("SignupAuthUseCase implements ISignupAuthUseCase")
}

func TestSignupAuthUseCase_Signup_RepoError(t *testing.T) {
	repoErr := errors.New("email already exists")
	repo := &mockSignupRepository{
		CreateUserFunc: func(ctx context.Context, name string, email string, password string, provider string) (uint, error) {
			return 0, repoErr
		},
	}
	uc := NewSignupAuthUseCase(repo, 10*time.Second)
	ctx := context.Background()
	req := &request.ReqSignUp{
		Email:       "test@example.com",
		Password:    "pass123",
		ServiceType: "game",
		Name:        "Test User",
	}

	res, err := uc.Signup(ctx, req)
	require.Error(t, err)
	assert.ErrorIs(t, err, repoErr)
	assert.Empty(t, res.AccessToken)
	assert.Empty(t, res.RefreshToken)
	t.Logf("Signup with repo error correctly propagated: %v", err)
}

func TestSignupAuthUseCase_Signup_SuccessWithMock(t *testing.T) {
	initJWTForTest(t)

	var createUserCalled bool
	repo := &mockSignupRepository{
		CreateUserFunc: func(ctx context.Context, name string, email string, password string, provider string) (uint, error) {
			createUserCalled = true
			assert.Equal(t, "Test User", name, "Name should be passed to repo")
			assert.Equal(t, "signup-unit@example.com", email, "Email should be passed to repo")
			assert.Equal(t, "secretpass123", password, "Password should be passed to repo")
			assert.Equal(t, "game", provider, "ServiceType should be passed as provider")
			return 42, nil
		},
	}
	uc := NewSignupAuthUseCase(repo, 10*time.Second)
	ctx := context.Background()
	req := &request.ReqSignUp{
		Email:       "signup-unit@example.com",
		Password:    "secretpass123",
		ServiceType: "game",
		Name:        "Test User",
	}

	res, err := uc.Signup(ctx, req)
	require.NoError(t, err, "Signup with mock should succeed")
	require.True(t, createUserCalled, "CreateUser should have been called")
	assert.NotEmpty(t, res.AccessToken, "AccessToken should be generated")
	assert.NotEmpty(t, res.RefreshToken, "RefreshToken should be generated")

	t.Logf("Signup success with mock: accessToken=%s..., refreshToken=%s...",
		res.AccessToken[:min(20, len(res.AccessToken))], res.RefreshToken[:min(20, len(res.RefreshToken))])
}
