package driver

import (
	"errors"
	"sync"
)

type IDriver interface {
	Name() string
	Init() error
	Instance() interface{}
	Close() error
}

type Driver struct {
	drivers []IDriver
	mu      *sync.Mutex
}

func NewDriver() *Driver {
	return &Driver{
		drivers: make([]IDriver, 0),
		mu:      &sync.Mutex{},
	}
}

func (d *Driver) Instance(name string) interface{} {
	if d.drivers == nil {
		return nil
	}

	for _, driver := range d.drivers {
		if driver.Name() == name {
			return driver.Instance()
		}
	}

	return nil

}

func (d *Driver) RunDriver(driver IDriver) error {
	if driver == nil {
		return errors.New("Driver cannot be nil")
	}

	defer func() {
		d.mu.Lock()
		d.drivers = append(d.drivers, driver)
		d.mu.Unlock()
	}()

	return driver.Init()
}

func (d *Driver) CloseAllDriver(handleError func(string, error)) error {
	if len(d.drivers) > 0 {
		for _, driver := range d.drivers {
			if err := driver.Close(); err != nil {
				handleError(driver.Name(), err)
				continue
			}
		}
	}

	d.drivers = nil

	return nil
}
