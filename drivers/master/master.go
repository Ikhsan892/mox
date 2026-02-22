package master

import (
	"bufio"
	"context"
	"fmt"
	"log/slog"
	"net"
	"os"
	"strings"
	"sync"

	core "mox/internal"
	"mox/pkg/driver"
	"mox/use_cases/mastercore"
)

var _ (driver.IDriver) = (*MasterAdapter)(nil)

const MasterAdapterName = "MasterAdapter"

type MasterAdapter struct {
	app        core.App
	ctx        context.Context
	mastercore *mastercore.Master
	l          net.Listener // instance listener
	wg         *sync.WaitGroup
}

func NewMasterAdapter(ctx context.Context, app core.App) *MasterAdapter {
	return &MasterAdapter{app: app, ctx: ctx, wg: &sync.WaitGroup{}}
}

// Close implements [driver.IDriver].
func (m *MasterAdapter) Close() error {
	m.mastercore.Connections().CloseAllConnections()

	m.mastercore.Stop()

	m.app.Logger().Info("master closed")

	return nil
}

func (m *MasterAdapter) getFD(l net.Listener) *os.File {
	f, err := l.(*net.TCPListener).File()
	if err != nil {
		m.app.Logger().Error("master cannot get their fd file", slog.String("err", err.Error()))
	}

	// -------------------------------------------------------------------------
	// [CRITICAL WARNING]
	// Jangan pernah mengganti implementasi di bawah ini dengan `int(f.Fd())`.
	//
	// Memanggil f.Fd() akan memaksa socket keluar dari Go Netpoller dan masuk
	// ke mode BLOCKING system call.
	//
	// Efek fatalnya:
	// 1. TCP Listener utama (l) akan ikut berubah jadi Blocking.
	// 2. l.Accept() akan macet total (hang) dan tidak bisa di-interrupt.
	// 3. l.SetDeadline() tidak akan berfungsi lagi.
	//
	// Gunakan SyscallConn().Control() untuk mengintip nilai FD dengan aman
	// tanpa mengubah mode socket.
	// referensi: https://morsmachine.dk/netpoller.html
	// -------------------------------------------------------------------------
	// conn, _ := f.SyscallConn()
	// fd := -1
	//
	// conn.Control(func(fdUint uintptr) {
	// 	fd = int(fdUint)
	// })

	return f
}

// Init implements [driver.IDriver].
func (m *MasterAdapter) InitOld() error {
	ctx := m.app.Context()

	socketPath := "/tmp/http_mgr.sock"

	os.Remove(socketPath)

	l, err := net.Listen("tcp", "localhost:1111")
	if err != nil {
		m.app.Logger().Error("master cannot run the listener", slog.String("err", err.Error()))
	}

	m.l = l

	master := mastercore.NewMasterCore(ctx, m.app)
	master.ListenerFile = m.getFD(l)

	m.mastercore = master

	m.app.Logger().Info(fmt.Sprintf("MASTER nyala (PID: %d). untuk Port :1111", os.Getpid()))

	unixLn, _ := net.ListenUnix("unix", &net.UnixAddr{Name: socketPath, Net: "unix"})
	go master.ListenWorker(unixLn)

	go m.serveMainListener(ctx)

	m.app.Logger().Info("Menunggu Worker untuk mengambil kunci...")

	return nil
}

func (m *MasterAdapter) Init() error {
	ctx := m.app.Context()
	operations := RegisterCommand(m.app)

	master := mastercore.NewMasterCore(
		ctx,
		m.app,
	)

	master.SetOperations(operations)

	master.Run()

	m.app.Logger().Debug("master already running")

	m.mastercore = master

	return nil
}

func (m *MasterAdapter) serveMainListener(ctx context.Context) {
	defer m.app.Logger().Info("Main listener goroutine exited.")

	for {
		conn, err := m.l.Accept()
		if err != nil {
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				// Loop lagi ke atas buat ngecek ctx.Done()
			}

			if strings.Contains(err.Error(), "use of closed network connection") {
				return
			}

			m.app.Logger().Error("Accept error", slog.String("err", err.Error()))

			return
		}

		go m.handleInstruction(conn)
	}
}

func (m *MasterAdapter) handleInstruction(conn net.Conn) {
	m.app.Logger().Info("Master nerima koneksi instruksi baru",
		slog.String("remote", conn.RemoteAddr().String()))

	// 2. Pake Scanner buat baca perintah per baris
	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		command := strings.TrimSpace(scanner.Text())
		if command == "" {
			continue
		}

		m.app.Logger().Info("[MASTER] Execute Command", slog.String("cmd", command))

		// 3. Logic Dispatcher (Bisa lu kembangin ke Orchestrator nanti)
		switch {
		case command == "ping":
			conn.Write([]byte("PONG\n"))

		case command == "status":
			total := m.mastercore.Connections().Total()
			conn.Write([]byte(fmt.Sprintf("Active Workers: %d\n", total)))

		case strings.HasPrefix(command, "broadcast "):
			m.mastercore.Connections().Broadcast([]byte("DIE\n"))
			// conn.Write([]byte("Broadcast sent!\n"))

		case command == "exit":
			conn.Write([]byte("Bye!\n"))

			defer m.app.Stop()

			return

		default:
			conn.Write([]byte(fmt.Sprintf("Unknown command: %s\n", command)))
		}

	}

	if err := scanner.Err(); err != nil {
		m.app.Logger().Error("Instruction read error", slog.String("err", err.Error()))
	}
}

// Instance implements [driver.IDriver].
func (m *MasterAdapter) Instance() interface{} {
	return m.mastercore
}

// Name implements [driver.IDriver].
func (m *MasterAdapter) Name() string {
	return MasterAdapterName
}
