package persistent

import (
	"database/sql"

	"mox/pkg/config"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/plugin/opentelemetry/tracing"
)

var (
	postgresgorm       = "gorm"
	postgresgormdriver = "postgres-gorm"
)

type PostgresGorm struct {
	db   *sql.DB
	gorm *gorm.DB
}

func (p *PostgresGorm) DriverName() string {
	return postgresgormdriver
}

func (p *PostgresGorm) Connect(config config.Database) (interface{}, error) {
	pgxpool := Postgres{}

	db, err := pgxpool.Connect(config)
	if err != nil {
		return nil, err
	}

	p.db = db.(*sql.DB)

	gormDB, err := gorm.Open(postgres.New(postgres.Config{
		Conn: p.db,
	}), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	if err := gormDB.Use(tracing.NewPlugin(tracing.WithoutMetrics())); err != nil {
		return nil, err
	}

	p.gorm = gormDB

	return gormDB, nil
}

func (p *PostgresGorm) Close() error {
	db, err := p.gorm.DB()
	if err != nil {
		db.Close()
		return err
	}

	p.db.Close()

	return db.Close()
}
