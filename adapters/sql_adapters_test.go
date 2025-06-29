package adapters

import (
	"database/sql"
	"errors"
	"testing"

	adapter_mock "goodin/mocks/goodin/adapters"
	"goodin/pkg/config"
	"github.com/stretchr/testify/assert"
)

func TestNewAdapter(t *testing.T) {
	mock := adapter_mock.NewMockAdapter(t)
	mock.EXPECT().DriverName().Return("test_adapter")

	// config default database
	cfg := config.Config{
		Database: config.Database{
			Adapter: "test_adapter",
		},
	}

	adapter := NewSqlAdapters(&cfg, []Adapter{mock})
	assert.NotNil(t, adapter.getAdapterByDriver("test_adapter"))
}

func TestDefaultConnectionAdapter(t *testing.T) {
	mock := adapter_mock.NewMockAdapter(t)
	f := func(e error) {
		assert.Fail(t, e.Error())
	}

	// config default database
	cfg := config.Config{
		Database: config.Database{
			Adapter: "test_adapter",
			Alias:   "test",
		},
	}

	mock.EXPECT().DriverName().Return(cfg.Database.Adapter)
	mock.EXPECT().Connect(cfg.Database).Return(&sql.DB{}, nil)

	adapter := NewSqlAdapters(&cfg, []Adapter{mock})
	adapter.Connect(f)

	assert.NotNil(t, adapter.config)
	assert.Len(t, adapter.adapters, 1)
	assert.NotNil(t, adapter.alias)
	assert.Contains(t, adapter.alias, "test")
}

func TestMultiConnectionAdapter(t *testing.T) {
	mock := adapter_mock.NewMockAdapter(t)
	mock2 := adapter_mock.NewMockAdapter(t)
	f := func(e error) {
		assert.Fail(t, e.Error())
	}

	// config default database
	cfg := config.Config{
		Database: config.Database{
			Adapter: "test_adapter",
			Alias:   "test",
		},
		ExternalDatabases: []config.Database{
			{
				Adapter: "test_adapter_2",
				Alias:   "test_2",
			},
		},
	}

	mock.EXPECT().DriverName().Return(cfg.Database.Adapter)
	mock.EXPECT().Connect(cfg.Database).Return(&sql.DB{}, nil)

	mock2.EXPECT().DriverName().Return(cfg.ExternalDatabases[0].Adapter)
	mock2.EXPECT().Connect(cfg.ExternalDatabases[0]).Return(&sql.DB{}, nil)

	adapter := NewSqlAdapters(&cfg, []Adapter{mock, mock2})
	adapter.Connect(f)

	assert.NotNil(t, adapter.config)
	assert.Len(t, adapter.adapters, 2)
	assert.NotNil(t, adapter.alias)
	assert.Contains(t, adapter.alias, "test")
	assert.Contains(t, adapter.alias, "test_2")
}

func TestInitConnectionAdapter(t *testing.T) {
	mock := adapter_mock.NewMockAdapter(t)
	f := func(e error) {
		assert.Fail(t, e.Error())
	}

	// config default database
	cfg := config.Config{
		Database: config.Database{
			Adapter: "test_adapter",
			Alias:   "test",
		},
	}

	mock.EXPECT().DriverName().Return("test_adapter")
	mock.EXPECT().Connect(cfg.Database).Return(&sql.DB{}, nil)

	adapter := NewSqlAdapters(&cfg, []Adapter{mock})
	adapter.initConnection(cfg.Database, f)
}

func TestGetAdapter(t *testing.T) {
	mock := adapter_mock.NewMockAdapter(t)

	// config default database
	cfg := config.Config{
		Database: config.Database{
			Adapter: "test_adapter",
			Alias:   "test",
		},
	}

	mock.EXPECT().DriverName().Return("test_adapter")

	adapter := NewSqlAdapters(&cfg, []Adapter{mock})

	assert.NotNil(t, adapter.getAdapterByDriver(cfg.Database.Adapter))
}

func TestGetAdapterNotFound(t *testing.T) {
	mock := adapter_mock.NewMockAdapter(t)

	// config default database
	cfg := config.Config{
		Database: config.Database{
			Adapter: "test_adapter",
			Alias:   "test",
		},
	}

	mock.EXPECT().DriverName().Return("test_adapter_2")

	adapter := NewSqlAdapters(&cfg, []Adapter{mock})

	assert.Nil(t, adapter.getAdapterByDriver(cfg.Database.Adapter))
}

func TestInitErrorConnectionAdapter(t *testing.T) {
	mock := adapter_mock.NewMockAdapter(t)
	f := func(e error) {

	}

	defer func() {
		if err := recover(); err == nil {
			assert.Fail(t, "Expected to panic in log handler")
		}
	}()

	// config default database
	cfg := config.Config{
		Database: config.Database{
			Adapter: "test_adapter",
			Alias:   "test",
		},
	}

	mock.EXPECT().DriverName().Return("test_adapter")
	mock.EXPECT().Connect(cfg.Database).Return(nil, errors.New("error"))

	adapter := NewSqlAdapters(&cfg, []Adapter{mock})
	adapter.initConnection(cfg.Database, f)
}

func TestInitErrorConnectionAdapterNotFound(t *testing.T) {
	mock := adapter_mock.NewMockAdapter(t)
	f := func(e error) {

	}

	// config default database
	cfg := config.Config{
		Database: config.Database{
			Adapter: "test_adapter",
			Alias:   "test",
		},
	}

	mock.EXPECT().DriverName().Return("test_adapter_2")

	adapter := NewSqlAdapters(&cfg, []Adapter{mock})

	defer func() {
		if err := recover(); err == nil {
			assert.Fail(t, "Expected to panic in log handler")
		}
	}()

	adapter.initConnection(cfg.Database, f)
}

func TestGetConnection(t *testing.T) {
	mock := adapter_mock.NewMockAdapter(t)
	f := func(e error) {
		assert.Fail(t, e.Error())
	}

	// config default database
	cfg := config.Config{
		Database: config.Database{
			Adapter: "test_adapter",
			Alias:   "test",
		},
	}

	mock.EXPECT().DriverName().Return(cfg.Database.Adapter)
	mock.EXPECT().Connect(cfg.Database).Return(&sql.DB{}, nil)

	adapter := NewSqlAdapters(&cfg, []Adapter{mock})
	adapter.Connect(f)

	assert.NotNil(t, adapter.Get("test"))
}

func TestGetConnectionNotFound(t *testing.T) {
	mock := adapter_mock.NewMockAdapter(t)
	f := func(e error) {
		assert.Fail(t, e.Error())
	}

	// config default database
	cfg := config.Config{
		Database: config.Database{
			Adapter: "test_adapter",
			Alias:   "test",
		},
	}

	mock.EXPECT().DriverName().Return(cfg.Database.Adapter)
	mock.EXPECT().Connect(cfg.Database).Return(&sql.DB{}, nil)

	adapter := NewSqlAdapters(&cfg, []Adapter{mock})
	adapter.Connect(f)

	assert.Nil(t, adapter.Get("test_2"))
}

func TestDisconnectConnection(t *testing.T) {
	mock := adapter_mock.NewMockAdapter(t)
	mock2 := adapter_mock.NewMockAdapter(t)
	f := func(e error) {
		if e != nil {
			assert.ErrorContains(t, e, e.Error())
		}
	}

	// config default database
	cfg := config.Config{
		Database: config.Database{
			Adapter: "test_adapter",
			Alias:   "test",
		},
		ExternalDatabases: []config.Database{
			{
				Adapter: "test_adapter_2",
				Alias:   "test_2",
			},
		},
	}

	mock.EXPECT().Close().Return(nil)
	mock.EXPECT().DriverName().Return("test_adapter")
	mock.EXPECT().Connect(cfg.Database).Return(&sql.DB{}, nil)

	mock2.EXPECT().Close().Return(errors.New(""))
	mock2.EXPECT().DriverName().Return("test_adapter_2")
	mock2.EXPECT().Connect(cfg.ExternalDatabases[0]).Return(&sql.DB{}, nil)

	adapter := NewSqlAdapters(&cfg, []Adapter{mock, mock2})
	adapter.Connect(f)
	adapter.Disconnect(f)

}
