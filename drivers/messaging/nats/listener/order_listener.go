package listener

import (
	"context"
	"log/slog"
	"time"

	nats_infra "mox/infrastructure/messaging/nats"
	core "mox/internal"
	"mox/use_cases/orders"
	"mox/use_cases/orders/dto"
	"github.com/nats-io/nats.go/jetstream"
	"google.golang.org/protobuf/proto"
)

type OrderListener struct {
	infra  nats_infra.NatsInfrastructure
	ctx    context.Context
	app    core.App
	stream string
}

type OrderListenerConfig struct {
	Stream string
}

func NewOrderListener(core core.App, infra nats_infra.NatsInfrastructure, cfg OrderListenerConfig) (*OrderListener, error) {
	op := &OrderListener{}

	ctx := context.Background()
	op.ctx = ctx
	op.app = core
	op.infra = infra
	op.stream = cfg.Stream

	return op, nil
}

func (c *OrderListener) SubscribeEventSourcing() (jetstream.ConsumeContext, error) {
	ctx := context.Background()
	consumer, err := c.infra.CreateOrUpdateConsumer(ctx, c.stream, jetstream.ConsumerConfig{
		Name:          "processor-2",
		Durable:       "processor-2",
		AckWait:       time.Second,
		AckPolicy:     jetstream.AckExplicitPolicy,
		Description:   "Handling Asset stream",
		MaxAckPending: 100,
		MaxDeliver:    5,
		BackOff: []time.Duration{
			5 * time.Second,
			30 * time.Second,
			300 * time.Second,
		},
	})
	if err != nil {
		c.app.Logger().Error("error while creating consumer", slog.Any("msg", err.Error()))
	}

	c.app.Logger().Info("Success creating consumer ")

	return consumer.Consume(func(msg jetstream.Msg) {
		meta, err := msg.Metadata()

		var req dto.Order
		if err := proto.Unmarshal(msg.Data(), &req); err != nil {
			c.app.Logger().Error("Error marshalling", slog.Any("msg", err.Error()))
		}

		if err != nil {
			c.app.Logger().Error("error getting metadata", slog.Any("msg", err.Error()))
		}

		srv := orders.NewOrderReportListenerImpl(c.app)

		c.app.Logger().Info("Sequence", slog.Uint64("sequence", meta.Sequence.Consumer))

		if err := srv.Receive(&req); err != nil {
			c.app.Logger().Error("Error processing message", slog.String("msg", err.Error()))
		}

		time.Sleep(2 * time.Second)
		msg.Ack()
	})
}

func (c *OrderListener) SubscribeEphemeral() (jetstream.ConsumeContext, error) {
	ctx := context.Background()

	consumer, err := c.infra.CreateOrUpdateConsumer(ctx, c.stream, jetstream.ConsumerConfig{
		Name:            "reporting-3",
		Durable:         "reporting-3",
		MaxRequestBatch: 1000,
		AckPolicy:       jetstream.AckNonePolicy,
	})
	if err != nil {
		c.app.Logger().Error(err.Error())
	}

	return consumer.Consume(func(msg jetstream.Msg) {
		meta, err := msg.Metadata()
		if err != nil {
			c.app.Logger().Error("error getting metadata", slog.Any("msg", err.Error()))
		}

		var req dto.Order
		if err := proto.Unmarshal(msg.Data(), &req); err != nil {
			c.app.Logger().Error("Error marshalling", slog.Any("msg", err.Error()))
		}

		if err != nil {
			c.app.Logger().Error("error getting metadata", slog.Any("msg", err.Error()))
		}

		srv := orders.NewOrderReportListenerImpl(c.app)

		c.app.Logger().Info("Sequence", slog.Uint64("sequence", meta.Sequence.Consumer))

		if err := srv.Receive(&req); err != nil {
			c.app.Logger().Error("Error processing message", slog.String("msg", err.Error()))
		}
	})
}
