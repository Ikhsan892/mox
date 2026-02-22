package workercore

import "net"

type WorkerBuilder struct {
	w *Worker
}

func NewWorkerBuilder() *WorkerBuilder {
	return &WorkerBuilder{
		w: &Worker{},
	}
}

func (d *WorkerBuilder) SetPID(pid int) *WorkerBuilder {
	d.w.pid = pid

	return d
}

func (d *WorkerBuilder) SetStatus(status WorkerState) *WorkerBuilder {
	d.w.status = status

	return d
}

func (d *WorkerBuilder) SetListener(l *net.UnixConn) *WorkerBuilder {
	d.w.l = l

	return d
}

func (d *WorkerBuilder) Build() *Worker {
	return d.w
}
