package mastercore

import (
	"bufio"
	"context"
	"fmt"
	"log/slog"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	core "mox/internal"
	"mox/use_cases/bus"
	"mox/use_cases/operation"
	"mox/use_cases/workerclient"
	"mox/use_cases/workercore"
)

type Master struct {
	app core.App

	// 1. Identitas & Alamat
	SocketPath   string
	FD           int
	workers      *ConnectionRegistry
	ListenerFile *os.File
	oob          []byte
	control      operation.IControl
	server       *bus.IPCServerGateway
	Orchestrator operation.SystemCore

	// 4. State Management
	Mu      sync.RWMutex    // Biar aman pas nambah/hapus worker dari goroutine
	Context context.Context // Buat koordinasi shutdown
	Cancel  context.CancelFunc
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

func (m *Master) getSCMRights() {
	rights := syscall.UnixRights(m.FD)
	m.oob = rights
}

func (m *Master) Connections() *ConnectionRegistry {
	return m.workers
}

func (m *Master) writeProceedConnection(conn *net.UnixConn) error {
	payload := []byte("PROCEED")

	// 1. BIKIN FRESH OOB
	// Ambil FD langsung dari file yang dijaga Master
	fdInt := int(m.ListenerFile.Fd())

	rights := syscall.UnixRights(fdInt)

	// 2. KIRIM
	n, oobn, err := conn.WriteMsgUnix(payload, rights, nil)
	if err != nil {
		return fmt.Errorf("gagal kirim msg unix: %v", err)
	}

	// 3. MANTRA ANTI-ZOMBIE 💀
	// Pastikan ListenerFile gak dimatiin GC pas lagi proses kirim
	// runtime.KeepAlive(m.ListenerFile)

	m.app.Logger().Info(fmt.Sprintf("[IPC-SEND] Success! Payload: %d bytes | OOB (FD): %d bytes | Target: %s", n, oobn, conn.RemoteAddr()))

	return nil
}

func (m *Master) spawnNewWorker(worker *workercore.Worker) {
	if err := worker.Start(m.Context); err != nil {
		m.app.Logger().Error("Gagal start worker", err)
		return // Jangan masukin registry
	}

	// m.workers.Add(worker)

	m.app.Logger().Info("Worker registered", slog.Int("pid", worker.PID()))
}

func (m *Master) handleHandshake(c *net.UnixConn) {
	c.SetReadDeadline(time.Now().Add(1 * time.Second))

	reader := bufio.NewReader(c)
	payload, err := reader.ReadString('\n')
	if err != nil {
		m.app.Logger().Error("Handshake failed: cannot read PID", err)
		c.Close()
		return
	}

	// 3. Bersihin string & Parse
	payload = strings.TrimSpace(payload)
	pid, err := strconv.Atoi(payload)
	if err != nil {
		m.app.Logger().Error("Handshake failed: invalid PID format", err)
		c.Close()
		return
	}

	c.SetReadDeadline(time.Time{})

	worker := workercore.NewWorkerBuilder().
		SetListener(c). // Koneksi diserahkan ke sini
		SetPID(pid).
		Build()

	go m.spawnNewWorker(worker)
}

func (m *Master) processWorker(listener *net.UnixListener) {
	duration := time.Duration(5 * time.Minute)

	for {
		if err := listener.SetDeadline(time.Now().Add(duration)); err != nil {
			m.app.Logger().Error(err.Error())
			continue
		}

		conn, err := listener.AcceptUnix()
		if err != nil {
			m.app.Logger().Warn("there is no connection replied, searching...", slog.String("err", err.Error()))
			continue // Jika satu gagal, jangan stop Master-nya, lanjut nunggu yang lain
		}

		m.writeProceedConnection(conn)

		go m.handleHandshake(conn)

		m.app.Logger().Info("berhasil mengirim kunci FD ke worker")

		return
	}
}

func (m *Master) ListenWorker(listener *net.UnixListener) {
	// m.getSCMRights()

	go m.workers.CheckHealthWorkers()

	for {
		select {
		case <-m.Context.Done():
			m.app.Logger().Info("IPC Server shutdown")
			return
		default:
			m.processWorker(listener)
		}
	}
}
