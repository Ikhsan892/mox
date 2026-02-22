package operation

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
)

var _ (IControl) = (*MasterRegistry)(nil)

// 4. The Registry (Thread-Safe)
type MasterRegistry struct {
	mu       sync.RWMutex
	commands map[string]MasterControlHandler
}

// Constructor
func NewMasterRegistry() *MasterRegistry {
	return &MasterRegistry{
		commands: make(map[string]MasterControlHandler),
	}
}

// Register Daftarin command baru
func (r *MasterRegistry) Register(name string, desc string, usage string, handler MasterControlHandler) {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Normalize ke uppercase biar case-insensitive (STOP == stop)
	r.commands[strings.ToUpper(name)] = handler
}

// Execute: Routing dari raw string telnet ke function
func (r *MasterRegistry) Execute(ctx context.Context, syscore SystemCore, cmd Command) (string, error) {
	r.mu.RLock()
	cmdFn, exists := r.commands[cmd.Name]
	r.mu.RUnlock()

	if !exists {
		return fmt.Sprintf("Unknown command: %s. Type HELP for list.", cmd.Name), errors.New(fmt.Sprintf("Unknown command: %s. Type HELP for list.", cmd.Name))
	}

	// 3. Eksekusi Handler
	if err := cmdFn(ctx, syscore, cmd); err != nil {
		return "", err
	}

	return "", nil
}
