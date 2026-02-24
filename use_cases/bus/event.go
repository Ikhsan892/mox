package bus

import (
	"context"
	"fmt"

	core "mox/internal"
	"mox/use_cases/operation"
	"mox/use_cases/workerclient"
)

var _ (Messaging) = (*EventBus)(nil)

type EventBus struct {
	app core.App
}

func NewEventBus(app core.App) *EventBus {
	return &EventBus{app: app}
}

// Broadcast implements [Messaging].
func (e *EventBus) Broadcast(ctx context.Context, payload operation.MessagePayload, workers []workerclient.WorkerProcess) error {
	for _, v := range workers {
		e.app.Logger().Info(fmt.Sprintf("broadcast to PID: %d, MsgType : %s", v.PID(), payload.Payload.Type.String()))

		if v.State() == workerclient.Disconnected {
			e.app.Logger().Info(fmt.Sprintf("PID: %d, msg : still disconnected", v.PID()))
			continue
		}

		if err := e.Send(ctx, payload, v); err != nil {
			v.Shutdown()
			e.app.Logger().Error(fmt.Sprintf("error sending heart PID: %d, msg : %s", v.PID(), err.Error()))
			continue
		}
	}
	return nil
}

// Send implements [Messaging]. disini bisa implement retryable process / middleware
func (e *EventBus) Send(ctx context.Context, msg operation.MessagePayload, payload workerclient.WorkerProcess) error {
	byteTotal, err := payload.Send(ctx, msg)
	if err != nil {
		e.app.Logger().Error(err.Error())
		return err
	}

	e.app.Logger().Info(fmt.Sprintf("byte total send to worker %d", byteTotal))

	return nil
}
