package workercore

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"os"
	"syscall"

	"mox/use_cases/operation"
)

var _ (WorkerProcess) = (*Worker)(nil)

type Worker struct {
	status    WorkerState
	pid       int
	ExtraFile *os.File // File object wrapper
	fd        int      // Raw FD number
	l         *net.UnixConn
}

// Read implements [WorkerProcess].
func (w *Worker) Read() {
	panic("unimplemented")
}

// Write implements [WorkerProcess].
func (w *Worker) Write() {
	panic("unimplemented")
}

func NewWorker() *Worker {
	return &Worker{
		status: Disconnected,
	}
}

func (w *Worker) HandleFD(fd uintptr, portName string) error {
	panic("unimplemented")
}

// PID implements [WorkerProcess].
func (w *Worker) PID() int {
	return w.pid
}

func (w *Worker) FD() int {
	return w.fd
}

func (w *Worker) createFDFiles(fd int) *os.File {
	InspectFD(fd, "Socket FD")
	file := os.NewFile(uintptr(fd), fmt.Sprintf("listener-%d", fd))

	return file
}

func (w *Worker) AcceptHandshake() error {
	// 1. Panggil fungsi private buat "nyolong" FD dari socket
	fd, msgPayload, err := w.receiveFD()
	if err != nil {
		return fmt.Errorf("gagal menerima FD: %w", err)
	}

	fmt.Printf("[WORKER] Handshake Sukses! Pesan Master: %s | FD: %d\n", msgPayload, fd)

	// 2. Simpan FD ke struct Worker
	w.fd = fd
	w.ExtraFile = w.createFDFiles(fd)

	// 3. Kirim laporan balik ke Master (PID)
	report := fmt.Sprintf("%d\n", w.pid)
	_, err = w.l.Write([]byte(report))
	if err != nil {
		return fmt.Errorf("gagal kirim ack ke master: %w", err)
	}

	return nil
}

func (w *Worker) receiveFD() (int, string, error) {
	oob := make([]byte, syscall.CmsgSpace(4))
	dummy := make([]byte, 128)

	// 1. ReadMsgUnix
	n, oobn, _, _, err := w.l.ReadMsgUnix(dummy, oob)
	if err != nil {
		return 0, "", err
	}

	// 2. Validasi Kritis: Ada data OOB gak?
	if oobn == 0 {
		return 0, string(dummy[:n]), fmt.Errorf("master kirim pesan '%s' tapi OOB DATA KOSONG", dummy[:n])
	}

	// 3. Parsing Control Message
	msgs, err := syscall.ParseSocketControlMessage(oob[:oobn])
	if err != nil {
		return 0, "", fmt.Errorf("parse control msg error: %v", err)
	}

	if len(msgs) == 0 {
		return 0, "", fmt.Errorf("control message kosong")
	}

	// 4. Debugging Log (Opsional, biar lu tetep bisa liat isinya)
	for i, m := range msgs {
		fmt.Printf("   ├─ Pesan OOB ke-%d: Level %d, Type %d\n", i, m.Header.Level, m.Header.Type)
	}

	// 5. Ekstrak FD dari UnixRights
	fds, err := syscall.ParseUnixRights(&msgs[0])
	if err != nil {
		return 0, "", fmt.Errorf("parse unix rights error: %v", err)
	}

	if len(fds) == 0 {
		return 0, "", fmt.Errorf("paket OOB diterima tapi ARRAY FD KOSONG")
	}

	fmt.Println(fds, "list fd files")

	fd := fds[0]

	return fd, string(dummy[:n]), nil
}

func (w *Worker) ReceiveMessage(ctx context.Context, cancelFunc context.CancelFunc) {
	scanner := bufio.NewScanner(w.l)
	// w.app.Logger().Info(fmt.Sprintf("worker %d listening messages", w.pid))

	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		if !scanner.Scan() {
			if err := scanner.Err(); err != nil {
				fmt.Println("Master hilang/mati. Worker ikut pamit!", err.Error())
			}
			cancelFunc()
			return
		}

		command := scanner.Bytes()

		fmt.Printf("[WORKER %d] Nerima Instruksi: %s\n", w.pid, string(command))

		var body operation.MessagePayload

		if err := json.Unmarshal(command, &body); err != nil {
			cancelFunc()
			return
		}

		if body.Payload.Type == operation.Shutdown {
			cancelFunc()
			return
		}
	}
}

// Send implements [WorkerProcess].
func (w *Worker) Send(ctx context.Context, payload []byte) error {
	_, err := w.l.Write(payload)
	if err != nil {
		return err
	}

	return nil
}

// SendHeartbeat implements [WorkerProcess].
func (w *Worker) SendPing(ctx context.Context) error {
	return w.Send(ctx, []byte("PING\n"))
}

// Shutdown implements [WorkerProcess].
func (w *Worker) Shutdown() error {
	defer func() {
		w.status = Disconnected
	}()

	if w.l == nil {
		return nil
	}

	if w.status == Disconnected {
		return errors.New("connection already closed")
	}

	_, err := w.l.Write([]byte("DIE"))
	if err != nil {
		return err
	}

	if err := w.ExtraFile.Close(); err != nil {
		return err
	}

	return w.l.Close()
}

// Start implements [WorkerProcess].
func (w *Worker) Start(ctx context.Context) error {
	w.status = Starting

	return nil
}

// Status implements [WorkerProcess].
func (w *Worker) Status() WorkerState {
	return w.status
}
