package workerclient

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"sync"
	"time"

	core "mox/internal"
	"mox/tools/utils"
	"mox/use_cases/operation"
)

type WorkerProcess interface {
	PID() int
	State() WorkerClientState
	IsAlive() bool
	Start() error
	Drain() error
	Shutdown() error

	Send(ctx context.Context, msg operation.MessagePayload) (int, error)
}

var _ (WorkerProcess) = (*WorkerClient)(nil)

type WorkerClient struct {
	status WorkerClientState
	pid    int
	app    core.App
	mu     *sync.Mutex
	l      *net.UnixConn
}

func NewWorkerClient(
	app core.App,
	l *net.UnixConn,
	pid int,
) *WorkerClient {
	return &WorkerClient{
		app:    app,
		status: Connecting,
		pid:    pid,
		mu:     &sync.Mutex{},
		l:      l,
	}
}

// Drain implements [WorkerProcess].
func (w *WorkerClient) Drain() error {
	_, err := w.Send(w.app.Context(), operation.MessagePayload{
		ID:      utils.GenerateUUID(),
		FromPID: w.PID(),
		Payload: operation.Command{
			Type: operation.Drain,
		},
		Timestamp: time.Now().UnixMilli(),
	})
	if err != nil {
		return err
	}

	w.app.Logger().Info("worker client already sent drain command")

	return nil
}

// IsAlive implements [WorkerProcess].
func (w *WorkerClient) IsAlive() bool {
	return w.status == Connected
}

// PID implements [WorkerProcess].
func (w *WorkerClient) PID() int {
	return w.pid
}

// Send implements [WorkerProcess].
func (w *WorkerClient) Send(ctx context.Context, msg operation.MessagePayload) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.l == nil {
		return 0, fmt.Errorf("worker %d has no active connection", w.pid)
	}

	b, err := json.Marshal(msg)
	if err != nil {
		return 0, fmt.Errorf("failed to marshal message: %w", err)
	}

	b = append(b, '\n')

	n, err := w.l.Write(b)
	if err != nil {
		return 0, fmt.Errorf("failed to send message to worker %d: %w", w.pid, err)
	}

	w.app.Logger().Info(fmt.Sprintf("send to tcp packet from pid %d and total byte %d", w.pid, n))

	return n, nil
}

// Shutdown implements [WorkerProcess].
func (w *WorkerClient) Shutdown() error {
	w.status = Disconnected

	if w.l == nil {
		w.app.Logger().Warn(fmt.Sprintf("listener for pid %d is nil", w.pid))
		return nil
	}

	// try to shutdown worker if possible
	go func() {
		_, err := w.Send(w.app.Context(), operation.MessagePayload{
			ID:      utils.GenerateUUID(),
			FromPID: -1,
			Payload: operation.Command{
				Type: operation.Shutdown,
			},
			Timestamp: time.Now().UnixMilli(),
		})
		if err != nil {
			w.app.Logger().Warn(err.Error())
		}
		w.app.Logger().Info("send message shutdown to workers")
	}()

	return nil
}

func (w *WorkerClient) Start() error {
	w.status = Connected

	_, err := w.Send(w.app.Context(), operation.MessagePayload{
		ID:      utils.GenerateUUID(),
		FromPID: w.PID(),
		Payload: operation.Command{
			Type: operation.Ping,
		},
		Timestamp: time.Now().UnixMilli(),
	})
	if err != nil {
		return err
	}

	return nil
}

func (w *WorkerClient) State() WorkerClientState {
	return w.status
}
