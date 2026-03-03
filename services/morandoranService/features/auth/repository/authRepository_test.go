package repository

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/JokerTrickster/joker_backend/services/morandoranService/features/auth/model/entity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	dsn := os.Getenv("TEST_DB_DSN")
	if dsn == "" {
		dsn = "root:rootpassword@tcp(localhost:3307)/test_db?charset=utf8mb4&parseTime=True&loc=Local"
	}
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Skipf("Skipping: test database unavailable: %v", err)
	}
	return db
}

func TestAuthRepository_FindUserByEmail(t *testing.T) {
	db := setupTestDB(t)
	repo := NewAuthRepository(db)

	// Ensure morandoran_users table exists (may need migration)
	if err := db.AutoMigrate(&entity.MorandoranUser{}); err != nil {
		t.Skipf("Skipping: cannot ensure morandoran_users table: %v", err)
	}

	ctx := context.Background()
	email := "auth-repo-test-" + time.Now().Format("20060102150405") + "@example.com"
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	require.NoError(t, err)

	user := &entity.MorandoranUser{
		Nickname: "auth-repo-test",
		Email:    email,
		Password: string(hashedPassword),
		Role:     "user",
	}
	err = db.WithContext(ctx).Create(user).Error
	if err != nil {
		t.Skipf("Skipping: cannot create test user (table may not exist): %v", err)
	}
	defer func() {
		db.WithContext(ctx).Unscoped().Delete(user)
	}()

	t.Run("success: finds user by email", func(t *testing.T) {
		found, err := repo.FindUserByEmail(ctx, email)
		require.NoError(t, err)
		require.NotNil(t, found)
		assert.Equal(t, email, found.Email)
		assert.Equal(t, "auth-repo-test", found.Nickname)
		t.Logf("Found user by email: id=%d email=%s", found.ID, found.Email)
	})

	t.Run("not found: returns error for nonexistent email", func(t *testing.T) {
		found, err := repo.FindUserByEmail(ctx, "nonexistent-"+email)
		require.Error(t, err)
		assert.Nil(t, found)
		assert.Contains(t, err.Error(), "user not found")
		t.Logf("Not found error: %v", err)
	})
}

func TestAuthRepository_FindUserByID(t *testing.T) {
	db := setupTestDB(t)
	repo := NewAuthRepository(db)

	if err := db.AutoMigrate(&entity.MorandoranUser{}); err != nil {
		t.Skipf("Skipping: cannot ensure morandoran_users table: %v", err)
	}

	ctx := context.Background()
	email := "auth-repo-id-test-" + time.Now().Format("20060102150405") + "@example.com"
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	require.NoError(t, err)

	user := &entity.MorandoranUser{
		Nickname: "auth-repo-id-test",
		Email:    email,
		Password: string(hashedPassword),
		Role:     "user",
	}
	err = db.WithContext(ctx).Create(user).Error
	if err != nil {
		t.Skipf("Skipping: cannot create test user: %v", err)
	}
	defer func() {
		db.WithContext(ctx).Unscoped().Delete(user)
	}()

	t.Run("success: finds user by ID", func(t *testing.T) {
		found, err := repo.FindUserByID(ctx, user.ID)
		require.NoError(t, err)
		require.NotNil(t, found)
		assert.Equal(t, user.ID, found.ID)
		assert.Equal(t, email, found.Email)
		t.Logf("Found user by ID: id=%d email=%s", found.ID, found.Email)
	})

	t.Run("not found: returns error for nonexistent ID", func(t *testing.T) {
		found, err := repo.FindUserByID(ctx, 999999)
		require.Error(t, err)
		assert.Nil(t, found)
		assert.Contains(t, err.Error(), "user not found")
		t.Logf("Not found error: %v", err)
	})
}
