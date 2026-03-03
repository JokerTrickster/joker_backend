package usecase

import (
	"strconv"
	"testing"
	"time"

	"github.com/JokerTrickster/joker_backend/services/tdService/features/td/model/entity"
	iface "github.com/JokerTrickster/joker_backend/services/tdService/features/td/model/interface"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

const (
	testJWTSecret  = "test-jwt-secret"
	testUserID     = uint(42)
	testPlayerID   = uint(1)
	testSessionID  = "test-session-id-123"
	testRoomCode   = "ABC123"
	testWsBaseURL  = "ws://localhost:8080"
	testUsername   = "testuser"
)

// MockPlayerRepository implements iface.PlayerRepository using testify/mock
type MockPlayerRepository struct {
	mock.Mock
}

func (m *MockPlayerRepository) CreatePlayer(player *entity.Player) error {
	args := m.Called(player)
	return args.Error(0)
}

func (m *MockPlayerRepository) GetPlayerByUserID(userID uint) (*entity.Player, error) {
	args := m.Called(userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Player), args.Error(1)
}

func (m *MockPlayerRepository) GetPlayerByID(playerID uint) (*entity.Player, error) {
	args := m.Called(playerID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Player), args.Error(1)
}

func (m *MockPlayerRepository) UpdatePlayer(player *entity.Player) error {
	args := m.Called(player)
	return args.Error(0)
}

func (m *MockPlayerRepository) CreateOrUpdateStats(stats *entity.PlayerStats) error {
	args := m.Called(stats)
	return args.Error(0)
}

func (m *MockPlayerRepository) UpdatePlayerStats(playerID uint, result *entity.GameResult) error {
	args := m.Called(playerID, result)
	return args.Error(0)
}

func (m *MockPlayerRepository) GetOrCreatePlayer(userID uint, nickname string, avatarID string) (*entity.Player, error) {
	args := m.Called(userID, nickname, avatarID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Player), args.Error(1)
}

// MockGameSessionRepository implements iface.GameSessionRepository using testify/mock
type MockGameSessionRepository struct {
	mock.Mock
}

func (m *MockGameSessionRepository) CreateSession(session *entity.GameSession) error {
	args := m.Called(session)
	return args.Error(0)
}

func (m *MockGameSessionRepository) GetSessionByID(sessionID string) (*entity.GameSession, error) {
	args := m.Called(sessionID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.GameSession), args.Error(1)
}

func (m *MockGameSessionRepository) GetSessionByRoomCode(roomCode string) (*entity.GameSession, error) {
	args := m.Called(roomCode)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.GameSession), args.Error(1)
}

func (m *MockGameSessionRepository) UpdateSession(session *entity.GameSession) error {
	args := m.Called(session)
	return args.Error(0)
}

func (m *MockGameSessionRepository) JoinSession(sessionID string, playerID uint) error {
	args := m.Called(sessionID, playerID)
	return args.Error(0)
}

func (m *MockGameSessionRepository) StartSession(sessionID string) error {
	args := m.Called(sessionID)
	return args.Error(0)
}

func (m *MockGameSessionRepository) EndSession(sessionID string) error {
	args := m.Called(sessionID)
	return args.Error(0)
}

func (m *MockGameSessionRepository) GetActiveSessions(playerID uint) ([]*entity.GameSession, error) {
	args := m.Called(playerID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.GameSession), args.Error(1)
}

func (m *MockGameSessionRepository) SaveGameResult(result *entity.GameResult) error {
	args := m.Called(result)
	return args.Error(0)
}

func (m *MockGameSessionRepository) GetGameResults(sessionID string) ([]*entity.GameResult, error) {
	args := m.Called(sessionID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.GameResult), args.Error(1)
}

func (m *MockGameSessionRepository) GetPlayerGameHistory(playerID uint, limit int) ([]*entity.GameResult, error) {
	args := m.Called(playerID, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.GameResult), args.Error(1)
}

// Ensure both mocks implement their interfaces
var (
	_ iface.PlayerRepository     = (*MockPlayerRepository)(nil)
	_ iface.GameSessionRepository = (*MockGameSessionRepository)(nil)
)

// newTestAuthUseCase creates AuthUseCase with mock player repo and test JWT secret
func newTestAuthUseCase(mockPlayer iface.PlayerRepository) *AuthUseCase {
	return NewAuthUseCase(mockPlayer, testJWTSecret)
}

// newTestGameUseCase creates GameUseCase with mock repos
func newTestGameUseCase(mockSession iface.GameSessionRepository, mockPlayer iface.PlayerRepository) *GameUseCase {
	return NewGameUseCase(mockSession, mockPlayer, testWsBaseURL)
}

// createValidJWT creates a valid JWT token for the given userID and username
func createValidJWT(t *testing.T, userID uint, username string) string {
	t.Helper()
	return createJWTWithExp(t, userID, username, time.Now().Add(24*time.Hour).Unix())
}

// createExpiredJWT creates an expired JWT token
func createExpiredJWT(t *testing.T, userID uint, username string) string {
	t.Helper()
	return createJWTWithExp(t, userID, username, time.Now().Add(-1*time.Hour).Unix())
}

func createJWTWithExp(t *testing.T, userID uint, username string, exp int64) string {
	t.Helper()
	claims := jwt.MapClaims{
		"userID":   strconv.FormatUint(uint64(userID), 10),
		"username": username,
		"exp":      exp,
		"iat":      time.Now().Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	s, err := token.SignedString([]byte(testJWTSecret))
	require.NoError(t, err)
	return s
}
