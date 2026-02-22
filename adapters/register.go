package adapters

import (
	"mox/infrastructure/persistent"
	"mox/pkg/config"
)

type Adapter interface {
	DriverName() string
	Connect(config config.Database) (interface{}, error)
	Close() error
}

var RegisteredSQLAdapters []Adapter = []Adapter{
	&persistent.Postgres{},
	&persistent.MySQL{},
	&persistent.PostgresGorm{},
	&persistent.SQLite{},
	&persistent.SQLiteGorm{},
}
