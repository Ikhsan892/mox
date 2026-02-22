package workercore

type WorkerState int

const (
	Disconnected WorkerState = iota
	Connected
	Connecting
	Starting
	Error
	Retrying
	Idle
)
