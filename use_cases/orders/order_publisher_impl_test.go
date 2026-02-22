package orders

import (
	"context"
	"errors"
	"testing"

	core "mox/internal"
	nats_infra "mox/mocks/goodin/infrastructure/messaging/nats"
	"mox/use_cases/orders/dto"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
)

func TestOrderPublisherImpl_Publish_PayloadNIl(t *testing.T) {
	mock_infra := nats_infra.NewMockNatsInfrastructure(t)

	publisher := NewOrderPublisherImpl(core.NewBaseApp(), mock_infra)

	if err := publisher.PublishOrder(nil); err != nil {
		assert.NotNil(t, err)
	} else {
		assert.Fail(t, "Should be error")
	}

}

func TestOrderPublisherImpl_Publish_Fail(t *testing.T) {
	payload := dto.Order{
		OrderId: "123",
		Items:   nil,
	}

	data, _ := proto.Marshal(&payload)

	mock_infra := nats_infra.NewMockNatsInfrastructure(t)
	mock_infra.EXPECT().Publish(context.Background(), "orders.123.PENDING", data).Return(nil, errors.New("Fail"))

	publisher := NewOrderPublisherImpl(core.NewBaseApp(), mock_infra)

	if err := publisher.PublishOrder(&payload); err != nil {
		assert.NotNil(t, err)
	} else {
		assert.Fail(t, "Should be error")
	}
}

func TestOrderPublisherImpl_Publish_Pass(t *testing.T) {
	payload := dto.Order{
		OrderId: "123",
		Items:   nil,
	}

	data, _ := proto.Marshal(&payload)

	mock_infra := nats_infra.NewMockNatsInfrastructure(t)
	mock_infra.EXPECT().Publish(context.Background(), "orders.123.PENDING", data).Return(nil, nil)

	publisher := NewOrderPublisherImpl(core.NewBaseApp(), mock_infra)

	if err := publisher.PublishOrder(&payload); err != nil {
		assert.Fail(t, err.Error())
	} else {
		assert.Nil(t, err)
	}
}
