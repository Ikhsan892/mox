package driver

import (
	"errors"
	"goodin/mocks/goodin/pkg/driver"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewDriver(t *testing.T) {
	drv := NewDriver()

	assert.NotNil(t, drv.drivers)
	assert.NotNil(t, drv.mu)
}

func TestRunDriver(t *testing.T) {
	t.Parallel()
	drv := NewDriver()

	mock := driver.NewMockIDriver(t)
	mock.EXPECT().Init().Return(nil)

	drv.RunDriver(mock)

	assert.Equal(t, len(drv.drivers), 1)
	mock.AssertCalled(t, "Init")
}

func TestGetInstanceDriver(t *testing.T) {
	t.Parallel()
	drv := NewDriver()

	mock := driver.NewMockIDriver(t)
	mock.EXPECT().Name().Return("mock-driver")
	mock.EXPECT().Init().Return(nil)
	mock.EXPECT().Instance().Return(mock)

	err := drv.RunDriver(mock)

	instance := drv.Instance("mock-driver")
	assert.NotNil(t, instance)

	assert.Nil(t, err)
}

func TestGetInstanceNilDriver(t *testing.T) {
	t.Parallel()
	drv := &Driver{}

	instance := drv.Instance("mock-driver")
	assert.Nil(t, instance)
}

func TestRunDriverEmpty(t *testing.T) {
	t.Parallel()
	drv := NewDriver()

	err := drv.RunDriver(nil)

	assert.NotNil(t, err)
	assert.Error(t, err)
}

func TestCloseAllDrivers(t *testing.T) {
	t.Parallel()
	mock := driver.NewMockIDriver(t)
	// mock.EXPECT().Name().Return("Test")
	mock.EXPECT().Close().Return(nil)
	mock.EXPECT().Init().Return(nil)

	drv := NewDriver()

	err := drv.RunDriver(mock)
	if err != nil {
		assert.Fail(t, "Should not failing while running")
	}

	assert.Equal(t, len(drv.drivers), 1)

	drv.CloseAllDriver(func(s string, err error) {
		assert.Nil(t, err)
	})

	assert.Nil(t, drv.drivers)
}

func TestCloseAllDriversWithError(t *testing.T) {
	t.Parallel()
	mock := driver.NewMockIDriver(t)
	mock.EXPECT().Name().Return("Test")
	mock.EXPECT().Close().Return(errors.New("Error Closing Driver"))
	mock.EXPECT().Init().Return(nil)

	drv := NewDriver()

	err := drv.RunDriver(mock)
	if err != nil {
		assert.Fail(t, "Should not failing while running")
	}

	assert.Equal(t, len(drv.drivers), 1)

	drv.CloseAllDriver(func(s string, err error) {
		assert.NotNil(t, err)
		assert.Error(t, err)
		assert.Equal(t, s, "Test")
	})

	assert.Nil(t, drv.drivers)
}
