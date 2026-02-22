package messaging

import "mox/use_cases/orders/dto"

type OrderPublisher interface {
	PublishOrder(payload *dto.Order) error
}
