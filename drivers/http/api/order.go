package api

import (
	"context"

	nats_messaging "goodin/drivers/messaging/nats"
	"goodin/drivers/monitoring"
	nats_infra "goodin/infrastructure/messaging/nats"
	core "goodin/internal"
	"goodin/repositories"
	"goodin/use_cases/orders"
	"goodin/use_cases/orders/dto"
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

type OrderController struct {
	ctx  context.Context
	conn *gorm.DB
	app  core.App
}

// Create Order
// @Summary API untuk Create Order
// @Tags    Menu
// @Accept  json
// @Produce json
// @Param   order  body dto.CreateOrderRequest  true  "Order Request"
// @Failure 400   {object} api.ApiError
// @Failure 401   {object} api.ApiError
// @Failure 404   {object} api.ApiError
// @Failure 500   {object} api.ApiError
// @Router  /v1/order [post]
func (d *OrderController) CreateOrder(c echo.Context) error {
	req := new(dto.CreateOrderRequest)

	tx := d.conn.Begin()
	defer tx.Rollback()

	ctx, span := monitoring.NewTraceContext(c.Request().Context()).
		WithTraceName("Order API").
		WithSpanName("CreateOrder").
		WithTraceparent(c.Request().Header.Get("traceparent")).
		Build()

	defer span.End()

	srv, err := orders.NewOrderServiceImpl(
		d.app,
		orders.WithOrderRepository(repositories.NewOrderPostgreRepository(tx)),
		orders.WithOrderPublisher(
			orders.NewOrderPublisherImpl(
				d.app,
				d.app.Driver().Instance(nats_messaging.NATS_DRIVER).(*nats_infra.NatsInfrastructureImpl),
			),
		),
		orders.WithCustomerDetailRequest(
			orders.NewCustomerDetailRequestImpl(
				d.app,
				d.app.Driver().Instance(nats_messaging.NATS_DRIVER).(*nats_infra.NatsInfrastructureImpl),
			),
		),
	)
	if err != nil {
		return NewInternalServerError(err)
	}

	if err := c.Bind(req); err != nil {
		return NewBadRequestError(err.Error(), nil)
	}

	resp, err := srv.CreateOrder(ctx, *req)
	if err != nil {
		return NewInternalServerError(err)
	}

	tx.Commit()

	return c.JSON(200, resp)
}

func bindOrderApi(e *echo.Group, ctx context.Context, conn core.App) {
	orderController := &OrderController{
		ctx:  ctx,
		conn: conn.Data().Get("sql", "gorm").(*gorm.DB),
		app:  conn,
	}

	e.POST("/v1/order", orderController.CreateOrder)
}
