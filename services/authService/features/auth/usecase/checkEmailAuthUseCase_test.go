package usecase

import (
	"context"
	"testing"
	"time"

	"github.com/JokerTrickster/joker_backend/services/authService/features/auth/model/request"
	"github.com/JokerTrickster/joker_backend/services/authService/features/auth/repository"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func setupTestDBForCheckEmail(t *testing.T) *gorm.DB {
	// authService용 데이터베이스 연결
	dsn := "root:rootpassword@tcp(localhost:3307)/test_db?charset=utf8mb4&parseTime=True&loc=Local"
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}
	return db
}

func TestCheckEmailAuthUseCase_EmailExists(t *testing.T) {
	db := setupTestDBForCheckEmail(t)

	// 먼저 테스트용 사용자 생성
	signupRepo := repository.NewSignupAuthRepository(db)
	signupUC := NewSignupAuthUseCase(signupRepo, 10*time.Second)

	ctx := context.Background()
	testEmail := "test-check-" + time.Now().Format("20060102150405") + "@example.com"

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
		Email:    "nonexistent-" + time.Now().Format("20060102150405") + "@example.com",
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
	email := "test-provider-" + time.Now().Format("20060102150405") + "@example.com"

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
