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
	"github.com/JokerTrickster/joker_backend/shared/jwt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/api/idtoken"
)

type mockGoogleSigninRepository struct {
	FindOrCreateUserByGoogleEmailFunc func(ctx context.Context, email string, name string) (uint, error)
}

func (m *mockGoogleSigninRepository) FindOrCreateUserByGoogleEmail(ctx context.Context, email string, name string) (uint, error) {
	if m.FindOrCreateUserByGoogleEmailFunc != nil {
		return m.FindOrCreateUserByGoogleEmailFunc(ctx, email, name)
	}
	return 0, nil
}

func TestNewGoogleSigninAuthUseCase_ReadsGoogleClientID(t *testing.T) {
	testClientID := "test-client-id-" + fmt.Sprintf("%d_%d", time.Now().UnixNano(), rand.Intn(100000))
	os.Setenv("GOOGLE_CLIENT_ID", testClientID)
	defer os.Unsetenv("GOOGLE_CLIENT_ID")

	repo := &mockGoogleSigninRepository{}
	uc := NewGoogleSigninAuthUseCase(repo, 10*time.Second).(*GoogleSigninAuthUseCase)

	require.NotNil(t, uc, "UseCase should not be nil")
	assert.Equal(t, testClientID, uc.GoogleClientID, "GoogleClientID should be read from GOOGLE_CLIENT_ID env")
	assert.Equal(t, repo, uc.Repository, "Repository should be set")
	assert.Equal(t, 10*time.Second, uc.ContextTimeout, "ContextTimeout should be set")

	t.Logf("NewGoogleSigninAuthUseCase correctly reads GOOGLE_CLIENT_ID=%s", uc.GoogleClientID)
}

func TestNewGoogleSigninAuthUseCase_EmptyClientIDWhenUnset(t *testing.T) {
	os.Unsetenv("GOOGLE_CLIENT_ID")

	repo := &mockGoogleSigninRepository{}
	uc := NewGoogleSigninAuthUseCase(repo, 5*time.Second).(*GoogleSigninAuthUseCase)

	require.NotNil(t, uc, "UseCase should not be nil")
	assert.Empty(t, uc.GoogleClientID, "GoogleClientID should be empty when env unset")
	t.Logf("NewGoogleSigninAuthUseCase with unset env: GoogleClientID=%q", uc.GoogleClientID)
}

func TestGoogleSigninAuthUseCase_ImplementsInterface(t *testing.T) {
	repo := &mockGoogleSigninRepository{}
	uc := NewGoogleSigninAuthUseCase(repo, 10*time.Second)

	var _ _interface.IGoogleSigninAuthUseCase = uc
	t.Logf("GoogleSigninAuthUseCase implements IGoogleSigninAuthUseCase")
}

func TestGoogleSigninAuthUseCase_GoogleSignin_InvalidTokenReturnsUnauthorized(t *testing.T) {
	uc := NewGoogleSigninAuthUseCase(&mockGoogleSigninRepository{}, 10*time.Second).(*GoogleSigninAuthUseCase)
	uc.googleTokenValidator = func(ctx context.Context, idToken string, clientID string) (*idtoken.Payload, error) {
		return nil, errors.New("invalid token")
	}

	ctx := context.Background()
	res, err := uc.GoogleSignin(ctx, "invalid-fake-token")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "Invalid Google ID token")
	assert.Empty(t, res.AccessToken)
	t.Logf("Invalid token correctly returns Unauthorized: %v", err)
}

func TestGoogleSigninAuthUseCase_GoogleSignin_EmailMissingInTokenReturnsBadRequest(t *testing.T) {
	uc := NewGoogleSigninAuthUseCase(&mockGoogleSigninRepository{}, 10*time.Second).(*GoogleSigninAuthUseCase)
	uc.googleTokenValidator = func(ctx context.Context, idToken string, clientID string) (*idtoken.Payload, error) {
		return &idtoken.Payload{Claims: map[string]interface{}{}}, nil
	}

	ctx := context.Background()
	res, err := uc.GoogleSignin(ctx, "any-token")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "Email not found")
	assert.Empty(t, res.AccessToken)
	t.Logf("Missing email in token correctly returns BadRequest: %v", err)
}

func TestGoogleSigninAuthUseCase_GoogleSignin_RepoErrorReturnsInternalServerError(t *testing.T) {
	initJWTForTest(t)

	repoErr := errors.New("database error")
	repo := &mockGoogleSigninRepository{
		FindOrCreateUserByGoogleEmailFunc: func(ctx context.Context, email string, name string) (uint, error) {
			return 0, repoErr
		},
	}
	uc := NewGoogleSigninAuthUseCase(repo, 10*time.Second).(*GoogleSigninAuthUseCase)
	uc.googleTokenValidator = func(ctx context.Context, idToken string, clientID string) (*idtoken.Payload, error) {
		return &idtoken.Payload{
			Claims: map[string]interface{}{"email": "oauth@example.com", "name": "OAuth User"},
		}, nil
	}

	ctx := context.Background()
	res, err := uc.GoogleSignin(ctx, "valid-token")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "Failed to find or create user")
	assert.Empty(t, res.AccessToken)
	t.Logf("Repo error correctly returns InternalServerError: %v", err)
}

func TestGoogleSigninAuthUseCase_GoogleSignin_SuccessWithMock(t *testing.T) {
	initJWTForTest(t)

	var findOrCreateCalled bool
	repo := &mockGoogleSigninRepository{
		FindOrCreateUserByGoogleEmailFunc: func(ctx context.Context, email string, name string) (uint, error) {
			findOrCreateCalled = true
			assert.Equal(t, "oauth-success@example.com", email)
			assert.Equal(t, "OAuth Name", name)
			return 200, nil
		},
	}
	uc := NewGoogleSigninAuthUseCase(repo, 10*time.Second).(*GoogleSigninAuthUseCase)
	uc.googleTokenValidator = func(ctx context.Context, idToken string, clientID string) (*idtoken.Payload, error) {
		return &idtoken.Payload{
			Claims: map[string]interface{}{"email": "oauth-success@example.com", "name": "OAuth Name"},
		}, nil
	}

	ctx := context.Background()
	res, err := uc.GoogleSignin(ctx, "valid-token")
	require.NoError(t, err)
	require.True(t, findOrCreateCalled)
	assert.NotEmpty(t, res.AccessToken)
	assert.NotEmpty(t, res.RefreshToken)

	_, _, err = jwt.ParseToken(res.AccessToken)
	require.NoError(t, err)
	t.Logf("GoogleSignin success with mock: tokens generated")
}

func TestGoogleSigninAuthUseCase_GoogleSignin_NameFallbackToEmail(t *testing.T) {
	initJWTForTest(t)

	repo := &mockGoogleSigninRepository{
		FindOrCreateUserByGoogleEmailFunc: func(ctx context.Context, email string, name string) (uint, error) {
			assert.Equal(t, "noname@example.com", email)
			assert.Equal(t, "noname@example.com", name, "When name is empty in token, use email as name")
			return 1, nil
		},
	}
	uc := NewGoogleSigninAuthUseCase(repo, 10*time.Second).(*GoogleSigninAuthUseCase)
	uc.googleTokenValidator = func(ctx context.Context, idToken string, clientID string) (*idtoken.Payload, error) {
		return &idtoken.Payload{
			Claims: map[string]interface{}{"email": "noname@example.com"},
		}, nil
	}

	ctx := context.Background()
	res, err := uc.GoogleSignin(ctx, "valid-token")
	require.NoError(t, err)
	assert.NotEmpty(t, res.AccessToken)
	t.Logf("Name fallback to email works correctly")
}
