package orders

import (
	"context"
	"errors"
	"log/slog"

	core "goodin/internal"
	"goodin/use_cases/orders/dto"
	"goodin/use_cases/orders/exception"
	"goodin/use_cases/orders/port/input/message/request"
	"goodin/use_cases/orders/port/input/service"
	"goodin/use_cases/orders/port/output/messaging"
	"goodin/use_cases/orders/port/output/repository"

	"go.opentelemetry.io/otel"
	oteltrace "go.opentelemetry.io/otel/trace"
)

var _ (service.OrderService) = (*OrderServiceImpl)(nil)

type OrderServiceImpl struct {
	app            core.App
	customerReq    request.CustomerDetailRequest
	orderRepo      repository.OrderRepository
	orderPublisher messaging.OrderPublisher
	tracer         oteltrace.Tracer
}

type OrderServiceConfiguration func(srv *OrderServiceImpl) error

func NewOrderServiceImpl(app core.App, cfgs ...OrderServiceConfiguration) (*OrderServiceImpl, error) {
	srv := &OrderServiceImpl{
		app:    app,
		tracer: otel.Tracer("OrderServiceImpl"),
	}
	for _, cfg := range cfgs {
		if err := cfg(srv); err != nil {
			return nil, err
		}
	}

	return srv, nil
}

func WithOrderPublisher(orderPublisher messaging.OrderPublisher) OrderServiceConfiguration {
	return func(srv *OrderServiceImpl) error {
		if orderPublisher == nil {
			return errors.New("OrderPublisher required")
		}

		srv.orderPublisher = orderPublisher
		return nil
	}
}

func WithTracer(tracer oteltrace.Tracer) OrderServiceConfiguration {
	return func(srv *OrderServiceImpl) error {
		srv.tracer = tracer
		return nil
	}
}

func WithCustomerDetailRequest(request request.CustomerDetailRequest) OrderServiceConfiguration {
	return func(srv *OrderServiceImpl) error {
		if request == nil {
			return errors.New("CustomerDetailRequest required")
		}

		srv.customerReq = request
		return nil
	}
}

func WithOrderRepository(orderRepo repository.OrderRepository) OrderServiceConfiguration {
	return func(srv *OrderServiceImpl) error {
		if orderRepo == nil {
			return errors.New("OrderRepository required")
		}

		srv.orderRepo = orderRepo
		return nil
	}
}

// CreateOrder implements service.OrderService.
func (o *OrderServiceImpl) CreateOrder(ctxParent context.Context, payload dto.CreateOrderRequest) (dto.CreateOrderResponse, error) {
	ctx := ctxParent
	if o.tracer != nil {
		ctx_, span := o.tracer.Start(ctxParent, "CreateOrder")
		ctx = ctx_

		defer span.End()
	}

	order, err := o.orderRepo.SaveOrder(ctx, payload)
	if err != nil {
		o.app.Logger().Error("Error saving to order", slog.Any("msg", err.Error()))
		return dto.CreateOrderResponse{}, exception.ErrOrderFailedToCreate
	}

	o.app.Logger().Info("Success create to order table ", slog.Any("order", order))

	result := dto.CreateOrderResponse{
		Id:           order.Id,
		CreatedAt:    order.CreatedAt,
		UpdatedAt:    order.UpdatedAt,
		CustomerName: order.CustomerName,
		Status:       order.Status,
		TotalAmount:  order.TotalAmount,
		Address:      order.Address,
		Items:        order.Items,
	}

	return result, nil
}
