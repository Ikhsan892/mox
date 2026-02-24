package mastercore

import (
	"fmt"
	"time"

	core "mox/internal"
	"mox/tools/utils"
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

// Drain implements [operation.SystemCore].
func (o *Orchestrator) Drain(pid int) error {
	worker := o.provider.Get(pid)
	if worker == nil {
		o.app.Logger().Error(fmt.Sprintf("there is no worker process found in pid %d", pid))
		return nil
	}

	command := "set maxconn frontend mygateway 100"

	return o.bus.Send(o.app.Context(), operation.MessagePayload{
		ID:      utils.GenerateUUID(),
		FromPID: pid,
		Payload: operation.Command{
			Name:        "Draining",
			Description: "Draining one PID to shut the gate",
			Usage:       command,
			Type:        operation.Drain,
			Payload:     []byte(command),
		},
		Timestamp: time.Now().UnixMilli(),
	}, worker)
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
