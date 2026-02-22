package workerclient

type WorkerClientState int

const (
	Disconnected WorkerClientState = iota
	Connected
	Connecting
	Starting
	Error
	Retrying
	Idle
)
