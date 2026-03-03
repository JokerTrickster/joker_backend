package fixtures

import (
	"time"
)

// User fixtures for testing
type UserFixture struct {
	ID        uint
	Email     string
	Nickname  string
	Password  string // hashed password for tests
	CreatedAt time.Time
}

// GetValidUser returns a valid user fixture
func GetValidUser() *UserFixture {
	return &UserFixture{
		ID:        1,
		Email:     "test@example.com",
		Nickname:  "testuser",
		Password:  "$2a$10$YourHashedPasswordHere", // bcrypt hash
		CreatedAt: time.Now(),
	}
}

// GetInvalidUsers returns various invalid user fixtures for error testing
func GetInvalidUsers() []UserFixture {
	return []UserFixture{
		{Email: "", Nickname: "test"},           // empty email
		{Email: "invalid", Nickname: "test"},    // invalid email format
		{Email: "test@test.com", Nickname: ""}, // empty nickname
		{Email: "test@test.com", Nickname: "ab"}, // nickname too short
	}
}

// Game session fixtures
type GameSessionFixture struct {
	ID          uint
	SessionID   string
	RoomCode    string
	Mode        string
	Status      string
	PlayerCount int
	PlayerID    uint
	CreatedAt   time.Time
}

// GetValidGameSession returns a valid game session fixture
func GetValidGameSession() *GameSessionFixture {
	return &GameSessionFixture{
		ID:          1,
		SessionID:   "session-123-456",
		RoomCode:    "ROOM001",
		Mode:        "single",
		Status:      "waiting",
		PlayerCount: 1,
		PlayerID:    1,
		CreatedAt:   time.Now(),
	}
}

// GetGameSessionsForTesting returns various game session states for testing
func GetGameSessionsForTesting() []GameSessionFixture {
	return []GameSessionFixture{
		{ID: 1, SessionID: "s1", Status: "waiting", PlayerCount: 1},
		{ID: 2, SessionID: "s2", Status: "playing", PlayerCount: 2},
		{ID: 3, SessionID: "s3", Status: "completed", PlayerCount: 4},
		{ID: 4, SessionID: "s4", Status: "cancelled", PlayerCount: 1},
	}
}

// Auth token fixtures
type TokenFixture struct {
	UserID      uint
	Token       string
	RefreshToken string
	ExpiresAt   time.Time
}

// GetValidToken returns a valid token fixture
func GetValidToken() *TokenFixture {
	return &TokenFixture{
		UserID:       1,
		Token:        "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.valid",
		RefreshToken: "refresh-token-123",
		ExpiresAt:    time.Now().Add(24 * time.Hour),
	}
}

// GetExpiredToken returns an expired token for testing
func GetExpiredToken() *TokenFixture {
	return &TokenFixture{
		UserID:       1,
		Token:        "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.expired",
		RefreshToken: "refresh-token-expired",
		ExpiresAt:    time.Now().Add(-1 * time.Hour), // expired 1 hour ago
	}
}

// Room fixtures for multiplayer testing
type RoomFixture struct {
	ID          uint
	Code        string
	HostID      uint
	MaxPlayers  int
	CurrentPlayers int
	Status      string
	CreatedAt   time.Time
}

// GetValidRoom returns a valid room fixture
func GetValidRoom() *RoomFixture {
	return &RoomFixture{
		ID:          1,
		Code:        "ABC123",
		HostID:      1,
		MaxPlayers:  4,
		CurrentPlayers: 1,
		Status:      "waiting",
		CreatedAt:   time.Now(),
	}
}

// GetFullRoom returns a room that's full
func GetFullRoom() *RoomFixture {
	return &RoomFixture{
		ID:          2,
		Code:        "FULL01",
		HostID:      1,
		MaxPlayers:  4,
		CurrentPlayers: 4,
		Status:      "full",
		CreatedAt:   time.Now(),
	}
}