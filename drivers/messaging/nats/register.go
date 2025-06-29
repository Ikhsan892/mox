package nats_messaging

import (
	"log/slog"

	"goodin/drivers/messaging/nats/listener"
)

func registerListener(n *NatsMesssaging) {
	order, err := listener.NewOrderListener(n.app, n.infra, listener.OrderListenerConfig{
		Stream: "orders",
	})
	if err != nil {
		n.app.Logger().Error("Error initialization order listener", slog.String("msg", err.Error()))
	}

	order.SubscribeEphemeral()
}
