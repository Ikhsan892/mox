package hooks

import (
	"errors"
	"sync"
)

var (
	ErrKeyAlreadyExists     error = errors.New("key already exists")
	ErrKeyNotExists         error = errors.New("key not exists")
	ErrHandlerNotRegistered error = errors.New("handler not registered")
)

type Handler[Param any] func(e Param) error

// For execute bv key
type HandlePair[Param any] map[string]Handler[Param]

type Hook[Param any] struct {
	mu       sync.RWMutex
	handlers HandlePair[Param]
}

func New[T any]() *Hook[T] {
	return &Hook[T]{
		handlers: make(HandlePair[T]),
	}
}

func (h *Hook[Param]) IsKeyAlreadySet(key string) bool {
	_, ok := h.handlers[key]
	if ok {
		return true
	}

	return false
}

func (h *Hook[Param]) Add(key string, handler Handler[Param]) error {
	h.mu.Lock()

	defer h.mu.Unlock()

	if h.IsKeyAlreadySet(key) {
		return ErrKeyAlreadyExists
	}

	h.handlers[key] = handler

	return nil
}

// Execute all hooks
func (h *Hook[Param]) Execute(param Param) error {
	if len(h.handlers) < 1 {
		return ErrHandlerNotRegistered
	}

	for k, _ := range h.handlers {
		if err := h.ExecuteOnly(k, param); err != nil {
			return err
		}
	}

	return nil
}

func (h *Hook[Param]) ExecuteOnly(key string, param Param) error {
	if !h.IsKeyAlreadySet(key) {
		return ErrKeyNotExists
	}

	if err := h.handlers[key](param); err != nil {
		return err
	}

	return nil

}
