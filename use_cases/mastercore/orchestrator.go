package mastercore

import (
	core "mox/internal"
	"mox/use_cases/bus"
	"mox/use_cases/operation"
	"mox/use_cases/workerclient"
)

var _ (operation.SystemCore) = (*Orchestrator)(nil)

type Orchestrator struct {
	app      core.App
	provider workerclient.WorkerProvider
	bus      bus.Messaging
}

// GetTotalWorkers implements [operation.SystemCore].
func (o *Orchestrator) GetTotalWorkers() int64 {
	return o.provider.Total()
}

func NewOrchestrator(app core.App, provider workerclient.WorkerProvider) *Orchestrator {
	return &Orchestrator{app: app, bus: bus.NewEventBus(app), provider: provider}
}

// CheckHealth implements [operation.SystemCore].
func (o *Orchestrator) CheckHealth() string {
	return "HEALTHY"
}

// ScaleDown implements [operation.SystemCore].
func (o *Orchestrator) ScaleDown() {
	panic("unimplemented")
}

// ScaleUp implements [operation.SystemCore].
func (o *Orchestrator) ScaleUp() {
	panic("unimplemented")
}
