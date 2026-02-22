package service

import (
	"context"

	"mox/use_cases/orders/dto"
)

type OrderService interface {
	CreateOrder(ctx context.Context, payload dto.CreateOrderRequest) (dto.CreateOrderResponse, error)
}
