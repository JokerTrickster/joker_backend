package migrate

import (
	"database/sql"
	"errors"
	"fmt"
	"path/filepath"

	"main/shared/logger"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/mysql"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"go.uber.org/zap"
)

// Config holds migration configuration
type Config struct {
	MigrationsPath string // Path to migrations directory
	DatabaseName   string // Database name for migration tracking
}

// Run executes database migrations
func Run(db *sql.DB, config Config) error {
	logger.Info("Starting database migration",
		zap.String("migrations_path", config.MigrationsPath),
		zap.String("database", config.DatabaseName),
	)

	// Create MySQL driver instance
	driver, err := mysql.WithInstance(db, &mysql.Config{
		DatabaseName: config.DatabaseName,
		NoLock:       true, // Don't lock the database connection
	})
	if err != nil {
		logger.Error("Failed to create migration driver", zap.Error(err))
		return fmt.Errorf("failed to create migration driver: %w", err)
	}

	// Get absolute path to migrations
	absPath, err := filepath.Abs(config.MigrationsPath)
	if err != nil {
		logger.Error("Failed to get absolute path", zap.Error(err))
		return fmt.Errorf("failed to get absolute path: %w", err)
	}

	// Create migrate instance
	m, err := migrate.NewWithDatabaseInstance(
		fmt.Sprintf("file://%s", absPath),
		config.DatabaseName,
		driver,
	)
	if err != nil {
		logger.Error("Failed to create migrate instance", zap.Error(err))
		return fmt.Errorf("failed to create migrate instance: %w", err)
	}
	// Don't defer m.Close() to avoid closing the database connection

	// Get current version
	version, dirty, err := m.Version()
	if err != nil && !errors.Is(err, migrate.ErrNilVersion) {
		logger.Error("Failed to get current version", zap.Error(err))
		return fmt.Errorf("failed to get current version: %w", err)
	}

	if dirty {
		logger.Warn("Database is in dirty state, forcing version",
			zap.Uint("version", version),
		)
		if err := m.Force(int(version)); err != nil {
			logger.Error("Failed to force version", zap.Error(err))
			return fmt.Errorf("failed to force version: %w", err)
		}
	}

	logger.Info("Current migration version",
		zap.Uint("version", version),
		zap.Bool("dirty", dirty),
	)

	// Run migrations
	if err := m.Up(); err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			logger.Info("No new migrations to apply")
			return nil
		}
		logger.Error("Failed to run migrations", zap.Error(err))
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	// Get new version
	newVersion, _, err := m.Version()
	if err != nil {
		logger.Error("Failed to get new version", zap.Error(err))
		return fmt.Errorf("failed to get new version: %w", err)
	}

	logger.Info("Migration completed successfully",
		zap.Uint("from_version", version),
		zap.Uint("to_version", newVersion),
	)

	return nil
}

// Down rolls back the last migration
func Down(db *sql.DB, config Config) error {
	logger.Info("Rolling back last migration",
		zap.String("migrations_path", config.MigrationsPath),
		zap.String("database", config.DatabaseName),
	)

	driver, err := mysql.WithInstance(db, &mysql.Config{
		DatabaseName: config.DatabaseName,
		NoLock:       true,
	})
	if err != nil {
		return fmt.Errorf("failed to create migration driver: %w", err)
	}

	absPath, err := filepath.Abs(config.MigrationsPath)
	if err != nil {
		return fmt.Errorf("failed to get absolute path: %w", err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		fmt.Sprintf("file://%s", absPath),
		config.DatabaseName,
		driver,
	)
	if err != nil {
		return fmt.Errorf("failed to create migrate instance: %w", err)
	}

	if err := m.Steps(-1); err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			logger.Info("No migrations to rollback")
			return nil
		}
		return fmt.Errorf("failed to rollback migration: %w", err)
	}

	logger.Info("Rollback completed successfully")
	return nil
}

// Version returns the current migration version
func Version(db *sql.DB, config Config) (uint, bool, error) {
	driver, err := mysql.WithInstance(db, &mysql.Config{
		DatabaseName: config.DatabaseName,
		NoLock:       true,
	})
	if err != nil {
		return 0, false, fmt.Errorf("failed to create migration driver: %w", err)
	}

	absPath, err := filepath.Abs(config.MigrationsPath)
	if err != nil {
		return 0, false, fmt.Errorf("failed to get absolute path: %w", err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		fmt.Sprintf("file://%s", absPath),
		config.DatabaseName,
		driver,
	)
	if err != nil {
		return 0, false, fmt.Errorf("failed to create migrate instance: %w", err)
	}

	version, dirty, err := m.Version()
	if err != nil && !errors.Is(err, migrate.ErrNilVersion) {
		return 0, false, fmt.Errorf("failed to get version: %w", err)
	}

	return version, dirty, nil
}
