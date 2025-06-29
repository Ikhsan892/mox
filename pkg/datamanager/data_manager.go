package datamanager

import (
	"fmt"
)

type DataAdapter interface {
	Connect(f func(e error))
	Get(connectionName string) interface{}
	Disconnect(f func(e error))
}

type DataManager struct {
	adapter map[string]DataAdapter
}

func New() *DataManager {
	return &DataManager{
		adapter: make(map[string]DataAdapter),
	}
}

func (d *DataManager) AddAdapter(key string, adapter DataAdapter) {
	d.adapter[key] = adapter
}

func (d *DataManager) Get(dbType, connectionName string) interface{} {
	adapter, ok := d.adapter[dbType]
	if !ok {
		panic(fmt.Sprintf("cannot find adapter for type %s", dbType))
	}

	return adapter.Get(connectionName)
}

func (d *DataManager) Connect(dbType string, f func(e error)) {
	adapter, ok := d.adapter[dbType]
	if !ok {
		panic(fmt.Sprintf("cannot find adapter for type %s", dbType))
	}

	adapter.Connect(f)
}

func (d *DataManager) Close(dbType string, f func(e error)) {
	adapter, ok := d.adapter[dbType]
	if !ok {
		panic(fmt.Sprintf("cannot find adapter for type %s", dbType))
	}

	adapter.Disconnect(f)
}
