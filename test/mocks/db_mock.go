package mocks

import (
	"database/sql"
	"github.com/DATA-DOG/go-sqlmock"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"testing"
)

// MockDB creates a mock database for testing
type MockDB struct {
	DB   *gorm.DB
	Mock sqlmock.Sqlmock
	SqlDB *sql.DB
}

// NewMockDB creates a new mock database instance
func NewMockDB(t *testing.T) *MockDB {
	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock database: %v", err)
	}

	// Create GORM DB with the mock SQL DB
	gormDB, err := gorm.Open(mysql.New(mysql.Config{
		Conn: sqlDB,
		SkipInitializeWithVersion: true,
	}), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to open gorm database: %v", err)
	}

	return &MockDB{
		DB:   gormDB,
		Mock: mock,
		SqlDB: sqlDB,
	}
}

// Close closes the mock database connection
func (m *MockDB) Close() error {
	return m.SqlDB.Close()
}

// ExpectBegin sets expectation for transaction begin
func (m *MockDB) ExpectBegin() {
	m.Mock.ExpectBegin()
}

// ExpectCommit sets expectation for transaction commit
func (m *MockDB) ExpectCommit() {
	m.Mock.ExpectCommit()
}

// ExpectRollback sets expectation for transaction rollback
func (m *MockDB) ExpectRollback() {
	m.Mock.ExpectRollback()
}

// MockRedisClient creates a mock Redis client for testing
type MockRedisClient struct {
	Data map[string]string
	TTL  map[string]int64
}

// NewMockRedisClient creates a new mock Redis client
func NewMockRedisClient() *MockRedisClient {
	return &MockRedisClient{
		Data: make(map[string]string),
		TTL:  make(map[string]int64),
	}
}

// Get simulates Redis GET command
func (m *MockRedisClient) Get(key string) (string, error) {
	if val, exists := m.Data[key]; exists {
		return val, nil
	}
	return "", sql.ErrNoRows
}

// Set simulates Redis SET command
func (m *MockRedisClient) Set(key string, value string, ttl int64) error {
	m.Data[key] = value
	m.TTL[key] = ttl
	return nil
}

// Delete simulates Redis DEL command
func (m *MockRedisClient) Delete(key string) error {
	delete(m.Data, key)
	delete(m.TTL, key)
	return nil
}

// Exists simulates Redis EXISTS command
func (m *MockRedisClient) Exists(key string) bool {
	_, exists := m.Data[key]
	return exists
}

// FlushAll simulates Redis FLUSHALL command
func (m *MockRedisClient) FlushAll() error {
	m.Data = make(map[string]string)
	m.TTL = make(map[string]int64)
	return nil
}