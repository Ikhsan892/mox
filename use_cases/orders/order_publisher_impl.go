package orders

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	nats_infra "goodin/infrastructure/messaging/nats"
	core "goodin/internal"
	"goodin/use_cases/orders/dto"
	"goodin/use_cases/orders/port/output/messaging"
	"google.golang.org/protobuf/proto"
)

var _ (messaging.OrderPublisher) = (*CreateOrderPublisherImpl)(nil)

type CreateOrderPublisherImpl struct {
	infra nats_infra.NatsInfrastructure
	app   core.App
}

func NewOrderPublisherImpl(app core.App, infra nats_infra.NatsInfrastructure) *CreateOrderPublisherImpl {
	return &CreateOrderPublisherImpl{
		infra: infra,
		app:   app,
	}
}

func createId(orderId, status string) string {
	return fmt.Sprintf("orders.%s.%s", orderId, status)
}

// this is using stream not req-reply
// Publish implements messaging.OrderPublisher.
func (c *CreateOrderPublisherImpl) PublishOrder(payload *dto.Order) error {
	ctx := context.Background()

	if payload == nil {
		c.app.Logger().Error("payload cannot be nil")
		return errors.New("payload cannot be nil")
	}

	data, err := proto.Marshal(payload)
	if err != nil {
		c.app.Logger().Error("Error proto.Marshal", slog.String("msg", err.Error()))
		return err
	}

	c.app.Logger().Info("publishing with ID", slog.String("id", createId(payload.OrderId, "PENDING")))

	_, err = c.infra.Publish(ctx, createId(payload.OrderId, "PENDING"), data)
	if err != nil {
		c.app.Logger().Error("Error publishing to stream", slog.Any("payload", &payload), slog.String("msg", err.Error()))
		return err
	}

	c.app.Logger().Info("Success publish order")

	return nil
}
