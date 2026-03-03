package jwt

import (
	"os"
	"testing"
	"time"
)

func setEnvForTest(t *testing.T, key, value string) {
	t.Helper()
	old, existed := os.LookupEnv(key)
	os.Setenv(key, value)
	t.Cleanup(func() {
		if existed {
			os.Setenv(key, old)
		} else {
			os.Unsetenv(key)
		}
	})
}

func clearJWTEnvVars(t *testing.T) {
	t.Helper()
	for _, key := range []string{"JWT_ACCESS_SECRET", "JWT_REFRESH_SECRET", "JWT_SECRET", "IS_LOCAL"} {
		old, existed := os.LookupEnv(key)
		os.Unsetenv(key)
		t.Cleanup(func() {
			if existed {
				os.Setenv(key, old)
			}
		})
	}
}

func initJWTLocal(t *testing.T) {
	t.Helper()
	clearJWTEnvVars(t)
	setEnvForTest(t, "IS_LOCAL", "true")
	if err := InitJwt(); err != nil {
		t.Fatalf("InitJwt should succeed in local mode: %v", err)
	}
}

func TestInitJwt_LocalMode(t *testing.T) {
	clearJWTEnvVars(t)
	setEnvForTest(t, "IS_LOCAL", "true")

	if err := InitJwt(); err != nil {
		t.Fatalf("InitJwt failed: %v", err)
	}
	if len(AccessTokenSecretKey) == 0 {
		t.Error("AccessTokenSecretKey should be set")
	}
	if len(RefreshTokenSecretKey) == 0 {
		t.Error("RefreshTokenSecretKey should be set")
	}
	if string(AccessTokenSecretKey) == string(RefreshTokenSecretKey) {
		t.Error("Access and refresh secrets should differ in local mode")
	}
	t.Logf("Local mode: access=%s, refresh=%s", string(AccessTokenSecretKey), string(RefreshTokenSecretKey))
}

func TestInitJwt_ExplicitSecrets(t *testing.T) {
	clearJWTEnvVars(t)
	setEnvForTest(t, "JWT_ACCESS_SECRET", "my-access-secret")
	setEnvForTest(t, "JWT_REFRESH_SECRET", "my-refresh-secret")

	if err := InitJwt(); err != nil {
		t.Fatalf("InitJwt failed: %v", err)
	}
	if string(AccessTokenSecretKey) != "my-access-secret" {
		t.Errorf("Expected access secret 'my-access-secret', got '%s'", string(AccessTokenSecretKey))
	}
	if string(RefreshTokenSecretKey) != "my-refresh-secret" {
		t.Errorf("Expected refresh secret 'my-refresh-secret', got '%s'", string(RefreshTokenSecretKey))
	}
	t.Logf("Explicit secrets loaded correctly")
}

func TestInitJwt_SharedSecret(t *testing.T) {
	clearJWTEnvVars(t)
	setEnvForTest(t, "JWT_SECRET", "shared-secret-value")

	if err := InitJwt(); err != nil {
		t.Fatalf("InitJwt failed: %v", err)
	}
	if string(AccessTokenSecretKey) != "shared-secret-value" {
		t.Errorf("Expected 'shared-secret-value', got '%s'", string(AccessTokenSecretKey))
	}
	if string(RefreshTokenSecretKey) != "shared-secret-value" {
		t.Errorf("Expected 'shared-secret-value', got '%s'", string(RefreshTokenSecretKey))
	}
	t.Logf("Shared secret applied to both access and refresh")
}

func TestInitJwt_ProductionNoSecret(t *testing.T) {
	clearJWTEnvVars(t)
	setEnvForTest(t, "IS_LOCAL", "false")

	err := InitJwt()
	if err == nil {
		t.Fatal("Production mode without secret should fail")
	}
	t.Logf("Production fail-fast: %v", err)
}

func TestGenerateToken(t *testing.T) {
	initJWTLocal(t)

	email := "test@example.com"
	userID := uint(42)

	accessToken, accessExp, refreshToken, refreshExp, err := GenerateToken(email, userID)
	if err != nil {
		t.Fatalf("GenerateToken failed: %v", err)
	}
	if accessToken == "" {
		t.Error("Access token should not be empty")
	}
	if refreshToken == "" {
		t.Error("Refresh token should not be empty")
	}
	if accessExp <= time.Now().Unix() {
		t.Error("Access token expiry should be in the future")
	}
	if refreshExp <= time.Now().Unix() {
		t.Error("Refresh token expiry should be in the future")
	}
	if refreshExp <= accessExp {
		t.Error("Refresh token should expire after access token")
	}
	if accessToken == refreshToken {
		t.Error("Access and refresh tokens should differ")
	}
	t.Logf("Generated tokens: access=%s..., refresh=%s..., accessExp=%d, refreshExp=%d",
		accessToken[:20], refreshToken[:20], accessExp, refreshExp)
}

func TestVerifyToken_Valid(t *testing.T) {
	initJWTLocal(t)

	accessToken, _, _, _, err := GenerateToken("user@test.com", 1)
	if err != nil {
		t.Fatalf("GenerateToken failed: %v", err)
	}

	if err := VerifyToken(accessToken); err != nil {
		t.Fatalf("Valid access token should verify: %v", err)
	}
	t.Logf("Valid token verified")
}

func TestVerifyToken_Invalid(t *testing.T) {
	initJWTLocal(t)

	if err := VerifyToken("invalid.token.string"); err == nil {
		t.Fatal("Invalid token should fail verification")
	} else {
		t.Logf("Invalid token rejected: %v", err)
	}
}

func TestVerifyToken_Tampered(t *testing.T) {
	initJWTLocal(t)

	accessToken, _, _, _, err := GenerateToken("user@test.com", 1)
	if err != nil {
		t.Fatalf("GenerateToken failed: %v", err)
	}

	tampered := accessToken[:len(accessToken)-5] + "XXXXX"
	if err := VerifyToken(tampered); err == nil {
		t.Fatal("Tampered token should fail verification")
	} else {
		t.Logf("Tampered token rejected: %v", err)
	}
}

func TestParseToken_Valid(t *testing.T) {
	initJWTLocal(t)

	email := "parse@test.com"
	userID := uint(99)

	accessToken, _, _, _, err := GenerateToken(email, userID)
	if err != nil {
		t.Fatalf("GenerateToken failed: %v", err)
	}

	parsedUserID, parsedEmail, err := ParseToken(accessToken)
	if err != nil {
		t.Fatalf("ParseToken failed: %v", err)
	}
	if parsedUserID != userID {
		t.Errorf("Expected userID %d, got %d", userID, parsedUserID)
	}
	if parsedEmail != email {
		t.Errorf("Expected email %s, got %s", email, parsedEmail)
	}
	t.Logf("Parsed token: userID=%d, email=%s", parsedUserID, parsedEmail)
}

func TestParseToken_Invalid(t *testing.T) {
	initJWTLocal(t)

	_, _, err := ParseToken("bad.token.data")
	if err == nil {
		t.Fatal("Parsing invalid token should fail")
	}
	t.Logf("Invalid token parse rejected: %v", err)
}

func TestVerifyRefreshToken_Valid(t *testing.T) {
	initJWTLocal(t)

	email := "refresh@test.com"
	userID := uint(55)

	_, _, refreshToken, _, err := GenerateToken(email, userID)
	if err != nil {
		t.Fatalf("GenerateToken failed: %v", err)
	}

	parsedUserID, parsedEmail, err := VerifyRefreshToken(refreshToken)
	if err != nil {
		t.Fatalf("VerifyRefreshToken failed: %v", err)
	}
	if parsedUserID != userID {
		t.Errorf("Expected userID %d, got %d", userID, parsedUserID)
	}
	if parsedEmail != email {
		t.Errorf("Expected email %s, got %s", email, parsedEmail)
	}
	t.Logf("Refresh token verified: userID=%d, email=%s", parsedUserID, parsedEmail)
}

func TestVerifyRefreshToken_Invalid(t *testing.T) {
	initJWTLocal(t)

	_, _, err := VerifyRefreshToken("invalid.refresh.token")
	if err == nil {
		t.Fatal("Invalid refresh token should fail")
	}
	t.Logf("Invalid refresh token rejected: %v", err)
}

func TestVerifyRefreshToken_AccessTokenAsRefresh(t *testing.T) {
	initJWTLocal(t)

	accessToken, _, _, _, err := GenerateToken("cross@test.com", 10)
	if err != nil {
		t.Fatalf("GenerateToken failed: %v", err)
	}

	// Access token signed with AccessTokenSecretKey should fail
	// when verified with RefreshTokenSecretKey (they differ in local mode)
	_, _, err = VerifyRefreshToken(accessToken)
	if err == nil {
		t.Fatal("Access token should not pass refresh token verification")
	}
	t.Logf("Cross-token verification correctly rejected: %v", err)
}

func TestGenerateAccessToken_Expiry(t *testing.T) {
	initJWTLocal(t)

	now := time.Now()
	_, expiredAt, err := GenerateAccessToken("expiry@test.com", now, 1)
	if err != nil {
		t.Fatalf("GenerateAccessToken failed: %v", err)
	}

	expectedExpiry := now.Add(time.Hour * AccessTokenExpiredTime).Unix()
	if expiredAt != expectedExpiry {
		t.Errorf("Expected expiry %d, got %d", expectedExpiry, expiredAt)
	}
	t.Logf("Access token expires at: %v (expected: %v)", time.Unix(expiredAt, 0), time.Unix(expectedExpiry, 0))
}

func TestGenerateRefreshToken_Expiry(t *testing.T) {
	initJWTLocal(t)

	now := time.Now()
	_, expiredAt, err := GenerateRefreshToken("expiry@test.com", now, 1)
	if err != nil {
		t.Fatalf("GenerateRefreshToken failed: %v", err)
	}

	expectedExpiry := now.Add(time.Hour * RefreshTokenExpiredTime).Unix()
	if expiredAt != expectedExpiry {
		t.Errorf("Expected expiry %d, got %d", expectedExpiry, expiredAt)
	}
	t.Logf("Refresh token expires at: %v (expected: %v)", time.Unix(expiredAt, 0), time.Unix(expectedExpiry, 0))
}
