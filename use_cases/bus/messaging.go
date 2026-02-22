package bus

import (
	"context"

	"mox/use_cases/operation"
	"mox/use_cases/workerclient"
)

type Messaging interface {
	Send(ctx context.Context, msg operation.MessagePayload, payload workerclient.WorkerProcess) error            // send to one PID
	Broadcast(ctx context.Context, payload operation.MessagePayload, workers []workerclient.WorkerProcess) error // send to all PID
}
