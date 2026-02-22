package nats_messaging

import (
	nats_infra "mox/infrastructure/messaging/nats"
	core "mox/internal"
	"mox/pkg/driver"
)

var _ (driver.IDriver) = (*NatsMesssaging)(nil)

const NATS_DRIVER = "NATS_DRIVER"

func NewNatsMessaging(core core.App, runListener bool) *NatsMesssaging {
	return &NatsMesssaging{
		app:         core,
		runListener: runListener,
	}
}

type NatsMesssaging struct {
	app         core.App
	infra       nats_infra.NatsInfrastructure
	runListener bool
}

// Instance implements driver.IDriver.
func (n *NatsMesssaging) Instance() interface{} {
	return n.infra
}

// Close implements driver.IDriver.
func (n *NatsMesssaging) Close() error {
	// if you want the consumer still can listen while server restarted you can
	// uncomment this

	// ctx := context.Background()
	// if err := n.infra.DeleteConsumer(ctx, "orders", "reporting-3"); err != nil {
	// 	n.app.Logger().Error("Error while deleting consumer", slog.String("msg", err.Error()))
	// }

	// n.app.Logger().Info("Cleaning up consumer reporting-3")

	return n.infra.Close()
}

// Init implements driver.IDriver.
func (n *NatsMesssaging) Init() error {
	infra := nats_infra.NewNatsInfrastructureImpl(n.app)

	if err := infra.Init(); err != nil {
		return err
	}

	n.infra = infra

	if n.runListener {
		registerListener(n)
	}

	return nil
}

// Name implements driver.IDriver.
func (n *NatsMesssaging) Name() string {
	return NATS_DRIVER
}
