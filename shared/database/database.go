package database

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/luxrobo/joker_backend/shared/config"
)

type DB struct {
	*sql.DB
}

func Connect(cfg config.DatabaseConfig) (*DB, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true&charset=utf8mb4",
		cfg.User,
		cfg.Password,
		cfg.Host,
		cfg.Port,
		cfg.Database,
	)

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// Connection pool settings
	db.SetMaxOpenConns(25)                           // Maximum open connections
	db.SetMaxIdleConns(5)                            // Maximum idle connections
	db.SetConnMaxLifetime(time.Hour)                 // Maximum connection lifetime (1 hour)
	db.SetConnMaxIdleTime(5 * time.Minute)           // Maximum idle time before closing (5 minutes)

	return &DB{db}, nil
}
