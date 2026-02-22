package orders

import (
	"context"
	"errors"
	"mox/mocks/goodin/use_cases/orders/port/input/message/request"
	"mox/mocks/goodin/use_cases/orders/port/output/messaging"
	"mox/mocks/goodin/use_cases/orders/port/output/repository"
	"mox/use_cases/orders/dto"
	"testing"

	core "mox/internal"

	"github.com/stretchr/testify/assert"
)

func TestNewOrderServiceImpl(t *testing.T) {
	t.Parallel()

	orders, err := NewOrderServiceImpl(nil)

	assert.Nil(t, err)
	assert.NotNil(t, orders)
}

func TestWithOrderPublisherConfig(t *testing.T) {
	t.Parallel()

	mock_pub := messaging.NewMockOrderPublisher(t)

	orders, _ := NewOrderServiceImpl(core.NewBaseApp(), WithOrderPublisher(mock_pub))

	assert.NotNil(t, orders.orderPublisher)
}

func TestWithOrderCustomerReqConfig(t *testing.T) {
	t.Parallel()

	mock_req := request.NewMockCustomerDetailRequest(t)

	orders, _ := NewOrderServiceImpl(core.NewBaseApp(), WithCustomerDetailRequest(mock_req))

	assert.NotNil(t, orders.customerReq)
}

func TestWithOrderRepositoryConfig(t *testing.T) {
	t.Parallel()

	mock_repo := repository.NewMockOrderRepository(t)
	orders, err := NewOrderServiceImpl(core.NewBaseApp(), WithOrderRepository(mock_repo))

	assert.NotNil(t, orders.orderRepo)
	assert.Nil(t, err)
}

func TestWithOrderPublisherConfigNil(t *testing.T) {
	t.Parallel()

	_, err := NewOrderServiceImpl(core.NewBaseApp(), WithOrderPublisher(nil))
	assert.Error(t, err)
}

func TestWithOrderCustmoerReqConfigNil(t *testing.T) {
	t.Parallel()

	_, err := NewOrderServiceImpl(core.NewBaseApp(), WithCustomerDetailRequest(nil))
	assert.Error(t, err)
}

func TestWithTwoConfigForOrder(t *testing.T) {
	t.Parallel()

	mock_pub := messaging.MockOrderPublisher(*messaging.NewMockOrderPublisher(t))
	_, err := NewOrderServiceImpl(core.NewBaseApp(), WithOrderPublisher(&mock_pub), WithOrderRepository(nil))
	assert.Error(t, err)
}

func TestValidateOrder(t *testing.T) {
	testValidatorTable := []struct {
		name    string
		payload dto.CreateOrderRequest
		wantErr bool
	}{
		{
			name: "Pass all validation",
			payload: dto.CreateOrderRequest{
				CustomerName: "ikhsan",
				TotalAmount:  0,
				Address:      "",
				Items: []dto.OrderItemRequest{
					{
						ProductId:   "123",
						ProductName: "Rumah",
						Quantity:    1,
						Price:       10,
					},
					{
						ProductId:   "456",
						ProductName: "Tanah",
						Quantity:    1,
						Price:       20,
					},
				},
			},
			wantErr: false,
		},
		{
			name: "Validation to CustomerName empty",
			payload: dto.CreateOrderRequest{
				CustomerName: "",
				TotalAmount:  0,
				Address:      "",
				Items: []dto.OrderItemRequest{
					{
						ProductId:   "123",
						ProductName: "Rumah",
						Quantity:    1,
						Price:       10,
					},
					{
						ProductId:   "456",
						ProductName: "Tanah",
						Quantity:    1,
						Price:       20,
					},
				},
			},
			wantErr: true,
		},
		{
			name: "Validation to Total Amount is less than 0",
			payload: dto.CreateOrderRequest{
				CustomerName: "",
				TotalAmount:  -1,
				Address:      "",
				Items: []dto.OrderItemRequest{
					{
						ProductId:   "123",
						ProductName: "Rumah",
						Quantity:    1,
						Price:       10,
					},
					{
						ProductId:   "456",
						ProductName: "Tanah",
						Quantity:    1,
						Price:       20,
					},
				},
			},
			wantErr: true,
		},
	}

	for _, test := range testValidatorTable {
		t.Run(test.name, func(t *testing.T) {
			err := test.payload.Validate()
			if test.wantErr {
				assert.NotNil(t, err, err.Error())
			} else {
				assert.Nil(t, err)
				assert.NotEmpty(t, test.payload.CustomerName)
			}
		})
	}
}

func TestCreateOrder(t *testing.T) {
	t.Parallel()

	testTableCreateOrder := []struct {
		name        string
		payload     dto.CreateOrderRequest
		return_repo struct {
			dto.CreateOrderRepositoryResult
			err error
		}
		payload_publisher *dto.Order
		return_publisher  error
		wantErr           bool
		useMockPublisher  bool
		useMockRepo       bool
		useMockCustReq    bool
		isValidationErr   bool
	}{
		{
			name: "Pass",
			payload: dto.CreateOrderRequest{
				CustomerName: "123",
				TotalAmount:  0,
				Address:      "",
				Items: []dto.OrderItemRequest{
					{
						ProductId:   "123",
						ProductName: "Rumah",
						Quantity:    1,
						Price:       10,
					},
					{
						ProductId:   "456",
						ProductName: "Tanah",
						Quantity:    1,
						Price:       20,
					},
				},
			},
			return_repo: struct {
				dto.CreateOrderRepositoryResult
				err error
			}{
				CreateOrderRepositoryResult: dto.CreateOrderRepositoryResult{
					Id:           "123",
					CustomerName: "123",
					TotalAmount:  0,
					Address:      "",
					Items: []dto.OrderItemRequest{
						{
							ProductId:   "123",
							ProductName: "Rumah",
							Quantity:    1,
							Price:       10,
						},
						{
							ProductId:   "456",
							ProductName: "Tanah",
							Quantity:    1,
							Price:       20,
						},
					},
					Status:    "PENDING",
					CreatedAt: 0,
					UpdatedAt: 0,
				},
				err: nil,
			},
			payload_publisher: &dto.Order{
				OrderId:      "123",
				CustomerName: "123",
				TotalAmount:  0,
				Status:       "PENDING",
				Items: []*dto.OrderItem{
					{
						ProductId:   "123",
						ProductName: "Rumah",
						Quantity:    1,
						Price:       10,
					},
					{
						ProductId:   "456",
						ProductName: "Tanah",
						Quantity:    1,
						Price:       20,
					},
				},
			},
			return_publisher: nil,
			wantErr:          false,
			useMockPublisher: true,
			useMockRepo:      true,
		},

		{
			name: "Error on SaveOrder Repository",
			payload: dto.CreateOrderRequest{
				CustomerName: "123",
				TotalAmount:  0,
				Address:      "",
				Items: []dto.OrderItemRequest{
					{
						ProductId:   "123",
						ProductName: "Rumah",
						Quantity:    1,
						Price:       10,
					},
					{
						ProductId:   "456",
						ProductName: "Tanah",
						Quantity:    1,
						Price:       20,
					},
				},
			},
			return_repo: struct {
				dto.CreateOrderRepositoryResult
				err error
			}{
				CreateOrderRepositoryResult: dto.CreateOrderRepositoryResult{},
				err:                         errors.New("Error on saving order, maybe databases error"),
			},
			payload_publisher: &dto.Order{},
			return_publisher:  nil,
			wantErr:           true,
			useMockPublisher:  false,
			useMockRepo:       true,
		},
	}

	ctx := context.Background()

	for _, test := range testTableCreateOrder {
		t.Run(test.name, func(t *testing.T) {
			mock_repo := repository.NewMockOrderRepository(t)

			if test.useMockRepo {
				mock_repo.EXPECT().SaveOrder(ctx, test.payload).Return(test.return_repo.CreateOrderRepositoryResult, test.return_repo.err)
			}

			if err := test.payload.Validate(); err != nil {
				assert.Fail(t, err.Error(), "validation error")
			}

			srv, err := NewOrderServiceImpl(
				core.NewTestApp(),
				WithOrderRepository(mock_repo),
				WithTracer(nil),
			)
			if err != nil {
				assert.Fail(t, err.Error())
			}

			resp, err := srv.CreateOrder(ctx, test.payload)
			if test.wantErr {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
				assert.Equal(t, resp.Address, test.payload.Address)
			}
		})
	}
}
