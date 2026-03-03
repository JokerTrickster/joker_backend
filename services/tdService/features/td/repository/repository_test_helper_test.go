package repository

import (
	"os"
	"testing"

	"github.com/JokerTrickster/joker_backend/services/tdService/features/td/model/entity"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// setupTestDB initializes a test database connection for integration tests.
// Skips the test if the database is unavailable.
// Auto-migrates tables and cleans them before each test to ensure isolation.
func setupTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	dsn := os.Getenv("TEST_DB_DSN")
	if dsn == "" {
		dsn = "root:rootpassword@tcp(localhost:3307)/test_td_service?charset=utf8mb4&parseTime=True&loc=Local"
	}
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Skipf("Integration test: requires test database: %v", err)
	}

	err = db.AutoMigrate(
		&entity.Player{},
		&entity.PlayerStats{},
		&entity.GameSession{},
		&entity.GameResult{},
	)
	if err != nil {
		t.Skipf("Integration test: migration failed: %v", err)
	}

	// Clean tables in order to respect foreign key constraints
	db.Exec("DELETE FROM game_results")
	db.Exec("DELETE FROM game_sessions")
	db.Exec("DELETE FROM player_stats")
	db.Exec("DELETE FROM players")

	return db
}
