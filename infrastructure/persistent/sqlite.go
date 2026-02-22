package persistent

import (
	"database/sql"
	"fmt"

	"mox/pkg/config"

	_ "modernc.org/sqlite"
)

type SQLite struct {
	db *sql.DB
}

func (s *SQLite) DriverName() string {
	return "sqlite"
}

func (s *SQLite) Connect(cfg config.Database) (interface{}, error) {
	dsn := cfg.Path
	if dsn == "" {
		dsn = cfg.DatabaseName // fallback: use database_name as file path
	}

	if dsn == "" {
		return nil, fmt.Errorf("sqlite requires a file path via 'path' or 'database_name' config field")
	}

	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open sqlite database: %w", err)
	}

	// Enable WAL mode for better concurrent read performance
	if _, err := db.Exec("PRAGMA journal_mode=WAL"); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to set WAL mode: %w", err)
	}

	// Enable foreign key constraints (disabled by default in SQLite)
	if _, err := db.Exec("PRAGMA foreign_keys=ON"); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to enable foreign keys: %w", err)
	}

	// Set busy timeout to 5 seconds to avoid SQLITE_BUSY errors
	if _, err := db.Exec("PRAGMA busy_timeout=5000"); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to set busy timeout: %w", err)
	}

	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping sqlite database: %w", err)
	}

	// SQLite performs best with limited connections due to file-level locking
	pool := cfg.Pool
	if pool <= 0 {
		pool = 1
	}
	db.SetMaxOpenConns(pool)
	db.SetMaxIdleConns(pool)

	s.db = db

	return s.db, nil
}

func (s *SQLite) Close() error {
	if s.db != nil {
		return s.db.Close()
	}
	return nil
}
