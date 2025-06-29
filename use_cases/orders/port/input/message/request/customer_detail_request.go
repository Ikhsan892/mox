package request

import "goodin/use_cases/orders/dto"

type CustomerDetailRequest interface {
	GetDetailCustomer(payload *dto.CustomerDetailRequest) dto.CustomerDetailResponse
}
