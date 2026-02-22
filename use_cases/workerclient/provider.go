package workerclient

type WorkerProvider interface {
	Get(pid int) WorkerProcess
	GetAll() []WorkerProcess
	Total() int64
}
