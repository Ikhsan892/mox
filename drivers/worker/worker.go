package worker

import (
	"log/slog"
	"net"
	"os"

	core "mox/internal"
	"mox/pkg/driver"
	"mox/use_cases/workercore"
)

var _ (driver.IDriver) = (*WorkerAdapter)(nil)

const WorkerAdapterName = "MasterAdapter"

type WorkerAdapter struct {
	app        core.App
	workercore *workercore.Worker
	conn       *net.UnixConn
}

func NewWorkerAdapter(app core.App) *WorkerAdapter {
	return &WorkerAdapter{app: app}
}

// Close implements [driver.IDriver].
func (w *WorkerAdapter) Close() error {
	if w.conn != nil {
		return w.conn.Close()
	}

	return nil
}

// Init implements [driver.IDriver].
func (w *WorkerAdapter) Init() error {
	pid := os.Getpid()

	ctx := w.app.Context()

	socketPath := "/tmp/http_mgr.sock"

	conn, err := net.DialUnix("unix", nil, &net.UnixAddr{Name: socketPath, Net: "unix"})
	if err != nil {
		w.app.Logger().Error("worker cannot run the listener", slog.String("err", err.Error()))
		return err
	}

	worker := workercore.NewWorkerBuilder().
		SetListener(conn).
		SetPID(pid).
		Build()

	if err := worker.AcceptHandshake(); err != nil {
		return err
	}

	w.workercore = worker
	w.conn = conn

	go worker.ReceiveMessage(ctx, func() {
		w.app.Stop()
	})

	return nil
}

// Instance implements [driver.IDriver].
func (w *WorkerAdapter) Instance() interface{} {
	return w.workercore
}

// Name implements [driver.IDriver].
func (w *WorkerAdapter) Name() string {
	return WorkerAdapterName
}
