package driver

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- mock driver for testing ---

type mockDriver struct {
	name     string
	instance interface{}
	initErr  error
	closeErr error
}

func (m *mockDriver) Name() string          { return m.name }
func (m *mockDriver) Init() error           { return m.initErr }
func (m *mockDriver) Instance() interface{} { return m.instance }
func (m *mockDriver) Close() error          { return m.closeErr }

// --- tests ---

func TestNewManager(t *testing.T) {
	mgr := NewManager()
	assert.NotNil(t, mgr)
}

func TestRunDriver(t *testing.T) {
	t.Parallel()
	mgr := NewManager()

	mock := &mockDriver{name: "test", instance: "hello"}
	err := mgr.RunDriver(mock)

	assert.NoError(t, err)
	assert.True(t, mgr.Has("test"))
}

func TestRunDriver_Nil(t *testing.T) {
	t.Parallel()
	mgr := NewManager()

	err := mgr.RunDriver(nil)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot be nil")
}

func TestRunDriver_InitError(t *testing.T) {
	t.Parallel()
	mgr := NewManager()

	mock := &mockDriver{
		name:    "failing",
		initErr: errors.New("init failed"),
	}
	err := mgr.RunDriver(mock)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "init failed")
	assert.False(t, mgr.Has("failing"), "driver should not be registered on init failure")
}

func TestGet_TypeSafe(t *testing.T) {
	t.Parallel()
	mgr := NewManager()

	mock := &mockDriver{name: "db", instance: "my-connection-string"}
	require.NoError(t, mgr.RunDriver(mock))

	val, err := Get[string](mgr, "db")
	assert.NoError(t, err)
	assert.Equal(t, "my-connection-string", val)
}

func TestGet_WrongType(t *testing.T) {
	t.Parallel()
	mgr := NewManager()

	mock := &mockDriver{name: "db", instance: "string-value"}
	require.NoError(t, mgr.RunDriver(mock))

	// Try to get as int — should return error, NOT panic
	val, err := Get[int](mgr, "db")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "expected int")
	assert.Equal(t, 0, val) // zero value
}

func TestGet_NotFound(t *testing.T) {
	t.Parallel()
	mgr := NewManager()

	val, err := Get[string](mgr, "nonexistent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
	assert.Equal(t, "", val) // zero value
}

func TestMustGet_Success(t *testing.T) {
	t.Parallel()
	mgr := NewManager()

	mock := &mockDriver{name: "cache", instance: 42}
	require.NoError(t, mgr.RunDriver(mock))

	val := MustGet[int](mgr, "cache")
	assert.Equal(t, 42, val)
}

func TestMustGet_Panics(t *testing.T) {
	t.Parallel()
	mgr := NewManager()

	assert.Panics(t, func() {
		MustGet[string](mgr, "nonexistent")
	})
}

func TestHas(t *testing.T) {
	t.Parallel()
	mgr := NewManager()

	assert.False(t, mgr.Has("nope"))

	mock := &mockDriver{name: "found", instance: true}
	require.NoError(t, mgr.RunDriver(mock))

	assert.True(t, mgr.Has("found"))
	assert.False(t, mgr.Has("still-nope"))
}

func TestCloseAll(t *testing.T) {
	t.Parallel()
	mgr := NewManager()

	m1 := &mockDriver{name: "a", instance: nil}
	m2 := &mockDriver{name: "b", instance: nil}
	require.NoError(t, mgr.RunDriver(m1))
	require.NoError(t, mgr.RunDriver(m2))

	err := mgr.CloseAll()
	assert.NoError(t, err)
	assert.False(t, mgr.Has("a"))
	assert.False(t, mgr.Has("b"))
}

func TestCloseAll_WithErrors(t *testing.T) {
	t.Parallel()
	mgr := NewManager()

	m1 := &mockDriver{name: "ok", instance: nil}
	m2 := &mockDriver{name: "fail", instance: nil, closeErr: errors.New("close failed")}
	require.NoError(t, mgr.RunDriver(m1))
	require.NoError(t, mgr.RunDriver(m2))

	err := mgr.CloseAll()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "close failed")
}

func TestGet_StructType(t *testing.T) {
	t.Parallel()

	type MyDB struct {
		Name string
	}

	mgr := NewManager()
	expected := &MyDB{Name: "primary"}

	mock := &mockDriver{name: "mydb", instance: expected}
	require.NoError(t, mgr.RunDriver(mock))

	val, err := Get[*MyDB](mgr, "mydb")
	assert.NoError(t, err)
	assert.Equal(t, "primary", val.Name)
}
