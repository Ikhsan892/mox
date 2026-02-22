package publisher

import (
	"context"

	nats_infra "mox/infrastructure/messaging/nats"
	core "mox/internal"
	"github.com/nats-io/nats.go/jetstream"
)

type OrderPublisher struct {
	infra  nats_infra.NatsInfrastructure
	ctx    context.Context
	app    core.App
	stream string
}

type OrderPublisherConfig struct {
	Stream string
}

func NewOrderPublisher(core core.App, infra nats_infra.NatsInfrastructure, cfg OrderPublisher) (*OrderPublisher, error) {
	op := &OrderPublisher{}

	ctx := context.Background()
	op.ctx = ctx
	op.app = core
	op.infra = infra

	infra.CreateConsumer(ctx, cfg.stream, jetstream.ConsumerConfig{})

	return op, nil
}

func (c *OrderPublisher) PublishOrder() {

	c.infra.Publish(c.ctx, "orders.20202.failed", []byte("Content"))

	c.app.Logger().Info("Success publishing 1 message for test")
}
