package adapters

import (
	"fmt"

	"mox/pkg/config"
)

type SqlAdapters struct {
	adapters []Adapter
	alias    map[string]Adapter
	result   map[string]interface{}
	config   *config.Config
}

func NewSqlAdapters(cfg *config.Config, adapters []Adapter) *SqlAdapters {
	return &SqlAdapters{
		adapters: adapters,
		alias:    make(map[string]Adapter, len(cfg.ExternalDatabases)+1),
		result:   make(map[string]interface{}, len(cfg.ExternalDatabases)+1),
		config:   cfg,
	}
}

func (a *SqlAdapters) getAdapterByDriver(driver string) Adapter {
	for _, adapter := range a.adapters {
		if adapter.DriverName() == driver {
			return adapter
		}
	}

	return nil
}

func (a *SqlAdapters) initConnection(cfg config.Database, f func(e error)) interface{} {
	adapter := a.getAdapterByDriver(cfg.Adapter)
	if adapter == nil {
		f(fmt.Errorf("cannot find adapter for %s", cfg.Adapter))
		panic("")
	}

	db, err := adapter.Connect(cfg)
	if err != nil {
		f(fmt.Errorf("cannot connect %s database detail=%s", cfg.Alias, err.Error()))
		panic("")
	}

	return db
}

func (a *SqlAdapters) Get(connectionName string) interface{} {
	return a.result[connectionName]
}

func (a *SqlAdapters) Connect(f func(e error)) {

	// default db
	df := a.initConnection(a.config.Database, f)
	a.result[a.config.Database.Alias] = df
	a.alias[a.config.Database.Alias] = a.getAdapterByDriver(a.config.Database.Adapter)

	// external database
	if len(a.config.ExternalDatabases) > 0 {
		for _, external := range a.config.ExternalDatabases {
			db := a.initConnection(external, f)
			a.result[external.Alias] = db
			a.alias[external.Alias] = a.getAdapterByDriver(external.Adapter)
		}
	}

}

func (a *SqlAdapters) Disconnect(f func(e error)) {
	if a.alias != nil {
		for _, adapter := range a.alias {
			if err := adapter.Close(); err != nil {
				f(fmt.Errorf("cannot close database %s detail=%s", adapter.DriverName(), err.Error()))
				continue
			}
		}
	}
}
