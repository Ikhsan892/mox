package operation

import (
	"context"
)

// ini diisi interface Orchestrator core sama master
type SystemCore interface {
	CheckHealth() string
	GetTotalWorkers() int64
	ScaleUp()
	ScaleDown()
}

type IControl interface {
	Execute(ctx context.Context, master SystemCore, cmd Command) (string, error)
}

type handler func(ctx context.Context, systemCore SystemCore, cmd Command) error

type MasterControlHandler handler

type WorkerControlHandler handler
