package usecase

import (
	"context"
	"os"
	"testing"
	"time"

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
	email := "signup-test-" + time.Now().Format("20060102150405") + "@example.com"

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
	email := "dup-test-" + time.Now().Format("20060102150405") + "@example.com"

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
