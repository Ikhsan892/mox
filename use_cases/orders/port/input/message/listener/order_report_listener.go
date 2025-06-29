package listener

import "goodin/use_cases/orders/dto"

type OrderReportListener interface {
	Receive(payload *dto.Order) error
}
