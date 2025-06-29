package persistent

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"goodin/pkg/config"
)

type MySQL struct {
	db *sql.DB
}

// Close implements SqlRegister.
func (m *MySQL) Close() error {
	return m.db.Close()
}

// Connect implements SqlRegister.
func (m *MySQL) Connect(config config.Database) (interface{}, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s", config.Username, config.Password, config.Host, config.Port, config.DatabaseName)

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}

	db.SetConnMaxLifetime(time.Minute * 3)
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(10)

	if err := db.Ping(); err != nil {
		return nil, err
	}

	m.db = db

	return m.db, nil
}

// DriverName implements SqlRegister.
func (m *MySQL) DriverName() string {
	return "mysql"
}
