package master

import (
	"context"
	"fmt"

	core "mox/internal"
	"mox/use_cases/operation"
)

func RegisterCommand(app core.App) operation.IControl {
	registry := operation.NewMasterRegistry()

	registry.Register("noop", "description", "usage", func(ctx context.Context, master operation.SystemCore, cmd operation.Command) error {
		app.Logger().Info(fmt.Sprintf("%s %v", "noop command", cmd))

		return nil
	})

	return registry
}
