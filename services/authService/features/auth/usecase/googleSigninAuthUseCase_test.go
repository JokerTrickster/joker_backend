package usecase

import (
	"context"
	"os"
	"testing"
	"time"

	_interface "github.com/JokerTrickster/joker_backend/services/authService/features/auth/model/interface"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
	testClientID := "test-client-id-" + time.Now().Format("20060102150405")
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

func TestGoogleSigninAuthUseCase_GoogleSignin_RequiresValidToken(t *testing.T) {
	t.Skip("idtoken.Validate requires real Google APIs; cannot validate token in unit tests without external calls")

	repo := &mockGoogleSigninRepository{
		FindOrCreateUserByGoogleEmailFunc: func(ctx context.Context, email string, name string) (uint, error) {
			return 1, nil
		},
	}
	uc := NewGoogleSigninAuthUseCase(repo, 10*time.Second)

	ctx := context.Background()
	_, err := uc.GoogleSignin(ctx, "invalid-fake-token")
	require.Error(t, err, "Invalid token should fail")
}
