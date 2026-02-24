package mastercore

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sync"
	"time"

	core "mox/internal"
	"mox/tools/utils"
	"mox/use_cases/bus"
	"mox/use_cases/operation"
	"mox/use_cases/workerclient"
)

type Registry interface {
	workerclient.WorkerProvider
	workerclient.WorkerRegistrar
}

var _ (Registry) = (*ConnectionRegistry)(nil)

type ConnectionRegistry struct {
	conns map[int]workerclient.WorkerProcess
	bus   bus.Messaging
	app   core.App
	ctx   context.Context
	mu    *sync.RWMutex
}

func NewConnectionRegistry(ctx context.Context, app core.App) *ConnectionRegistry {
	return &ConnectionRegistry{
		mu:    &sync.RWMutex{},
		ctx:   ctx,
		app:   app,
		bus:   bus.NewEventBus(app),
		conns: make(map[int]workerclient.WorkerProcess),
	}
}

// Remove implements [Registry].
func (c *ConnectionRegistry) Remove(pid int) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.conns, pid)

	return errors.New("no worker found to remove")
}

// Get implements [workerclient.WorkerProvider].
func (c *ConnectionRegistry) Get(pid int) workerclient.WorkerProcess {
	c.mu.RLock()
	defer c.mu.RUnlock()
	w, ok := c.conns[pid]
	if !ok {
		return nil
	}

	return w
}

// GetAll implements [workerclient.WorkerProvider].
func (c *ConnectionRegistry) GetAll() []workerclient.WorkerProcess {
	c.mu.RLock()
	defer c.mu.RUnlock()

	workers := make([]workerclient.WorkerProcess, 0, len(c.conns))
	for _, w := range c.conns {
		workers = append(workers, w)
	}
	return workers
}

func (c *ConnectionRegistry) Total() int64 {
	c.mu.RLock() // Gunakan Read Lock agar efisien
	defer c.mu.RUnlock()
	return int64(len(c.conns))
}

func (c *ConnectionRegistry) Broadcast(payload []byte) {
	c.bus.Broadcast(c.ctx, operation.MessagePayload{
		ID: utils.GenerateUUID(),
		Payload: operation.Command{
			Type:    operation.Chat,
			Payload: payload,
		},
		Timestamp: time.Now().UnixMilli(),
	}, c.GetAll())
}

func (c *ConnectionRegistry) pingWorkers() {
	if err := c.bus.Broadcast(c.ctx, operation.MessagePayload{
		ID:      utils.GenerateUUID(),
		FromPID: -1,
		Payload: operation.Command{
			Type: operation.Ping,
		},
		Timestamp: time.Now().UnixMilli(),
	}, c.GetAll()); err != nil {
		c.app.Logger().Error(fmt.Sprintf("error sending heart msg : %s", err.Error()))
		return
	}
}

func (c *ConnectionRegistry) eliminateWorkers() {
	for pid, worker := range c.conns {
		if worker.State() == workerclient.Disconnected {
			if err := c.Remove(pid); err != nil {
				c.app.Logger().Error(err.Error())
			}

			c.app.Logger().Info(fmt.Sprintf("worker %d removed", pid))
		}
	}
}

func (c *ConnectionRegistry) CheckHealthWorkers() {
	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-c.ctx.Done():
			return
		case <-ticker.C:
			c.pingWorkers()
			c.eliminateWorkers()
		}
	}
}

func (c *ConnectionRegistry) CloseAllConnections() {
	c.app.Logger().Debug("closing", slog.Int("total", int(c.Total())))
	for pid, v := range c.conns {
		c.app.Logger().Info("closing connection", slog.Int("pid", pid))

		if err := v.Shutdown(); err != nil {
			c.app.Logger().Error(fmt.Sprintf("error closing PID: %d, msg : %s", pid, err.Error()))
			continue
		}
	}
}

func (c *ConnectionRegistry) Add(worker workerclient.WorkerProcess) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.conns[worker.PID()] = worker

	c.app.Logger().Debug("Worker added to map", slog.Int("pid", worker.PID()), slog.Int("current_total", len(c.conns)))
	return nil
}
