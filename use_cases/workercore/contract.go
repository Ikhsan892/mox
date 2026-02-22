package workercore

import "context"

type WorkerProcess interface {
	// Lifecycle Management
	// Start menjalankan loop utama worker. Menerima context untuk graceful shutdown.
	Start(ctx context.Context) error
	// Shutdown melakukan cleanup (tutup koneksi, hapus file socket, dll)
	Shutdown() error // TODO: ganti ke close

	// IPC & Coordination
	// HandleFD menerima File Descriptor dari Master dan mengubahnya jadi listener aktif
	HandleFD(fd uintptr, portName string) error
	// Heartbeat mengirimkan status dan metrik ke Master secara berkala
	SendPing(ctx context.Context) error

	Send(ctx context.Context, payload []byte) error

	// Identity & Monitoring
	PID() int

	Status() WorkerState

	Write()

	Read()

	// Metrik buat AI di Master
	// GetMetrics() WorkerMetrics
}
