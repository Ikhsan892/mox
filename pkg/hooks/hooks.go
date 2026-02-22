package hooks

import (
	"errors"
	"fmt"
	"sort"
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

	for k := range h.handlers {
		if err := h.ExecuteOnly(k, param); err != nil {
			return err
		}
	}

	return nil
}

func (h *Hook[Param]) orderedLoop(data HandlePair[Param]) []string {
	keys := make([]string, 0, len(data))
	for key := range data {
		keys = append(keys, key)
	}

	sort.Strings(keys)

	return keys
}

func (h *Hook[Param]) ExecuteWithExclude(param Param, excludedKeys []string) error {
	if len(h.handlers) < 1 {
		return ErrHandlerNotRegistered
	}

	ex := make(map[string]struct{}, len(excludedKeys))
	for _, k := range excludedKeys {
		ex[k] = struct{}{}
	}

	// Preserve the same iteration order as orderedLoop.
	keys := h.orderedLoop(h.handlers)

	// Filter keys by exclusion set.
	filtered := make([]string, 0, len(keys))
	for _, k := range keys {
		if _, skip := ex[k]; skip {
			continue
		}
		filtered = append(filtered, k)
	}

	// Nothing to execute after exclusions — treat as no-op (or return a custom error if you prefer).
	if len(filtered) == 0 {
		return nil
	}

	// Execute remaining handlers in order.
	for _, k := range filtered {
		if err := h.ExecuteOnly(k, param); err != nil {
			return fmt.Errorf("hook %s: %w", k, err)
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
