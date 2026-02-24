package bus

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
	"mox/use_cases/operation"
	"mox/use_cases/workerclient"
)

type IPCServerGateway struct {
	SocketPath  string
	Type        NetworkType
	Port        int
	Event       chan Event
	WorkerEvent chan workerclient.WorkerProcess

	app          core.App
	l            net.Listener
	fdFile       *os.File // file descriptor
	unixListener *net.UnixListener
	mu           *sync.RWMutex
}

func NewIPCServerGateway(
	app core.App,
	SocketPath string,
	Host string,
	Port int,
	Type NetworkType,
) *IPCServerGateway {
	return &IPCServerGateway{
		app:         app,
		SocketPath:  SocketPath,
		Port:        Port,
		Type:        Type,
		mu:          &sync.RWMutex{},
		WorkerEvent: make(chan workerclient.WorkerProcess, 1),
		Event:       make(chan Event, 1),
	}
}

func (c *IPCServerGateway) ListenAndServe() {
	address := fmt.Sprintf(":%d", c.Port)
	os.Remove(c.SocketPath)
	l, err := net.Listen(c.Type.String(), address)
	if err != nil {
		c.app.Logger().Error("cannot run the listener", slog.String("err", err.Error()))
	}

	unixListener, err := net.ListenUnix("unix", &net.UnixAddr{
		Name: c.SocketPath,
		Net:  "unix",
	})
	if err != nil {
		c.app.Logger().Error("cannot run the unix listener", slog.String("err", err.Error()))
		c.app.Stop()
		return
	}

	c.mu.Lock()
	c.l = l
	c.fdFile = c.getFD(l)
	c.unixListener = unixListener
	c.mu.Unlock()

	c.app.Logger().Info(fmt.Sprintf("IPC Server gateway listening on %s", address))
	c.app.Logger().Info(fmt.Sprintf("IPC Server gateway unix listening on socket path %s", c.SocketPath))

	go c.handleWorker(c.app.Context())
}

func (c *IPCServerGateway) Close() {
	if c.l == nil {
		return
	}

	if err := c.fdFile.Close(); err != nil {
		c.app.Logger().Error(err.Error())
	}

	if err := c.l.Close(); err != nil {
		c.app.Logger().Error(err.Error())
	}
}

func (c *IPCServerGateway) handleHandshake(conn *net.UnixConn) {
	conn.SetReadDeadline(time.Now().Add(1 * time.Second))

	reader := bufio.NewReader(conn)
	payload, err := reader.ReadString('\n')
	if err != nil {
		c.app.Logger().Error("Handshake failed: cannot read PID", err)
		conn.Close()
		return
	}

	// 3. Bersihin string & Parse
	payload = strings.TrimSpace(payload)
	pid, err := strconv.Atoi(payload)
	if err != nil {
		c.app.Logger().Error("Handshake failed: invalid PID format", err)
		conn.Close()
		return
	}

	worker := workerclient.NewWorkerClient(c.app, conn, pid)

	// regitering
	c.WorkerEvent <- worker

	c.app.Logger().Debug(fmt.Sprintf("got pid %d", pid))
}

func (c *IPCServerGateway) getFD(l net.Listener) *os.File {
	f, err := l.(*net.TCPListener).File()
	if err != nil {
		c.app.Logger().Error("master cannot get their fd file", slog.String("err", err.Error()))
	}

	// -------------------------------------------------------------------------
	// NOTE: [CRITICAL WARNING]
	// Jangan pernah mengganti implementasi di bawah ini dengan `int(f.Fd())`.
	// Memanggil f.Fd() akan memaksa socket keluar dari Go Netpoller dan masuk
	// ke mode BLOCKING system call.
	// Efek fatalnya:
	// 1. TCP Listener utama (l) akan ikut berubah jadi Blocking.
	// 2. l.Accept() akan macet total (hang) dan tidak bisa di-interrupt.
	// 3. l.SetDeadline() tidak akan berfungsi lagi.
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

func (c *IPCServerGateway) handleWorker(ctx context.Context) {
	for {
		select {
		case <-c.app.Context().Done():
			c.app.Logger().Info("IPC Server shutdown")
			return
		default:
			c.processWorker()
		}
	}
}

func (c *IPCServerGateway) processWorker() {
	listener := c.unixListener

	duration := time.Duration(5 * time.Minute)

	for {
		if err := listener.SetDeadline(time.Now().Add(duration)); err != nil {
			c.app.Logger().Error(err.Error())
			continue
		}

		conn, err := listener.AcceptUnix()
		if err != nil {
			c.app.Logger().Warn("there is no connection replied, searching...", slog.String("err", err.Error()))
			continue // Jika satu gagal, jangan stop Master-nya, lanjut nunggu yang lain
		}

		go func(conn *net.UnixConn) {
			if err := c.writeProceedConnection(conn); err != nil {
				c.app.Logger().Error("failed to send FD", slog.String("err", err.Error()))
				conn.Close()
				return
			}

			c.handleHandshake(conn)
		}(conn)

		c.app.Logger().Info("berhasil mengirim kunci FD ke worker")

		return
	}
}

func (m *IPCServerGateway) writeProceedConnection(conn *net.UnixConn) error {
	payload := []byte("PROCEED")

	fdInt := int(m.fdFile.Fd())

	rights := syscall.UnixRights(fdInt)

	// 2. KIRIM
	n, oobn, err := conn.WriteMsgUnix(payload, rights, nil)
	if err != nil {
		return fmt.Errorf("gagal kirim msg unix: %v", err)
	}

	m.app.Logger().Info(fmt.Sprintf("[IPC-SEND] Success! Payload: %d bytes | OOB (FD): %d bytes | Target: %s | FD %d", n, oobn, conn.RemoteAddr(), fdInt))

	return nil
}

func (c *IPCServerGateway) handleController(ctx context.Context) {
	for {
		conn, err := c.l.Accept()
		if err != nil {
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
			}

			// B. Kalau bangun karena Listener ditutup paksa (Close)
			if strings.Contains(err.Error(), "use of closed network connection") {
				return
			}

			c.app.Logger().Error("Accept error", slog.String("err", err.Error()))

			return
		}

		go c.handleConnectionController(conn)
	}
}

func (c *IPCServerGateway) handleConnectionController(conn net.Conn) {
	defer conn.Close()

	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		text := scanner.Bytes()

		// COPY data biar aman dari race condition buffer scanner
		payload := make([]byte, len(text))
		copy(payload, text)

		// Bungkus & Lempar ke Master
		c.Event <- Event{
			SourceID: conn.RemoteAddr().String(),
			Payload: operation.Command{
				Name:    string(payload),
				Type:    operation.Ping,
				Payload: payload,
			},
			Output: conn,
			Closer: conn,
		}
	}

	if err := scanner.Err(); err != nil {
		c.app.Logger().Error(err.Error())
		return
	}
}
