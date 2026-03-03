package repository

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	mysqlDriver "gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func TestRepositoryStructInstantiation(t *testing.T) {
	dsn := os.Getenv("TEST_DB_DSN")
	if dsn == "" {
		dsn = "root:rootpassword@tcp(localhost:3307)/test_db?charset=utf8mb4&parseTime=True&loc=Local"
	}
	db, err := gorm.Open(mysqlDriver.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Skipf("Skipping repository struct test: DB unavailable: %v", err)
		return
	}

	signin := NewSigninAuthRepository(db)
	require.NotNil(t, signin, "SigninAuthRepository should be instantiated")

	signup := NewSignupAuthRepository(db)
	require.NotNil(t, signup, "SignupAuthRepository should be instantiated")

	logout := NewLogoutAuthRepository(db)
	require.NotNil(t, logout, "LogoutAuthRepository should be instantiated")

	refresh := NewRefreshTokenAuthRepository(db)
	require.NotNil(t, refresh, "RefreshTokenAuthRepository should be instantiated")

	checkEmail := NewCheckEmailAuthRepository(db)
	require.NotNil(t, checkEmail, "CheckEmailAuthRepository should be instantiated")

	googleSignin := NewGoogleSigninAuthRepository(db)
	require.NotNil(t, googleSignin, "GoogleSigninAuthRepository should be instantiated")

	t.Logf("All repository structs instantiate correctly")
}
