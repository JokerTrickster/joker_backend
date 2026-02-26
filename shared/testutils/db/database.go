package db

import (
	"fmt"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var (
	testDBOnce sync.Once
	testDBConn *gorm.DB
	testDBErr  error
)

// TestDBConfig holds configuration for test database
type TestDBConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
}

// GetDefaultTestDBConfig returns default test database configuration
func GetDefaultTestDBConfig() *TestDBConfig {
	return &TestDBConfig{
		Host:     getEnvOrDefault("TEST_DB_HOST", "localhost"),
		Port:     getEnvOrDefault("TEST_DB_PORT", "3307"),
		User:     getEnvOrDefault("TEST_DB_USER", "root"),
		Password: getEnvOrDefault("TEST_DB_PASSWORD", "rootpassword"),
		DBName:   fmt.Sprintf("test_%d", time.Now().Unix()),
	}
}

// SetupTestDB creates a new test database for isolated testing
func SetupTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	config := GetDefaultTestDBConfig()

	// Create test database
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/?charset=utf8mb4&parseTime=True&loc=Local",
		config.User, config.Password, config.Host, config.Port)

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err, "Failed to connect to MySQL")

	// Create test database
	err = db.Exec(fmt.Sprintf("CREATE DATABASE IF NOT EXISTS %s", config.DBName)).Error
	require.NoError(t, err, "Failed to create test database")

	// Connect to test database
	testDSN := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		config.User, config.Password, config.Host, config.Port, config.DBName)

	testDB, err := gorm.Open(mysql.Open(testDSN), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	require.NoError(t, err, "Failed to connect to test database")

	// Register cleanup
	t.Cleanup(func() {
		CleanupTestDB(testDB, config.DBName)
	})

	return testDB
}

// SetupSharedTestDB returns a shared test database connection (for faster tests)
func SetupSharedTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	testDBOnce.Do(func() {
		config := GetDefaultTestDBConfig()
		config.DBName = "test_shared"

		dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/?charset=utf8mb4&parseTime=True&loc=Local",
			config.User, config.Password, config.Host, config.Port)

		db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
			Logger: logger.Default.LogMode(logger.Silent),
		})
		if err != nil {
			testDBErr = err
			return
		}

		// Create shared test database
		err = db.Exec(fmt.Sprintf("CREATE DATABASE IF NOT EXISTS %s", config.DBName)).Error
		if err != nil {
			testDBErr = err
			return
		}

		testDSN := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
			config.User, config.Password, config.Host, config.Port, config.DBName)

		testDBConn, testDBErr = gorm.Open(mysql.Open(testDSN), &gorm.Config{
			Logger: logger.Default.LogMode(logger.Silent),
		})
	})

	require.NoError(t, testDBErr, "Failed to setup shared test database")
	return testDBConn
}

// CleanupTestDB drops the test database
func CleanupTestDB(db *gorm.DB, dbName string) {
	sqlDB, _ := db.DB()
	if sqlDB != nil {
		sqlDB.Close()
	}

	// Drop test database
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/?charset=utf8mb4&parseTime=True&loc=Local",
		getEnvOrDefault("TEST_DB_USER", "root"),
		getEnvOrDefault("TEST_DB_PASSWORD", "rootpassword"),
		getEnvOrDefault("TEST_DB_HOST", "localhost"),
		getEnvOrDefault("TEST_DB_PORT", "3307"))

	cleanupDB, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err == nil {
		cleanupDB.Exec(fmt.Sprintf("DROP DATABASE IF EXISTS %s", dbName))
	}
}

// TruncateTables clears all data from specified tables
func TruncateTables(db *gorm.DB, tables ...string) error {
	for _, table := range tables {
		if err := db.Exec(fmt.Sprintf("TRUNCATE TABLE %s", table)).Error; err != nil {
			return fmt.Errorf("failed to truncate table %s: %w", table, err)
		}
	}
	return nil
}

// RunMigrations runs database migrations for testing
func RunMigrations(db *gorm.DB, models ...interface{}) error {
	return db.AutoMigrate(models...)
}

// WithTransaction wraps test in a transaction that gets rolled back
func WithTransaction(t *testing.T, db *gorm.DB, fn func(*gorm.DB)) {
	t.Helper()

	tx := db.Begin()
	defer tx.Rollback()

	fn(tx)
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}