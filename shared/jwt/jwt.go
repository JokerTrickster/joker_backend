package jwt

import (
	"context"
	"fmt"
	"os"
	"runtime"
	"time"

	"github.com/golang-jwt/jwt"
	echojwt "github.com/labstack/echo-jwt"
)

type JwtCustomClaims struct {
	CreateTime int64  `json:"createTime"`
	UserID     uint   `json:"userID"`
	Email      string `json:"email"`
	jwt.StandardClaims
}

var AccessTokenSecretKey []byte
var RefreshTokenSecretKey []byte
var JwtConfig echojwt.Config

const (
	AccessTokenExpiredTime  = 24         //hours
	RefreshTokenExpiredTime = 1 * 24 * 7 //hours
)

// Error constants
const (
	ErrBadToken     = "BAD_TOKEN"
	ErrFromInternal = "INTERNAL_ERROR"
	ErrFromClient   = "CLIENT_ERROR"
)

// Helper functions
func TimeToEpochMillis(t time.Time) int64 {
	return t.UnixNano() / int64(time.Millisecond)
}

func Trace() string {
	pc := make([]uintptr, 10)
	runtime.Callers(2, pc)
	f := runtime.FuncForPC(pc[0])
	file, line := f.FileLine(pc[0])
	return fmt.Sprintf("%s:%d", file, line)
}

func ErrorMsg(ctx context.Context, code string, trace string, msg string, errType string) error {
	return fmt.Errorf("[%s] %s: %s (at %s)", errType, code, msg, trace)
}

func InitJwt() error {
	accessSecret := os.Getenv("JWT_ACCESS_SECRET")
	refreshSecret := os.Getenv("JWT_REFRESH_SECRET")

	if accessSecret == "" {
		accessSecret = os.Getenv("JWT_SECRET")
	}
	if refreshSecret == "" {
		refreshSecret = os.Getenv("JWT_SECRET")
	}

	isLocal := os.Getenv("IS_LOCAL")
	if accessSecret == "" {
		if isLocal == "true" {
			accessSecret = "local-dev-secret-do-not-use-in-production"
		} else {
			return fmt.Errorf("JWT_SECRET or JWT_ACCESS_SECRET must be set in production")
		}
	}
	if refreshSecret == "" {
		if isLocal == "true" {
			refreshSecret = "local-dev-refresh-secret-do-not-use-in-production"
		} else {
			return fmt.Errorf("JWT_SECRET or JWT_REFRESH_SECRET must be set in production")
		}
	}

	AccessTokenSecretKey = []byte(accessSecret)
	RefreshTokenSecretKey = []byte(refreshSecret)
	return nil
}

func GenerateToken(email string, userID uint) (string, int64, string, int64, error) {
	now := time.Now()
	accessToken, accessTknExpiredAt, err := GenerateAccessToken(email, now, userID)
	if err != nil {
		return "", 0, "", 0, err
	}
	refreshToken, refreshTknExpiredAt, err := GenerateRefreshToken(email, now, userID)
	if err != nil {
		return "", 0, "", 0, err
	}
	return accessToken, accessTknExpiredAt, refreshToken, refreshTknExpiredAt, nil
}

func GenerateAccessToken(email string, now time.Time, userID uint) (string, int64, error) {
	// Set custom claims
	expiredAt := now.Add(time.Hour * AccessTokenExpiredTime).Unix()
	claims := &JwtCustomClaims{
		TimeToEpochMillis(now),
		userID,
		email,
		jwt.StandardClaims{
			ExpiresAt: expiredAt,
		},
	}

	// Create token with claims
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	// Generate encoded token and send it as response.
	accessToken, err := token.SignedString(AccessTokenSecretKey)
	if err != nil {
		return "", 0, err
	}
	return accessToken, expiredAt, nil
}

func GenerateRefreshToken(email string, now time.Time, userID uint) (string, int64, error) {
	expiredAt := now.Add(time.Hour * RefreshTokenExpiredTime).Unix()
	claims := &JwtCustomClaims{
		TimeToEpochMillis(now),
		userID,
		email,
		jwt.StandardClaims{
			ExpiresAt: expiredAt,
		},
	}

	// Create token with claims
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	// Generate encoded token and send it as response.
	refreshToken, err := token.SignedString(RefreshTokenSecretKey)
	if err != nil {
		return "", 0, ErrorMsg(context.TODO(), ErrBadToken, Trace(), fmt.Sprintf("failed to generate refresh token - %v", err), ErrFromInternal)
	}
	return refreshToken, expiredAt, nil
}
func VerifyToken(tokenString string) error {
	// Parse the token
	token, err := jwt.ParseWithClaims(tokenString, &JwtCustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		return AccessTokenSecretKey, nil
	})
	if err != nil {
		return ErrorMsg(context.TODO(), ErrBadToken, Trace(), fmt.Sprintf("failed to parse token - %v err - %v", token, err.Error()), ErrFromClient)
	}

	// Check token validity
	if !token.Valid {
		return ErrorMsg(context.TODO(), ErrBadToken, Trace(), fmt.Sprintf("invalid token - %v ", token), ErrFromClient)
	}
	return nil
}
func ParseToken(tokenString string) (uint, string, error) {
	token, err := jwt.ParseWithClaims(tokenString, &JwtCustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		return AccessTokenSecretKey, nil
	})
	if err != nil {
		return 0, "", ErrorMsg(context.TODO(), ErrBadToken, Trace(), fmt.Sprintf("failed to parse token - %v", err), ErrFromClient)
	}

	claims, ok := token.Claims.(*JwtCustomClaims)
	if !ok {
		return 0, "", ErrorMsg(context.TODO(), ErrBadToken, Trace(), fmt.Sprintf("failed to extract claims - %v", token), ErrFromClient)
	}

	return claims.UserID, claims.Email, nil
}

func VerifyRefreshToken(tokenString string) (uint, string, error) {
	// Parse the refresh token
	token, err := jwt.ParseWithClaims(tokenString, &JwtCustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		return RefreshTokenSecretKey, nil
	})
	if err != nil {
		return 0, "", ErrorMsg(context.TODO(), ErrBadToken, Trace(), fmt.Sprintf("failed to parse refresh token - %v", err.Error()), ErrFromClient)
	}

	// Check token validity
	if !token.Valid {
		return 0, "", ErrorMsg(context.TODO(), ErrBadToken, Trace(), "refresh token is invalid or expired", ErrFromClient)
	}

	// Extract claims
	claims, ok := token.Claims.(*JwtCustomClaims)
	if !ok {
		return 0, "", ErrorMsg(context.TODO(), ErrBadToken, Trace(), fmt.Sprintf("failed to extract claims - %v", token), ErrFromClient)
	}

	// Check expiration time
	if time.Now().Unix() > claims.ExpiresAt {
		return 0, "", ErrorMsg(context.TODO(), ErrBadToken, Trace(), "refresh token is expired", ErrFromClient)
	}

	return claims.UserID, claims.Email, nil
}
