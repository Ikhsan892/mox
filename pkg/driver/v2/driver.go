package driver

import (
	"errors"
	"fmt"
	"sync"
)

// IDriver defines the contract for a pluggable driver.
// This is intentionally compatible with pkg/driver.IDriver (v1)
// so existing driver implementations can be used with both versions.
type IDriver interface {
	Name() string
	Init() error
	Instance() interface{}
	Close() error
}

// Manager manages the lifecycle of registered drivers.
// Unlike v1 Driver, it provides type-safe access via Get[T]() and MustGet[T]().
type Manager struct {
	drivers []IDriver
	mu      sync.RWMutex
}

// NewManager creates a new driver manager.
func NewManagerV2() *Manager {
	return &Manager{
		drivers: make([]IDriver, 0),
	}
}

// instance returns the raw driver instance by name (internal use).
func (m *Manager) Instance(name string) interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, d := range m.drivers {
		if d.Name() == name {
			return d.Instance()
		}
	}

	return nil
}

// Has returns true if a driver with the given name is registered.
func (m *Manager) Has(name string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, d := range m.drivers {
		if d.Name() == name {
			return true
		}
	}

	return false
}

// RunDriver initializes and registers a driver.
// Returns an error if the driver is nil or initialization fails.
// The driver is only appended after successful Init().
func (m *Manager) RunDriver(d IDriver) error {
	if d == nil {
		return errors.New("driver cannot be nil")
	}

	if err := d.Init(); err != nil {
		return fmt.Errorf("failed to initialize driver %q: %w", d.Name(), err)
	}

	m.mu.Lock()
	m.drivers = append(m.drivers, d)
	m.mu.Unlock()

	return nil
}

// CloseAll gracefully closes all registered drivers.
// Collects and returns all errors encountered during close.
func (m *Manager) CloseAllDriver(handleError func(string, error)) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	var errs []error

	for _, d := range m.drivers {
		if err := d.Close(); err != nil {
			handleError(d.Name(), err)
			errs = append(errs, fmt.Errorf("driver %q: %w", d.Name(), err))
		}
	}

	m.drivers = nil

	return errors.Join(errs...)
}

// Get retrieves a driver instance by name and asserts its type.
// Returns an error if the driver is not found or the type assertion fails.
//
// Usage:
//
//	db, err := driver.Get[*sql.DB](mgr, "default")
func Get[T any](m *Manager, name string) (T, error) {
	raw := m.Instance(name)

	var zero T

	if raw == nil {
		return zero, fmt.Errorf("driver %q not found", name)
	}

	typed, ok := raw.(T)
	if !ok {
		return zero, fmt.Errorf("driver %q is %T, expected %T", name, raw, zero)
	}

	return typed, nil
}

// MustGet retrieves a driver instance by name and asserts its type.
// Panics with a descriptive message if the driver is not found or type mismatches.
// Use this only during application init/wiring where panic is acceptable.
//
// Usage:
//
//	db := driver.MustGet[*sql.DB](mgr, "default")
func MustGet[T any](m *Manager, name string) T {
	val, err := Get[T](m, name)
	if err != nil {
		panic(fmt.Sprintf("driver.MustGet: %s", err))
	}

	return val
}
