package datamanager_test

import (
	"testing"

	datamanager_mock "goodin/mocks/goodin/pkg/datamanager"
	"goodin/pkg/datamanager"
	"github.com/stretchr/testify/assert"
	mock_obj "github.com/stretchr/testify/mock"
)

func TestNewDataManager(t *testing.T) {
	t.Parallel()

	d := datamanager.New()

	assert.NotNil(t, d)
}

func TestAddAdapter(t *testing.T) {
	t.Parallel()

	mock := datamanager_mock.NewMockDataAdapter(t)
	mock.EXPECT().Get("conn_test").Return("test")

	d := datamanager.New()

	defer func() {
		if err := recover(); err != nil {
			assert.Fail(t, "Should not be panicked")
		}
	}()

	d.AddAdapter("sql", mock)
	assert.NotNil(t, d.Get("sql", "conn_test"))
	assert.Equal(t, "test", d.Get("sql", "conn_test"))
}

func TestGetAdapterNotFound(t *testing.T) {
	t.Parallel()

	mock := datamanager_mock.NewMockDataAdapter(t)

	d := datamanager.New()

	defer func() {
		if err := recover(); err == nil {
			assert.Fail(t, "Should be panicked")
		}
	}()

	d.AddAdapter("sql", mock)
	d.Get("nosql", "conn_test")
}

func TestConnectAdapter(t *testing.T) {
	t.Parallel()

	mock := datamanager_mock.NewMockDataAdapter(t)
	mock.EXPECT().Connect(mock_obj.AnythingOfType("func(error)")).Return()

	f := func(e error) {}

	d := datamanager.New()

	defer func() {
		if err := recover(); err != nil {
			assert.Fail(t, "Should not be panicked")
		}
	}()

	d.AddAdapter("sql", mock)
	d.Connect("sql", f)

}

func TestConnectAdapterFail(t *testing.T) {
	t.Parallel()

	f := func(e error) {}

	d := datamanager.New()

	defer func() {
		if err := recover(); err == nil {
			assert.Fail(t, "Should be panicked")
		}
	}()

	d.Connect("no_sql", f)

}

func TestCloseAdapter(t *testing.T) {
	t.Parallel()

	mock := datamanager_mock.NewMockDataAdapter(t)
	mock.EXPECT().Disconnect(mock_obj.AnythingOfType("func(error)")).Return()

	f := func(e error) {}

	d := datamanager.New()

	defer func() {
		if err := recover(); err != nil {
			assert.Fail(t, "Should not be panicked")
		}
	}()

	d.AddAdapter("sql", mock)
	d.Close("sql", f)

}

func TestCloseAdapterFail(t *testing.T) {
	t.Parallel()

	f := func(e error) {}

	d := datamanager.New()

	defer func() {
		if err := recover(); err == nil {
			assert.Fail(t, "Should be panicked")
		}
	}()

	d.Close("no_sql", f)

}
