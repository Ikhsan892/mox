package persistent

import (
	"database/sql"
	"fmt"

	"mox/pkg/config"

	gormsqlite "gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/plugin/opentelemetry/tracing"
)

var (
	sqlitegormdriver = "sqlite-gorm"
)

type SQLiteGorm struct {
	db   *sql.DB
	gorm *gorm.DB
}

func (s *SQLiteGorm) DriverName() string {
	return sqlitegormdriver
}

func (s *SQLiteGorm) Connect(cfg config.Database) (interface{}, error) {
	// Reuse the raw SQLite adapter for the underlying connection
	raw := SQLite{}
	conn, err := raw.Connect(cfg)
	if err != nil {
		return nil, err
	}

	s.db = conn.(*sql.DB)

	gormDB, err := gorm.Open(gormsqlite.Dialector{
		Conn: s.db,
	}, &gorm.Config{})
	if err != nil {
		s.db.Close()
		return nil, fmt.Errorf("failed to open gorm sqlite connection: %w", err)
	}

	if err := gormDB.Use(tracing.NewPlugin(tracing.WithoutMetrics())); err != nil {
		s.db.Close()
		return nil, fmt.Errorf("failed to register otel tracing plugin: %w", err)
	}

	s.gorm = gormDB

	return gormDB, nil
}

func (s *SQLiteGorm) Close() error {
	if s.gorm != nil {
		db, err := s.gorm.DB()
		if err != nil {
			return err
		}
		return db.Close()
	}

	if s.db != nil {
		return s.db.Close()
	}

	return nil
}
