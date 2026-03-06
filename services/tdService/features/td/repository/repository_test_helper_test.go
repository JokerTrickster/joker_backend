package repository

import (
	"fmt"
	"math/rand"
	"os"
	"testing"
	"time"

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
		dsn = "root:rootpassword@tcp(localhost:3307)/test_db?charset=utf8mb4&parseTime=True&loc=Local"
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

// uniqueSuffix returns a unique string for test data isolation.
func uniqueSuffix() string {
	return fmt.Sprintf("%d_%d", time.Now().UnixNano(), rand.Intn(100000))
}

// uniqueUserID returns a unique uint for test player UserID.
func uniqueUserID() uint {
	return uint(1000000 + rand.Intn(8999999))
}

// deferCleanupPlayers registers cleanup of created players and their stats.
func deferCleanupPlayers(t *testing.T, db *gorm.DB, playerIDs []uint) {
	t.Helper()
	t.Cleanup(func() {
		for _, pid := range playerIDs {
			db.Where("player_id = ?", pid).Delete(&entity.PlayerStats{})
			db.Delete(&entity.Player{}, pid)
		}
	})
}
