package datamanager

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- mock adapter for testing ---

type mockAdapter struct {
	data map[string]interface{}
}

func newMockAdapter(data map[string]interface{}) *mockAdapter {
	return &mockAdapter{data: data}
}

func (m *mockAdapter) Connect(f func(e error))     {}
func (m *mockAdapter) Disconnect(f func(e error))  {}
func (m *mockAdapter) Get(name string) interface{} { return m.data[name] }

// --- tests ---

func TestNew(t *testing.T) {
	dm := New()
	assert.NotNil(t, dm)
}

func TestAddAdapter(t *testing.T) {
	t.Parallel()
	dm := New()

	mock := newMockAdapter(map[string]interface{}{"default": "connection"})
	dm.AddAdapter("sql", mock)

	assert.True(t, dm.Has("sql"))
	assert.False(t, dm.Has("nosql"))
}

func TestGetRaw_Success(t *testing.T) {
	t.Parallel()
	dm := New()

	mock := newMockAdapter(map[string]interface{}{"default": "my-conn"})
	dm.AddAdapter("sql", mock)

	val, err := dm.GetRaw("sql", "default")
	assert.NoError(t, err)
	assert.Equal(t, "my-conn", val)
}

func TestGetRaw_AdapterNotFound(t *testing.T) {
	t.Parallel()
	dm := New()

	val, err := dm.GetRaw("nosql", "default")
	assert.Error(t, err)
	assert.Nil(t, val)
	assert.Contains(t, err.Error(), "cannot find adapter")
}

func TestGet_TypeSafe(t *testing.T) {
	t.Parallel()
	dm := New()

	mock := newMockAdapter(map[string]interface{}{"primary": "typed-connection"})
	dm.AddAdapter("sql", mock)

	val, err := Get[string](dm, "sql", "primary")
	assert.NoError(t, err)
	assert.Equal(t, "typed-connection", val)
}

func TestGet_WrongType(t *testing.T) {
	t.Parallel()
	dm := New()

	mock := newMockAdapter(map[string]interface{}{"default": "string-value"})
	dm.AddAdapter("sql", mock)

	// Try to get as int — should return error, NOT panic
	val, err := Get[int](dm, "sql", "default")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "expected int")
	assert.Equal(t, 0, val)
}

func TestGet_AdapterNotFound(t *testing.T) {
	t.Parallel()
	dm := New()

	val, err := Get[string](dm, "nosql", "default")
	assert.Error(t, err)
	assert.Equal(t, "", val)
}

func TestGet_ConnectionNotFound(t *testing.T) {
	t.Parallel()
	dm := New()

	mock := newMockAdapter(map[string]interface{}{})
	dm.AddAdapter("sql", mock)

	val, err := Get[string](dm, "sql", "nonexistent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
	assert.Equal(t, "", val)
}

func TestMustGet_Success(t *testing.T) {
	t.Parallel()
	dm := New()

	mock := newMockAdapter(map[string]interface{}{"cache": 42})
	dm.AddAdapter("kv", mock)

	val := MustGet[int](dm, "kv", "cache")
	assert.Equal(t, 42, val)
}

func TestMustGet_Panics(t *testing.T) {
	t.Parallel()
	dm := New()

	assert.Panics(t, func() {
		MustGet[string](dm, "nosql", "default")
	})
}

func TestConnect_Success(t *testing.T) {
	t.Parallel()
	dm := New()

	mock := newMockAdapter(nil)
	dm.AddAdapter("sql", mock)

	err := dm.Connect("sql", func(e error) {})
	assert.NoError(t, err)
}

func TestConnect_NotFound(t *testing.T) {
	t.Parallel()
	dm := New()

	err := dm.Connect("nosql", func(e error) {})
	assert.Error(t, err)
}

func TestClose_Success(t *testing.T) {
	t.Parallel()
	dm := New()

	mock := newMockAdapter(nil)
	dm.AddAdapter("sql", mock)

	err := dm.Close("sql", func(e error) {})
	assert.NoError(t, err)
}

func TestClose_NotFound(t *testing.T) {
	t.Parallel()
	dm := New()

	err := dm.Close("nosql", func(e error) {})
	assert.Error(t, err)
}

func TestGet_StructType(t *testing.T) {
	t.Parallel()

	type FakeDB struct {
		ConnString string
	}

	dm := New()
	expected := &FakeDB{ConnString: "sqlite:///data.db"}

	mock := newMockAdapter(map[string]interface{}{"default": expected})
	dm.AddAdapter("sql", mock)

	val, err := Get[*FakeDB](dm, "sql", "default")
	assert.NoError(t, err)
	require.NotNil(t, val)
	assert.Equal(t, "sqlite:///data.db", val.ConnString)
}

func TestHas(t *testing.T) {
	t.Parallel()
	dm := New()

	assert.False(t, dm.Has("sql"))

	dm.AddAdapter("sql", newMockAdapter(nil))
	assert.True(t, dm.Has("sql"))
	assert.False(t, dm.Has("nosql"))
}
