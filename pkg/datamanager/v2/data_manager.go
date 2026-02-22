package datamanager

import (
	"fmt"
)

// DataAdapter defines the contract for a data source adapter.
// This is intentionally compatible with pkg/datamanager.DataAdapter (v1)
// so existing adapter implementations work with both versions.
type DataAdapter interface {
	Connect(f func(e error))
	Get(connectionName string) interface{}
	Disconnect(f func(e error))
}

// DataManager manages multiple data adapters (SQL, NoSQL, search, etc.).
// Unlike v1, Get() returns errors instead of panicking.
type DataManager struct {
	adapter map[string]DataAdapter
}

// New creates a new DataManager.
func New() *DataManager {
	return &DataManager{
		adapter: make(map[string]DataAdapter),
	}
}

// AddAdapter registers a data adapter by key (e.g., "sql", "nosql", "search").
func (d *DataManager) AddAdapter(key string, adapter DataAdapter) {
	d.adapter[key] = adapter
}

// Has returns true if an adapter with the given key is registered.
func (d *DataManager) Has(dbType string) bool {
	_, ok := d.adapter[dbType]
	return ok
}

// GetRaw retrieves a raw connection by adapter type and connection name.
// Returns an error instead of panicking when the adapter is not found.
func (d *DataManager) GetRaw(dbType, connectionName string) (interface{}, error) {
	adapter, ok := d.adapter[dbType]
	if !ok {
		return nil, fmt.Errorf("cannot find adapter for type %q", dbType)
	}

	return adapter.Get(connectionName), nil
}

// Connect connects a specific adapter type.
// Returns an error instead of panicking when the adapter is not found.
func (d *DataManager) Connect(dbType string, f func(e error)) error {
	adapter, ok := d.adapter[dbType]
	if !ok {
		return fmt.Errorf("cannot find adapter for type %q", dbType)
	}

	adapter.Connect(f)
	return nil
}

// Close disconnects a specific adapter type.
// Returns an error instead of panicking when the adapter is not found.
func (d *DataManager) Close(dbType string, f func(e error)) error {
	adapter, ok := d.adapter[dbType]
	if !ok {
		return fmt.Errorf("cannot find adapter for type %q", dbType)
	}

	adapter.Disconnect(f)
	return nil
}

// Get retrieves a typed connection using generics.
// Returns an error if the adapter is not found or the type assertion fails.
//
// Usage:
//
//	db, err := datamanager.Get[*sql.DB](dm, "sql", "default")
//	gormDB, err := datamanager.Get[*gorm.DB](dm, "sql", "gorm")
func Get[T any](d *DataManager, dbType, connectionName string) (T, error) {
	var zero T

	raw, err := d.GetRaw(dbType, connectionName)
	if err != nil {
		return zero, err
	}

	if raw == nil {
		return zero, fmt.Errorf("connection %q not found in adapter %q", connectionName, dbType)
	}

	typed, ok := raw.(T)
	if !ok {
		return zero, fmt.Errorf("connection %q in adapter %q is %T, expected %T", connectionName, dbType, raw, zero)
	}

	return typed, nil
}

// MustGet retrieves a typed connection using generics.
// Panics with a descriptive message if the adapter or connection is not found,
// or if the type assertion fails.
// Use this only during application init/wiring where panic is acceptable.
//
// Usage:
//
//	db := datamanager.MustGet[*sql.DB](dm, "sql", "default")
func MustGet[T any](d *DataManager, dbType, connectionName string) T {
	val, err := Get[T](d, dbType, connectionName)
	if err != nil {
		panic(fmt.Sprintf("datamanager.MustGet: %s", err))
	}

	return val
}
