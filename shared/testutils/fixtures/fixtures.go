package fixtures

import (
	"encoding/json"
	"fmt"
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// UserFixture represents a test user
type UserFixture struct {
	ID           uint      `json:"id"`
	Email        string    `json:"email"`
	Password     string    `json:"password"` // Plain text password for testing
	PasswordHash string    `json:"-"`        // Hashed password stored in DB
	Name         string    `json:"name"`
	ServiceType  string    `json:"service_type"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// RoomFixture represents a test room
type RoomFixture struct {
	ID          uint      `json:"id"`
	Name        string    `json:"name"`
	MaxPlayers  int       `json:"max_players"`
	Status      string    `json:"status"`
	HostUserID  uint      `json:"host_user_id"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// FileFixture represents a test file
type FileFixture struct {
	ID         uint      `json:"id"`
	UserID     uint      `json:"user_id"`
	FileName   string    `json:"file_name"`
	FileSize   int64     `json:"file_size"`
	FileType   string    `json:"file_type"`
	FilePath   string    `json:"file_path"`
	IsPublic   bool      `json:"is_public"`
	CreatedAt  time.Time `json:"created_at"`
}

// FixtureLoader manages loading test data
type FixtureLoader struct {
	db *gorm.DB
}

// NewFixtureLoader creates a new fixture loader
func NewFixtureLoader(db *gorm.DB) *FixtureLoader {
	return &FixtureLoader{db: db}
}

// LoadUser creates a test user with hashed password
func (f *FixtureLoader) LoadUser(user *UserFixture) (*UserFixture, error) {
	if user.Password != "" && user.PasswordHash == "" {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
		if err != nil {
			return nil, fmt.Errorf("failed to hash password: %w", err)
		}
		user.PasswordHash = string(hashedPassword)
	}

	if user.CreatedAt.IsZero() {
		user.CreatedAt = time.Now()
		user.UpdatedAt = time.Now()
	}

	// Insert into database
	query := `
		INSERT INTO users (email, password, name, service_type, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?)
	`
	result := f.db.Exec(query, user.Email, user.PasswordHash, user.Name, user.ServiceType, user.CreatedAt, user.UpdatedAt)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to insert user: %w", result.Error)
	}

	// Get the inserted ID
	var lastID uint
	f.db.Raw("SELECT LAST_INSERT_ID()").Scan(&lastID)
	user.ID = lastID

	return user, nil
}

// LoadUsers creates multiple test users
func (f *FixtureLoader) LoadUsers(users []*UserFixture) ([]*UserFixture, error) {
	var result []*UserFixture
	for _, user := range users {
		loadedUser, err := f.LoadUser(user)
		if err != nil {
			return nil, err
		}
		result = append(result, loadedUser)
	}
	return result, nil
}

// LoadRoom creates a test room
func (f *FixtureLoader) LoadRoom(room *RoomFixture) (*RoomFixture, error) {
	if room.CreatedAt.IsZero() {
		room.CreatedAt = time.Now()
		room.UpdatedAt = time.Now()
	}

	query := `
		INSERT INTO rooms (name, max_players, status, host_user_id, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?)
	`
	result := f.db.Exec(query, room.Name, room.MaxPlayers, room.Status, room.HostUserID, room.CreatedAt, room.UpdatedAt)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to insert room: %w", result.Error)
	}

	var lastID uint
	f.db.Raw("SELECT LAST_INSERT_ID()").Scan(&lastID)
	room.ID = lastID

	return room, nil
}

// LoadFile creates a test file record
func (f *FixtureLoader) LoadFile(file *FileFixture) (*FileFixture, error) {
	if file.CreatedAt.IsZero() {
		file.CreatedAt = time.Now()
	}

	query := `
		INSERT INTO files (user_id, file_name, file_size, file_type, file_path, is_public, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`
	result := f.db.Exec(query, file.UserID, file.FileName, file.FileSize, file.FileType, file.FilePath, file.IsPublic, file.CreatedAt)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to insert file: %w", result.Error)
	}

	var lastID uint
	f.db.Raw("SELECT LAST_INSERT_ID()").Scan(&lastID)
	file.ID = lastID

	return file, nil
}

// LoadFromJSON loads fixtures from JSON data
func (f *FixtureLoader) LoadFromJSON(jsonData []byte, target interface{}) error {
	return json.Unmarshal(jsonData, target)
}

// GetDefaultUsers returns a set of default test users
func GetDefaultUsers() []*UserFixture {
	return []*UserFixture{
		{
			Email:       "test-user-1@example.com",
			Password:    "password123",
			Name:        "Test User 1",
			ServiceType: "game",
		},
		{
			Email:       "test-user-2@example.com",
			Password:    "password456",
			Name:        "Test User 2",
			ServiceType: "game",
		},
		{
			Email:       "admin@example.com",
			Password:    "admin123",
			Name:        "Admin User",
			ServiceType: "admin",
		},
	}
}

// GetDefaultRoom returns a default test room
func GetDefaultRoom(hostUserID uint) *RoomFixture {
	return &RoomFixture{
		Name:       "Test Room",
		MaxPlayers: 4,
		Status:     "waiting",
		HostUserID: hostUserID,
	}
}