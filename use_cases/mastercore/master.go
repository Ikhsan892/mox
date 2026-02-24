package mastercore

import (
	"context"
	"sync"

	core "mox/internal"
	"mox/use_cases/bus"
	"mox/use_cases/operation"
	"mox/use_cases/workerclient"
)

type Master struct {
	app     core.App
	workers *ConnectionRegistry
	control operation.IControl
	server  *bus.IPCServerGateway

	Orchestrator operation.SystemCore
	Mu           sync.RWMutex    // Biar aman pas nambah/hapus worker dari goroutine
	Context      context.Context // Buat koordinasi shutdown
	Cancel       context.CancelFunc
}

func NewMasterCore(
	ctx context.Context,
	app core.App,
) *Master {
	conns := NewConnectionRegistry(ctx, app)
	orchestrator := NewOrchestrator(app, conns)

	return &Master{
		app:          app,
		Context:      ctx,
		workers:      conns,
		Orchestrator: orchestrator,
	}
}

func (m *Master) SetOperations(control operation.IControl) *Master {
	m.control = control

	return m
}

func (m *Master) Stop() {
	m.server.Close()
}

func (m *Master) Run() error {
	m.app.Logger().Info("running all IPC Server")

	server := bus.NewIPCServerGateway(
		m.app,
		"/tmp/http_mgr.sock",
		"localhost",
		1111,
		bus.TCP,
	)

	// nangkep listen
	go func(evt chan bus.Event) {
		m.app.Logger().Info("listening all messages")
		for e := range evt {
			_, err := m.control.Execute(m.Context, m.Orchestrator, e.Payload)
			if err != nil {
				m.app.Logger().Error(err.Error())
				return
			}

		}
	}(server.Event)

	// nangkep new connectin
	go func(evt chan workerclient.WorkerProcess) {
		m.app.Logger().Info("registering new worker")
		for e := range evt {
			if err := e.Start(); err != nil {
				m.app.Logger().Error(err.Error())
				return
			}

			m.workers.Add(e)
		}
	}(server.WorkerEvent)

	// run for this TCP server
	go server.ListenAndServe()
	go m.workers.CheckHealthWorkers()

	m.server = server

	return nil
}

func (m *Master) Connections() *ConnectionRegistry {
	return m.workers
}
