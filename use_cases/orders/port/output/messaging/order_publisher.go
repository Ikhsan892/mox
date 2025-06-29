package messaging

import "goodin/use_cases/orders/dto"

type OrderPublisher interface {
	PublishOrder(payload *dto.Order) error
}
