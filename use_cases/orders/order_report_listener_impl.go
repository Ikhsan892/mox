package orders

import (
	"log/slog"

	core "mox/internal"
	"mox/use_cases/orders/dto"
	"mox/use_cases/orders/port/input/message/listener"
)

var _ (listener.OrderReportListener) = (*OrderReportListenerImpl)(nil)

type OrderReportListenerImpl struct {
	app core.App
}

func NewOrderReportListenerImpl(app core.App) *OrderReportListenerImpl {
	return &OrderReportListenerImpl{
		app: app,
	}
}

// Receive implements listener.OrderReportListener.
func (o *OrderReportListenerImpl) Receive(payload *dto.Order) error {

	o.app.Logger().Info("Processing payload in receive ", slog.Any("payload", payload))

	return nil
}
