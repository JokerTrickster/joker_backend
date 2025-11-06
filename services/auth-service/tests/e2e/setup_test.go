package e2e

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/luxrobo/joker_backend/services/auth-service/internal/handler"
	"github.com/luxrobo/joker_backend/shared/config"
	"github.com/luxrobo/joker_backend/shared/database"
	customErrors "github.com/luxrobo/joker_backend/shared/errors"
	"github.com/luxrobo/joker_backend/shared/logger"
	customMiddleware "github.com/luxrobo/joker_backend/shared/middleware"
	"github.com/luxrobo/joker_backend/shared/migrate"
)

var (
	testDB     *database.DB
	testServer *echo.Echo
)

func TestMain(m *testing.M) {
	// Setup
	if err := setupTestEnvironment(); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to setup test environment: %v\n", err)
		os.Exit(1)
	}

	// Run tests
	code := m.Run()

	// Teardown
	teardownTestEnvironment()

	os.Exit(code)
}

func setupTestEnvironment() error {
	// Set test environment variables
	os.Setenv("DB_HOST", getEnv("TEST_DB_HOST", "localhost"))
	os.Setenv("DB_PORT", getEnv("TEST_DB_PORT", "3306"))
	os.Setenv("DB_USER", getEnv("TEST_DB_USER", "joker_user"))
	os.Setenv("DB_PASSWORD", getEnv("TEST_DB_PASSWORD", "joker_password"))
	os.Setenv("DB_NAME", getEnv("TEST_DB_NAME", "backend_dev_test"))
	os.Setenv("LOG_LEVEL", "error") // Reduce noise in tests

	// Initialize logger
	logger.Init("error")

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Connect to database
	db, err := database.Connect(cfg.Database)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	testDB = db

	// Create test database if not exists
	if err := createTestDatabase(); err != nil {
		return fmt.Errorf("failed to create test database: %w", err)
	}

	// Run migrations using migrate package
	migrateConfig := migrate.Config{
		MigrationsPath: "../../../../migrations",
		DatabaseName:   getEnv("TEST_DB_NAME", "backend_dev_test"),
	}
	if err := migrate.Run(testDB.DB, migrateConfig); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	// Setup Echo server
	testServer = setupEchoServer()

	return nil
}

func teardownTestEnvironment() {
	if testDB != nil {
		// Clean up test data
		cleanupTestData()
		testDB.Close()
	}
	logger.Sync()
}

func setupEchoServer() *echo.Echo {
	e := echo.New()
	e.HideBanner = true
	e.HidePort = true

	// Set custom error handler
	e.HTTPErrorHandler = customErrors.CustomErrorHandler

	// Core middleware
	e.Use(customMiddleware.RequestID())
	e.Use(customMiddleware.Recovery())

	// Health check endpoint
	e.GET("/health", func(c echo.Context) error {
		return c.JSON(200, map[string]interface{}{
			"success": true,
			"message": "Test server is running",
		})
	})

	// API v1 routes
	v1 := e.Group("/api/v1")
	handler.RegisterRoutes(v1, testDB)

	return e
}

func createTestDatabase() error {
	// Create test database if it doesn't exist
	createDBSQL := "CREATE DATABASE IF NOT EXISTS backend_dev_test CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci"
	_, err := testDB.Exec(createDBSQL)
	return err
}

func cleanupTestData() {
	// Clean all test data
	testDB.Exec("DELETE FROM users")
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// Test fixtures

type TestUser struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func createTestUser(name, email string) (*TestUser, error) {
	query := "INSERT INTO users (name, email) VALUES (?, ?)"
	result, err := testDB.Exec(query, name, email)
	if err != nil {
		return nil, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	return &TestUser{
		ID:        id,
		Name:      name,
		Email:     email,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}, nil
}

func deleteTestUser(id int64) error {
	_, err := testDB.Exec("DELETE FROM users WHERE id = ?", id)
	return err
}

func countUsers() (int, error) {
	var count int
	err := testDB.QueryRow("SELECT COUNT(*) FROM users").Scan(&count)
	return count, err
}
