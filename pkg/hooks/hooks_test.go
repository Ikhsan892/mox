package hooks_test

import (
	"errors"
	"fmt"
	"testing"

	"mox/pkg/hooks"
	"github.com/stretchr/testify/assert"
)

func TestAddHook(t *testing.T) {
	t.Parallel()

	hook := hooks.New[int64]()

	add := hook.Add("angka", func(e int64) error {

		fmt.Println(e)

		return nil
	})

	assert.Nil(t, add, "Berhasil diatur")
}

func TestAddHookKeyAlreadyExists(t *testing.T) {
	t.Parallel()

	handler := func(e int64) error {
		return nil
	}

	hook := hooks.New[int64]()

	hook.Add("angka", handler)
	err := hook.Add("angka", handler)

	assert.Error(t, err)
}

func TestIsTrueKeyAlreadyExists(t *testing.T) {
	t.Parallel()

	key := "@angka"
	handler := func(e int64) error {
		return nil
	}

	hook := hooks.New[int64]()

	hook.Add(key, handler)
	isSet := hook.IsKeyAlreadySet(key)

	assert.True(t, isSet)
}

func TestIsFalseKeyAlreadyExists(t *testing.T) {
	t.Parallel()

	key := "@angka"

	hook := hooks.New[int64]()

	isSet := hook.IsKeyAlreadySet(key)

	assert.False(t, isSet)
}

func TestExecuteHookByName(t *testing.T) {
	t.Parallel()

	errFired := errors.New("Error fired")

	tableTest := []struct {
		param   int64
		key     string
		handler func(e int64) error
		err     error
		isErr   bool
	}{
		{
			param: int64(10),
			key:   "@angka",
			handler: func(e int64) error {
				return nil
			},
			err:   nil,
			isErr: false,
		},
		{
			param:   int64(10),
			key:     "@not_found",
			handler: nil,
			err:     hooks.ErrKeyNotExists,
			isErr:   true,
		},
		{
			param: int64(10),
			key:   "@angka_with_error",
			handler: func(e int64) error {
				return errFired
			},
			err:   errFired,
			isErr: true,
		},
	}

	for _, v := range tableTest {
		hook := hooks.New[int64]()

		if v.handler != nil {
			hook.Add(v.key, v.handler)
		}

		err := hook.ExecuteOnly(v.key, int64(v.param))
		if v.isErr {
			assert.ErrorIs(t, v.err, err)
		} else {
			assert.Nil(t, err)
		}

	}

}

func TestExecute(t *testing.T) {
	t.Parallel()

	count := 0
	handler := func(e int64) error {
		count++

		return nil
	}

	hook := hooks.New[int64]()

	hook.Add("@angka1", handler)
	hook.Add("@angka2", handler)
	hook.Add("@angka3", handler)
	hook.Add("@angka4", handler)

	err := hook.Execute(int64(10))

	assert.Equal(t, 4, count)
	assert.Nil(t, err)

}

func TestExecuteError(t *testing.T) {
	t.Parallel()

	hook := hooks.New[int64]()

	err := hook.Execute(int64(10))

	assert.Error(t, err, hooks.ErrHandlerNotRegistered)
}

func TestExecuteHandlerError(t *testing.T) {
	t.Parallel()

	errFired := errors.New("error fired")

	handler1 := func(e int64) error {
		return nil
	}

	handler2 := func(e int64) error {
		return errFired
	}

	hook := hooks.New[int64]()

	hook.Add("@angka1", handler1)
	hook.Add("@angka2", handler2)
	err := hook.Execute(int64(10))

	assert.Error(t, err, errFired)
}
