package usecase

import (
	"errors"
	"testing"

	"github.com/JokerTrickster/joker_backend/services/tdService/features/td/model/entity"
	"github.com/JokerTrickster/joker_backend/services/tdService/features/td/model/request"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestAuthUseCase_Login_Success(t *testing.T) {
	t.Log("AuthUseCase.Login: GetOrCreatePlayer returns player -> verify token and response")
	mockPlayer := new(MockPlayerRepository)
	uc := newTestAuthUseCase(mockPlayer)

	player := &entity.Player{
		ID:          testPlayerID,
		UserID:      testUserID,
		Nickname:    testUsername,
		AvatarID:    "default_avatar",
		Level:       1,
		Experience:  0,
		Stats:       entity.PlayerStats{},
	}
	mockPlayer.On("GetOrCreatePlayer", mock.AnythingOfType("uint"), testUsername, "default_avatar").
		Return(player, nil)

	req := &request.LoginRequest{Username: testUsername, Password: "password123"}
	resp, err := uc.Login(req)

	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.NotEmpty(t, resp.Token, "token should not be empty")
	assert.NotEmpty(t, resp.UserID, "userId should not be empty")
	assert.Equal(t, testUsername, resp.Profile.Nickname)
	assert.Equal(t, "default_avatar", resp.Profile.AvatarID)
	assert.Equal(t, 1, resp.Profile.Level)
	assert.Equal(t, 0, resp.Profile.Experience)
	assert.Nil(t, resp.Profile.Stats, "stats should be nil when player has no stats")
	mockPlayer.AssertExpectations(t)
}

func TestAuthUseCase_Login_WithStats(t *testing.T) {
	t.Log("AuthUseCase.Login: player has stats -> verify stats in response")
	mockPlayer := new(MockPlayerRepository)
	uc := newTestAuthUseCase(mockPlayer)

	player := &entity.Player{
		ID:         testPlayerID,
		UserID:     testUserID,
		Nickname:   testUsername,
		AvatarID:   "default_avatar",
		Level:      5,
		Experience: 100,
		Stats: entity.PlayerStats{
			ID:             1,
			PlayerID:        testPlayerID,
			GamesPlayed:    10,
			Victories:      7,
			TotalScore:     50000,
			HighestScore:   8000,
			HighestWave:    15,
			UnitsPlaced:    100,
			EnemiesKilled:  500,
		},
	}
	mockPlayer.On("GetOrCreatePlayer", mock.AnythingOfType("uint"), testUsername, "default_avatar").
		Return(player, nil)

	req := &request.LoginRequest{Username: testUsername, Password: "password123"}
	resp, err := uc.Login(req)

	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, 5, resp.Profile.Level)
	require.NotNil(t, resp.Profile.Stats)
	assert.Equal(t, 10, resp.Profile.Stats.GamesPlayed)
	assert.Equal(t, 7, resp.Profile.Stats.Victories)
	assert.Equal(t, int64(50000), resp.Profile.Stats.TotalScore)
	assert.Equal(t, int64(8000), resp.Profile.Stats.HighestScore)
	assert.Equal(t, 15, resp.Profile.Stats.HighestWave)
	assert.Equal(t, 100, resp.Profile.Stats.UnitsPlaced)
	assert.Equal(t, 500, resp.Profile.Stats.EnemiesDefeated)
	mockPlayer.AssertExpectations(t)
}

func TestAuthUseCase_Login_RepoError(t *testing.T) {
	t.Log("AuthUseCase.Login: GetOrCreatePlayer returns error -> verify error")
	mockPlayer := new(MockPlayerRepository)
	uc := newTestAuthUseCase(mockPlayer)

	repoErr := errors.New("database connection failed")
	mockPlayer.On("GetOrCreatePlayer", mock.AnythingOfType("uint"), testUsername, "default_avatar").
		Return(nil, repoErr)

	req := &request.LoginRequest{Username: testUsername, Password: "password123"}
	resp, err := uc.Login(req)

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "failed to get player profile")
	assert.ErrorIs(t, err, repoErr)
	mockPlayer.AssertExpectations(t)
}

func TestAuthUseCase_RefreshToken_Success(t *testing.T) {
	t.Log("AuthUseCase.RefreshToken: valid token -> verify new token returned")
	mockPlayer := new(MockPlayerRepository)
	uc := newTestAuthUseCase(mockPlayer)

	validToken := createValidJWT(t, testUserID, testUsername)

	req := &request.RefreshTokenRequest{RefreshToken: validToken}
	resp, err := uc.RefreshToken(req)

	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.NotEmpty(t, resp.Token)
	mockPlayer.AssertNotCalled(t, "GetOrCreatePlayer")
}

func TestAuthUseCase_RefreshToken_InvalidToken(t *testing.T) {
	t.Log("AuthUseCase.RefreshToken: bad token string -> error")
	mockPlayer := new(MockPlayerRepository)
	uc := newTestAuthUseCase(mockPlayer)

	req := &request.RefreshTokenRequest{RefreshToken: "invalid.jwt.token"}
	resp, err := uc.RefreshToken(req)

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "invalid refresh token")
	mockPlayer.AssertNotCalled(t, "GetOrCreatePlayer")
}

func TestAuthUseCase_ValidateTokenAndGetUserID_Success(t *testing.T) {
	t.Log("AuthUseCase.ValidateTokenAndGetUserID: valid token -> correct userID")
	mockPlayer := new(MockPlayerRepository)
	uc := newTestAuthUseCase(mockPlayer)

	validToken := createValidJWT(t, testUserID, testUsername)

	userID, err := uc.ValidateTokenAndGetUserID(validToken)

	require.NoError(t, err)
	assert.Equal(t, testUserID, userID)
	mockPlayer.AssertNotCalled(t, "GetOrCreatePlayer")
}

func TestAuthUseCase_ValidateTokenAndGetUserID_InvalidToken(t *testing.T) {
	t.Log("AuthUseCase.ValidateTokenAndGetUserID: bad token -> error")
	mockPlayer := new(MockPlayerRepository)
	uc := newTestAuthUseCase(mockPlayer)

	userID, err := uc.ValidateTokenAndGetUserID("not-a-valid-jwt")

	assert.Error(t, err)
	assert.Equal(t, uint(0), userID)
	mockPlayer.AssertNotCalled(t, "GetOrCreatePlayer")
}

func TestAuthUseCase_ValidateTokenAndGetUserID_ExpiredToken(t *testing.T) {
	t.Log("AuthUseCase.ValidateTokenAndGetUserID: expired token -> error")
	mockPlayer := new(MockPlayerRepository)
	uc := newTestAuthUseCase(mockPlayer)

	expiredToken := createExpiredJWT(t, testUserID, testUsername)

	userID, err := uc.ValidateTokenAndGetUserID(expiredToken)

	assert.Error(t, err)
	assert.Equal(t, uint(0), userID)
	mockPlayer.AssertNotCalled(t, "GetOrCreatePlayer")
}
