package workerclient

type WorkerRegistrar interface {
	Add(worker WorkerProcess) error
	Remove(pid int) error
}
