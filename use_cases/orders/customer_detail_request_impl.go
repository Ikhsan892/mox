package orders

import (
	"log/slog"
	"time"

	nats_infra "mox/infrastructure/messaging/nats"
	core "mox/internal"
	"mox/use_cases/orders/dto"
	"mox/use_cases/orders/port/input/message/request"
	"github.com/nats-io/nats.go"
	"github.com/sony/gobreaker/v2"
	"google.golang.org/protobuf/proto"
)

var _ (request.CustomerDetailRequest) = (*CustomerDetailRequestImpl)(nil)

type CustomerDetailRequestImpl struct {
	infra nats_infra.NatsInfrastructure
	app   core.App
	cb    *gobreaker.CircuitBreaker[*nats.Msg]
}

func NewCustomerDetailRequestImpl(app core.App, infra nats_infra.NatsInfrastructure) *CustomerDetailRequestImpl {
	payload := &CustomerDetailRequestImpl{
		infra: infra,
		app:   app,
	}
	var st gobreaker.Settings
	st.Name = "Customer Detail Request"
	st.ReadyToTrip = func(counts gobreaker.Counts) bool {
		failureRatio := float64(counts.TotalFailures) / float64(counts.Requests)
		return counts.Requests >= 3 && failureRatio >= 0.6
	}
	st.Timeout = 3 * time.Second
	st.MaxRequests = 3
	st.OnStateChange = func(name string, from, to gobreaker.State) {
		if to == gobreaker.StateOpen {
			app.Logger().Info("State Open !")
		}

		if from == gobreaker.StateOpen && to == gobreaker.StateHalfOpen {
			app.Logger().Info("Going from open to Half Open")
		}

		if from == gobreaker.StateHalfOpen && to == gobreaker.StateClosed {
			app.Logger().Info("Going from half open to closed")
		}

		app.Logger().Debug("OnStateChange", slog.String("name", name), slog.String("from", from.String()), slog.String("to", to.String()))
	}

	payload.cb = gobreaker.NewCircuitBreaker[*nats.Msg](st)

	return payload
}

// GetDetailCustomer implements request.CustomerDetailRequest.
func (c *CustomerDetailRequestImpl) GetDetailCustomer(payload *dto.CustomerDetailRequest) dto.CustomerDetailResponse {

	data, _ := proto.Marshal(payload)

	c.app.Logger().Info("get request detail...")

	// This code is implementing circuit breaker for request-reply
	// pattern, so when it failed. it can retry to max request and
	// can be load balancing to across microservices

	for i := 0; i < 50; i++ {
		_, err := c.cb.Execute(func() (*nats.Msg, error) {
			c.app.Logger().Debug("execute inside circuit breaker")
			return c.infra.Request("customer.get.detail", data, 2*time.Second)
		})
		if err != nil {
			c.app.Logger().Error("error cannot get reply", slog.String("msg", err.Error()))
		}
		time.Sleep(time.Second)
	}

	c.app.Logger().Info("get reply detail")

	return dto.CustomerDetailResponse{}
}
