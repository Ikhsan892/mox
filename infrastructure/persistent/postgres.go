package persistent

import (
	"context"
	"database/sql"
	"fmt"

	"mox/pkg/config"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
)

type Postgres struct {
	db *sql.DB
}

func (p *Postgres) DriverName() string {
	return "postgres-sql/db"
}

func (p *Postgres) Connect(config config.Database) (interface{}, error) {
	sslmode := "disable"
	if config.SSLMode {
		sslmode = "enable"
	}

	dsn := fmt.Sprintf("user=%s password=%s host=%s port=%d database=%s sslmode=%s", config.Username, config.Password, config.Host, config.Port, config.DatabaseName, sslmode)

	if config.Schema != "" {
		dsn = fmt.Sprintf("%s search_path=%s", dsn, config.Schema)
	}

	pool, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		return nil, err
	}

	if err := pool.Ping(context.Background()); err != nil {
		return nil, err
	}

	p.db = stdlib.OpenDBFromPool(pool)

	return p.db, nil
}

func (p *Postgres) Close() error {
	return p.db.Close()
}
