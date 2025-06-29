package adapters

import (
	"goodin/infrastructure/persistent"
	"goodin/pkg/config"
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
}
